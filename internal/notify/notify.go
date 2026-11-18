// Package notify sends alerts to the channels the user configured: ntfy,
// Telegram, a generic webhook or e-mail (SMTP). Message text is plain text;
// device names come from the network and are never interpreted as markup.
package notify

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

// Channel types.
const (
	TypeNtfy     = "ntfy"
	TypeTelegram = "telegram"
	TypeWebhook  = "webhook"
	TypeEmail    = "email"
)

// Channel is one place alerts are sent to. Only the fields of its Type are used.
// Secret fields (Token, Password) are write-only through the API.
type Channel struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`

	URL      string `json:"url,omitempty"`      // ntfy topic URL, webhook URL
	Token    string `json:"token,omitempty"`    // ntfy access token, Telegram bot token
	ChatID   string `json:"chat_id,omitempty"`  // Telegram
	Host     string `json:"host,omitempty"`     // SMTP host:port
	Username string `json:"username,omitempty"` // SMTP
	Password string `json:"password,omitempty"` // SMTP
	From     string `json:"from,omitempty"`     // e-mail
	To       string `json:"to,omitempty"`       // e-mail
}

// Message is what gets sent.
type Message struct {
	Title string
	Body  string
}

// Sender delivers a message through a channel.
type Sender struct {
	Client  *http.Client
	Timeout time.Duration
	// TelegramBase overrides https://api.telegram.org (tests).
	TelegramBase string
	// SMTPDial overrides how the SMTP connection is made (tests).
	SMTPDial func(ctx context.Context, addr string) (net.Conn, error)
}

func (s *Sender) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: s.timeout()}
}

func (s *Sender) timeout() time.Duration {
	if s.Timeout > 0 {
		return s.Timeout
	}
	return 15 * time.Second
}

// Validate checks a channel's settings without sending anything.
func Validate(c Channel) error {
	if len(c.Name) > 64 {
		return errors.New("name too long")
	}
	switch c.Type {
	case TypeNtfy, TypeWebhook:
		return validURL(c.URL)
	case TypeTelegram:
		if c.Token == "" || c.ChatID == "" {
			return errors.New("Telegram needs a bot token and a chat id")
		}
		if strings.ContainsAny(c.Token, "/? #") {
			return errors.New("invalid bot token")
		}
	case TypeEmail:
		if _, _, err := net.SplitHostPort(c.Host); err != nil {
			return errors.New("the SMTP server must be host:port")
		}
		if _, err := mail.ParseAddress(c.From); err != nil {
			return errors.New("invalid sender address")
		}
		if _, err := mail.ParseAddress(c.To); err != nil {
			return errors.New("invalid recipient address")
		}
		if strings.ContainsAny(c.Username+c.Password, "\r\n") {
			return errors.New("invalid SMTP credentials")
		}
	default:
		return errors.New("unknown channel type")
	}
	return nil
}

func validURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("the address must be an http:// or https:// URL")
	}
	return nil
}

// Send delivers m through c.
func (s *Sender) Send(ctx context.Context, c Channel, m Message) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	m.Title, m.Body = oneLine(m.Title), strings.TrimSpace(m.Body)
	switch c.Type {
	case TypeNtfy:
		return s.post(ctx, c.URL, "text/plain; charset=utf-8", []byte(m.Body),
			map[string]string{"Title": asciiHeader(m.Title), "Authorization": bearer(c.Token), "Tags": "satellite"})
	case TypeWebhook:
		b, _ := json.Marshal(map[string]string{"title": m.Title, "message": m.Body, "source": "atalaia"})
		return s.post(ctx, c.URL, "application/json", b, nil)
	case TypeTelegram:
		base := s.TelegramBase
		if base == "" {
			base = "https://api.telegram.org"
		}
		b, _ := json.Marshal(map[string]any{"chat_id": c.ChatID, "text": m.Title + "\n\n" + m.Body, "disable_web_page_preview": true})
		return s.post(ctx, base+"/bot"+c.Token+"/sendMessage", "application/json", b, nil)
	case TypeEmail:
		return s.email(ctx, c, m)
	}
	return errors.New("unknown channel type")
}

func bearer(tok string) string {
	if tok == "" {
		return ""
	}
	return "Bearer " + tok
}

func (s *Sender) post(ctx context.Context, target, ctype string, body []byte, hdr map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return errors.New("invalid address")
	}
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", "atalaia")
	for k, v := range hdr {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	res, err := s.client().Do(req)
	if err != nil {
		// never echo the URL: it can contain a bot token
		return errors.New("could not reach the server")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("the server answered %d", res.StatusCode)
	}
	return nil
}

func (s *Sender) email(ctx context.Context, c Channel, m Message) error {
	dial := s.SMTPDial
	if dial == nil {
		d := &net.Dialer{Timeout: s.timeout()}
		dial = func(ctx context.Context, addr string) (net.Conn, error) { return d.DialContext(ctx, "tcp", addr) }
	}
	conn, err := dial(ctx, c.Host)
	if err != nil {
		return errors.New("could not reach the mail server")
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	host, _, _ := net.SplitHostPort(c.Host)
	cl, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return errors.New("the mail server did not greet us")
	}
	defer cl.Close()
	if ok, _ := cl.Extension("STARTTLS"); ok {
		if err := cl.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return errors.New("could not start TLS with the mail server")
		}
	} else if c.Password != "" {
		// never send a password in the clear
		return errors.New("the mail server does not offer TLS; refusing to send the password")
	}
	if c.Password != "" {
		if err := cl.Auth(smtp.PlainAuth("", c.Username, c.Password, host)); err != nil {
			return errors.New("the mail server refused the login")
		}
	}
	from, _ := mail.ParseAddress(c.From)
	to, _ := mail.ParseAddress(c.To)
	if err := cl.Mail(from.Address); err != nil {
		return errors.New("the mail server refused the sender")
	}
	if err := cl.Rcpt(to.Address); err != nil {
		return errors.New("the mail server refused the recipient")
	}
	w, err := cl.Data()
	if err != nil {
		return errors.New("the mail server refused the message")
	}
	msg := "From: " + from.String() + "\r\nTo: " + to.String() + "\r\nSubject: " + mimeSubject(m.Title) +
		"\r\nDate: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" +
		strings.ReplaceAll(strings.ReplaceAll(m.Body, "\r\n", "\n"), "\n", "\r\n") + "\r\n"
	if _, err := io.WriteString(w, msg); err != nil {
		return errors.New("could not send the message")
	}
	if err := w.Close(); err != nil {
		return errors.New("the mail server did not accept the message")
	}
	return cl.Quit()
}

// oneLine removes line breaks and control characters (header injection).
func oneLine(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s))
}

func asciiHeader(s string) string {
	return strings.Map(func(r rune) rune {
		if r > 0x7e {
			return '?'
		}
		return r
	}, s)
}

func mimeSubject(s string) string {
	return mimeQ(s)
}

func mimeQ(s string) string {
	for _, r := range s {
		if r > 0x7e {
			return "=?UTF-8?B?" + b64(s) + "?="
		}
	}
	return s
}

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
