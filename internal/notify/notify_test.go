package notify

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNtfyWebhookTelegram(t *testing.T) {
	type got struct {
		path, ctype, auth, title, body string
	}
	var last got
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		last = got{r.URL.Path, r.Header.Get("Content-Type"), r.Header.Get("Authorization"), r.Header.Get("Title"), string(b)}
		if strings.HasSuffix(r.URL.Path, "/fail") {
			http.Error(w, "nope", 500)
		}
	}))
	defer ts.Close()
	s := &Sender{TelegramBase: ts.URL}
	ctx := context.Background()
	m := Message{Title: "Título\r\nInjected: x", Body: "line"}

	if err := s.Send(ctx, Channel{Type: TypeNtfy, URL: ts.URL + "/topic", Token: "tk"}, m); err != nil {
		t.Fatal(err)
	}
	if last.path != "/topic" || last.auth != "Bearer tk" || last.body != "line" || strings.ContainsAny(last.title, "\r\n") || strings.Contains(last.title, "í") {
		t.Fatalf("ntfy: %+v", last)
	}

	if err := s.Send(ctx, Channel{Type: TypeWebhook, URL: ts.URL + "/h"}, m); err != nil {
		t.Fatal(err)
	}
	var j map[string]string
	if err := json.Unmarshal([]byte(last.body), &j); err != nil || j["message"] != "line" || j["source"] != "atalaia" || strings.ContainsAny(j["title"], "\r\n") {
		t.Fatalf("webhook: %+v %v", last, err)
	}

	if err := s.Send(ctx, Channel{Type: TypeTelegram, Token: "1:abc", ChatID: "7"}, m); err != nil {
		t.Fatal(err)
	}
	if last.path != "/bot1:abc/sendMessage" || !strings.Contains(last.body, `"chat_id":"7"`) {
		t.Fatalf("telegram: %+v", last)
	}

	// errors never reveal the address (it may hold a token)
	err := s.Send(ctx, Channel{Type: TypeWebhook, URL: ts.URL + "/fail"}, m)
	if err == nil || strings.Contains(err.Error(), ts.URL) || !strings.Contains(err.Error(), "500") {
		t.Fatalf("error: %v", err)
	}
	dead := httptest.NewServer(nil)
	addr := dead.URL
	dead.Close()
	s.TelegramBase = addr
	if err := s.Send(ctx, Channel{Type: TypeTelegram, Token: "SECRET", ChatID: "1"}, m); err == nil || strings.Contains(err.Error(), "SECRET") {
		t.Fatalf("unreachable: %v", err)
	}
}

func TestValidate(t *testing.T) {
	ok := []Channel{
		{Type: TypeNtfy, URL: "https://ntfy.sh/mytopic"},
		{Type: TypeWebhook, URL: "http://homeassistant.local:8123/api/webhook/x"},
		{Type: TypeTelegram, Token: "1:abc", ChatID: "9"},
		{Type: TypeEmail, Host: "smtp.example.com:587", From: "a@example.com", To: "b@example.com"},
	}
	for _, c := range ok {
		if err := Validate(c); err != nil {
			t.Errorf("%s: %v", c.Type, err)
		}
	}
	bad := []Channel{
		{Type: TypeNtfy, URL: "ntfy.sh/x"},
		{Type: TypeWebhook, URL: "file:///etc/passwd"},
		{Type: TypeTelegram, Token: "", ChatID: "1"},
		{Type: TypeEmail, Host: "smtp.example.com", From: "a@example.com", To: "b@example.com"},
		{Type: TypeEmail, Host: "h:25", From: "a@example.com", To: "x"},
		{Type: "pager"},
	}
	for _, c := range bad {
		if err := Validate(c); err == nil {
			t.Errorf("%+v should be refused", c)
		}
	}
}

// A tiny SMTP server without STARTTLS: enough to see that a password is never sent in the clear
// and that a plain message is delivered with injection-safe headers.
func fakeSMTP(t *testing.T) (addr string, data chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	data = make(chan string, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		r := bufio.NewReader(c)
		w := func(s string) { _, _ = c.Write([]byte(s + "\r\n")) }
		w("220 test")
		var msg strings.Builder
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					data <- msg.String()
					w("250 ok")
					continue
				}
				msg.WriteString(line + "\n")
				continue
			}
			switch {
			case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
				w("250 test")
			case strings.HasPrefix(line, "DATA"):
				inData = true
				w("354 go")
			case strings.HasPrefix(line, "QUIT"):
				w("221 bye")
				return
			default:
				w("250 ok")
			}
		}
	}()
	return ln.Addr().String(), data
}

func TestEmailPlainDeliveryAndPasswordGuard(t *testing.T) {
	addr, data := fakeSMTP(t)
	s := &Sender{}
	c := Channel{Type: TypeEmail, Host: addr, From: "atalaia@example.com", To: "me@example.com"}
	if err := s.Send(context.Background(), c, Message{Title: "Novo\r\nBcc: x@evil", Body: "hello"}); err != nil {
		t.Fatal(err)
	}
	msg := <-data
	if strings.Contains(msg, "\nBcc:") || !strings.Contains(msg, "Subject: Novo  Bcc: x@evil") || !strings.Contains(msg, "hello") {
		t.Fatalf("message:\n%s", msg)
	}

	addr2, _ := fakeSMTP(t)
	c2 := Channel{Type: TypeEmail, Host: addr2, From: "a@example.com", To: "b@example.com", Username: "u", Password: "pw"}
	err := s.Send(context.Background(), c2, Message{Title: "t", Body: "b"})
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("a password must not be sent without TLS: %v", err)
	}
}
