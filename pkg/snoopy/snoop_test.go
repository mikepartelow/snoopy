package snoopy_test

import (
	"log/slog"
	"mp/snoopy/pkg/snoopy"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnoopLogg(t *testing.T) {
	upstream := newUpstream(t)
	defer upstream.Close()

	snoop := &snoopy.Snoop{
		Upstream: upstream.URL,
		Logfile:  path.Join(t.TempDir(), "log.txt"),
		Logger:   slog.Default(),
	}

	request, _ := http.NewRequest(http.MethodGet, "/spam", nil)
	response := httptest.NewRecorder()

	snoop.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.FileExists(t, snoop.Logfile)

	logBytes, err := os.ReadFile(snoop.Logfile)
	assert.NoError(t, err)
	assert.Equal(t, upstream.URL+"/spam\n", string(logBytes))
	assert.Equal(t, "and eggs", response.Body.String())
}

func TestSnoopRewrite(t *testing.T) {
	upstream := newUpstream(t)
	defer upstream.Close()

	snoop := &snoopy.Snoop{
		Upstream: upstream.URL,
		Logfile:  path.Join(t.TempDir(), "log.txt"),
		Logger:   slog.Default(),
		RespnoseRewrites: []struct {
			Old         string `yaml:"old"`
			New         string `yaml:"new"`
			MustRewrite bool   `yaml:"must-rewrite"`
		}{
			{"and eggs", "musubi", true},
		},
	}

	request, _ := http.NewRequest(http.MethodGet, "/spam", nil)
	response := httptest.NewRecorder()

	snoop.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "musubi", response.Body.String())
}

func TestSnoopHeader(t *testing.T) {
	upstream := newUpstream(t)
	defer upstream.Close()

	snoop := &snoopy.Snoop{
		Upstream: upstream.URL,
		Logfile:  path.Join(t.TempDir(), "log.txt"),
		Logger:   slog.Default(),
		Headers: []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		}{
			{"charlie", "brown"},
			{"user-agent", "spam"},
		},
	}

	request, _ := http.NewRequest(http.MethodGet, "/header-test", nil)
	response := httptest.NewRecorder()

	snoop.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
}

func newUpstream(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/spam" && r.URL.Path != "/header-test" {
			t.Errorf("expected request to '/spam' or '/header-test', got: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected HTTP GET, got: %s", r.Method)
		}
		if r.URL.Path == "/header-test" && r.Header.Get("charlie") != "brown" {
			t.Errorf("expected header 'charlie' = 'brown', got: %v", r.Header)
		}
		if r.URL.Path == "/header-test" && r.Header.Get("user-agent") != "spam" {
			t.Errorf("expected header 'user-agent' = 'spam', got: %v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("and eggs"))
	}))
}
