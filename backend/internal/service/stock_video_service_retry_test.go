package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGenerateImageLocalHub_Retry(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		// Always fail to test all retries
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"status":"error","message":"service unavailable"}`)
	}))
	defer server.Close()

	sv := &StockVideoService{
		localHubURL: server.URL,
		httpClient:  &http.Client{},
	}

	ctx := context.Background()
	_, err := sv.generateImageLocalHub(ctx, "test prompt", "landscape")

	if err == nil {
		t.Error("Expected error after 5 failed attempts, got nil")
	}

	expectedCalls := int32(5)
	if atomic.LoadInt32(&callCount) != expectedCalls {
		t.Errorf("Expected %d calls to the server, but got %d", expectedCalls, atomic.LoadInt32(&callCount))
	}
}

func TestGenerateImageRemoteHub_Retry(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		// Always fail to test all retries
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"status":"error","message":"internal server error"}`)
	}))
	defer server.Close()

	sv := &StockVideoService{
		remoteHubURL:   server.URL,
		remoteHubToken: "fake-token",
		httpClient:     &http.Client{},
	}

	ctx := context.Background()
	_, err := sv.generateImageRemoteHub(ctx, "test prompt", "landscape")

	if err == nil {
		t.Error("Expected error after 5 failed attempts, got nil")
	}

	expectedCalls := int32(5)
	if atomic.LoadInt32(&callCount) != expectedCalls {
		t.Errorf("Expected %d calls to the server, but got %d", expectedCalls, atomic.LoadInt32(&callCount))
	}
}
