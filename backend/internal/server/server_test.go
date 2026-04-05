package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_NilHandlerReturnsError(t *testing.T) {
	_, err := New(Config{Handler: nil})
	if err == nil {
		t.Fatal("expected error for nil Handler, got nil")
	}
}

func TestNew_DefaultAddr(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	s, err := New(Config{Handler: h})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.http.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", s.http.Addr)
	}
}

func TestNew_CustomAddr(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	s, err := New(Config{Handler: h, Addr: ":9999"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.http.Addr != ":9999" {
		t.Errorf("Addr = %q, want :9999", s.http.Addr)
	}
}

func TestNew_WrapsCORSMiddleware(t *testing.T) {
	// The handler should be wrapped with CORS — OPTIONS preflight returns 204.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s, err := New(Config{Handler: inner, Addr: ":0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	s.http.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestRun_ShutdownOnContextCancel(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s, err := New(Config{Handler: handler, Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	// Give the server a moment to start.
	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Run did not return after context cancel")
	}
}

func TestRun_ReturnsErrorOnBadAddr(t *testing.T) {
	// Occupy a port first so our server can't bind to it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listener: %v", err)
	}
	defer ln.Close()
	takenAddr := ln.Addr().String()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	s, err := New(Config{Handler: handler, Addr: takenAddr})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	runErr := make(chan error, 1)
	go func() { runErr <- s.Run(ctx) }()

	select {
	case err := <-runErr:
		if err == nil {
			t.Error("expected Run to return an error when port is already in use")
		}
	case <-time.After(3 * time.Second):
		t.Error("Run did not return error in time")
	}
}
