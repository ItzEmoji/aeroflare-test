package auth

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestDeviceCode(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code": "dc123", "user_code": "123-456", "verification_uri": "https://github.com/login/device", "interval": 5}`))
	}))
	defer ts.Close()

	res, err := requestDeviceCode("test-client", ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.UserCode != "123-456" {
		t.Errorf("expected 123-456, got %s", res.UserCode)
	}
}

func TestPollAccessTokenSuccess(t *testing.T) {
	t.Parallel()
	var requests int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNum := atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		if reqNum == 1 {
			_, _ = w.Write([]byte(`{"error": "authorization_pending"}`))
			return
		}
		if reqNum == 2 {
			_, _ = w.Write([]byte(`{"error": "slow_down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token": "gho_12345"}`))
	}))
	defer ts.Close()

	token, err := pollAccessToken("client_id", "device_code", 1, ts.URL, time.Millisecond)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if token != "gho_12345" {
		t.Fatalf("expected token gho_12345, got: %s", token)
	}
}

func TestPollAccessTokenTransientError(t *testing.T) {
	t.Parallel()
	var requests int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNum := atomic.AddInt32(&requests, 1)
		if reqNum == 1 {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`<html>502 Bad Gateway</html>`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "gho_12345"}`))
	}))
	defer ts.Close()

	token, err := pollAccessToken("client_id", "device_code", 1, ts.URL, time.Millisecond)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if token != "gho_12345" {
		t.Fatalf("expected token gho_12345, got: %s", token)
	}
}

func TestPollAccessTokenError(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error": "expired_token"}`))
	}))
	defer ts.Close()

	_, err := pollAccessToken("client_id", "device_code", 1, ts.URL, time.Millisecond)
	if err == nil || err.Error() != "expired_token" {
		t.Fatalf("expected expired_token error, got: %v", err)
	}
}

func TestPollAccessTokenClientError(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "access_denied"}`))
	}))
	defer ts.Close()

	_, err := pollAccessToken("client_id", "device_code", 1, ts.URL, time.Millisecond)
	if err == nil || err.Error() != "access_denied" {
		t.Fatalf("expected access_denied, got: %v", err)
	}
}
