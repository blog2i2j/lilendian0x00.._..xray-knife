package subs

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchAll_Base64Encoded(t *testing.T) {
	configs := "vless://uuid@example.com:443?type=tcp#Config1\nvmess://base64data\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(configs))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(encoded))
	}))
	defer server.Close()

	s := Subscription{Url: server.URL}
	links, err := s.FetchAll()
	if err != nil {
		t.Fatalf("FetchAll error: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}

	if !strings.HasPrefix(links[0], "vless://") {
		t.Errorf("expected first link to start with vless://, got %q", links[0])
	}

	// Verify ConfigLinks is set
	if len(s.ConfigLinks) != len(links) {
		t.Errorf("ConfigLinks length mismatch: got %d, want %d", len(s.ConfigLinks), len(links))
	}
}

func TestFetchAll_PlainText(t *testing.T) {
	configs := "trojan://password@host:443?sni=example.com#Trojan1\nvless://uuid@host:443#VLESS1\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(configs))
	}))
	defer server.Close()

	s := Subscription{Url: server.URL}
	links, err := s.FetchAll()
	if err != nil {
		t.Fatalf("FetchAll error: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}
}

func TestFetchAll_FiltersEmptyLines(t *testing.T) {
	configs := "link1\n\n  \nlink2\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(configs))
	}))
	defer server.Close()

	s := Subscription{Url: server.URL}
	links, err := s.FetchAll()
	if err != nil {
		t.Fatalf("FetchAll error: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("expected 2 non-empty links, got %d: %v", len(links), links)
	}
}

func TestFetchAll_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	s := Subscription{Url: server.URL}
	_, err := s.FetchAll()
	if err == nil {
		t.Fatal("expected error for HTTP 404, got nil")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to mention 404, got: %v", err)
	}
}

func TestFetchAll_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	s := Subscription{Url: server.URL}
	_, err := s.FetchAll()
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestFetchAll_InvalidURL(t *testing.T) {
	s := Subscription{Url: "://invalid"}
	_, err := s.FetchAll()
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestFetchAll_CustomUserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.Write([]byte("link1\n"))
	}))
	defer server.Close()

	s := Subscription{Url: server.URL, UserAgent: "CustomAgent/1.0"}
	_, err := s.FetchAll()
	if err != nil {
		t.Fatalf("FetchAll error: %v", err)
	}

	if receivedUA != "CustomAgent/1.0" {
		t.Errorf("expected User-Agent 'CustomAgent/1.0', got %q", receivedUA)
	}
}

func TestRemoveDuplicate(t *testing.T) {
	s := Subscription{
		ConfigLinks: []string{"link1", "link2", "link1", "link3", "link2"},
	}
	s.RemoveDuplicate(false)

	if len(s.ConfigLinks) != 3 {
		t.Fatalf("expected 3 unique links, got %d: %v", len(s.ConfigLinks), s.ConfigLinks)
	}

	expected := []string{"link1", "link2", "link3"}
	for i, link := range s.ConfigLinks {
		if link != expected[i] {
			t.Errorf("link[%d] = %q, want %q", i, link, expected[i])
		}
	}
}

func TestRemoveDuplicate_NoDuplicates(t *testing.T) {
	s := Subscription{
		ConfigLinks: []string{"a", "b", "c"},
	}
	s.RemoveDuplicate(true)

	if len(s.ConfigLinks) != 3 {
		t.Fatalf("expected 3 links, got %d", len(s.ConfigLinks))
	}
}
