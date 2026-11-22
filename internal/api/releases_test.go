package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestReleasesNeedLoginAndListNewestFirst(t *testing.T) {
	f := newFixture(t)
	if w := f.do("GET", "/api/releases", nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("without login: %d", w.Code)
	}
	c := f.setup()
	w := f.do("GET", "/api/releases", nil, c)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var got struct {
		Current  string `json:"current"`
		Releases []struct {
			Version string                 `json:"version"`
			EN      struct{ Title string } `json:"en"`
			PT      struct{ Title string } `json:"pt-BR"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Releases) < 2 || got.Releases[0].EN.Title == "" || got.Releases[0].PT.Title == "" {
		t.Fatalf("unexpected body: %s", w.Body)
	}
	if got.Releases[len(got.Releases)-1].Version != "0.1.0" {
		t.Fatalf("oldest release = %s", got.Releases[len(got.Releases)-1].Version)
	}
}
