package threads

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/malston/saved-threads/internal/api"
)

func TestCreateImageContainer(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "img-container-1"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := CreateImageContainer(context.Background(), client, "me", "https://example.com/photo.jpg", "check this out")
	if err != nil {
		t.Fatalf("CreateImageContainer returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotBody.Get("media_type") != "IMAGE" {
		t.Errorf("media_type = %q, want %q", gotBody.Get("media_type"), "IMAGE")
	}
	if gotBody.Get("image_url") != "https://example.com/photo.jpg" {
		t.Errorf("image_url = %q, want %q", gotBody.Get("image_url"), "https://example.com/photo.jpg")
	}
	if gotBody.Get("text") != "check this out" {
		t.Errorf("text = %q, want %q", gotBody.Get("text"), "check this out")
	}
	if id != "img-container-1" {
		t.Errorf("id = %q, want %q", id, "img-container-1")
	}
}

func TestCreateVideoContainer(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "vid-container-1"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := CreateVideoContainer(context.Background(), client, "me", "https://example.com/clip.mp4", "watch this")
	if err != nil {
		t.Fatalf("CreateVideoContainer returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotBody.Get("media_type") != "VIDEO" {
		t.Errorf("media_type = %q, want %q", gotBody.Get("media_type"), "VIDEO")
	}
	if gotBody.Get("video_url") != "https://example.com/clip.mp4" {
		t.Errorf("video_url = %q, want %q", gotBody.Get("video_url"), "https://example.com/clip.mp4")
	}
	if gotBody.Get("text") != "watch this" {
		t.Errorf("text = %q, want %q", gotBody.Get("text"), "watch this")
	}
	if id != "vid-container-1" {
		t.Errorf("id = %q, want %q", id, "vid-container-1")
	}
}

func TestCreateCarousel(t *testing.T) {
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "carousel-1"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := CreateCarousel(context.Background(), client, "me", []string{"item-1", "item-2", "item-3"}, "my carousel")
	if err != nil {
		t.Fatalf("CreateCarousel returned error: %v", err)
	}

	if gotBody.Get("media_type") != "CAROUSEL" {
		t.Errorf("media_type = %q, want %q", gotBody.Get("media_type"), "CAROUSEL")
	}
	if gotBody.Get("children") != "item-1,item-2,item-3" {
		t.Errorf("children = %q, want %q", gotBody.Get("children"), "item-1,item-2,item-3")
	}
	if gotBody.Get("text") != "my carousel" {
		t.Errorf("text = %q, want %q", gotBody.Get("text"), "my carousel")
	}
	if id != "carousel-1" {
		t.Errorf("id = %q, want %q", id, "carousel-1")
	}
}

func TestCheckStatus(t *testing.T) {
	var gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "FINISHED"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	status, err := CheckStatus(context.Background(), client, "container-99")
	if err != nil {
		t.Fatalf("CheckStatus returned error: %v", err)
	}

	if gotPath != "/container-99" {
		t.Errorf("path = %q, want %q", gotPath, "/container-99")
	}
	if gotQuery.Get("fields") != "status,error_message" {
		t.Errorf("fields = %q, want %q", gotQuery.Get("fields"), "status,error_message")
	}
	if status != "FINISHED" {
		t.Errorf("status = %q, want %q", status, "FINISHED")
	}
}

func TestCheckStatusWithError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":        "ERROR",
			"error_message": "video too long",
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	status, err := CheckStatus(context.Background(), client, "container-bad")
	if err != nil {
		t.Fatalf("CheckStatus returned error: %v", err)
	}
	if status != "ERROR" {
		t.Errorf("status = %q, want %q", status, "ERROR")
	}
}

func TestWaitForReadyFinished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "FINISHED"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := WaitForReady(context.Background(), client, "container-1", 10*time.Millisecond, 1*time.Second)
	if err != nil {
		t.Fatalf("WaitForReady returned error: %v", err)
	}
}

func TestWaitForReadyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":        "ERROR",
			"error_message": "upload failed",
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := WaitForReady(context.Background(), client, "container-bad", 10*time.Millisecond, 1*time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "container error: upload failed" {
		t.Errorf("error = %q, want %q", got, "container error: upload failed")
	}
}

func TestWaitForReadyTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "IN_PROGRESS"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := WaitForReady(context.Background(), client, "container-slow", 5*time.Millisecond, 30*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if got := err.Error(); got != "timed out waiting for container container-slow" {
		t.Errorf("error = %q, want %q", got, "timed out waiting for container container-slow")
	}
}

func TestWaitForReadyPolls(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.WriteHeader(http.StatusOK)
		status := "IN_PROGRESS"
		if n >= 3 {
			status = "FINISHED"
		}
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := WaitForReady(context.Background(), client, "container-poll", 5*time.Millisecond, 1*time.Second)
	if err != nil {
		t.Fatalf("WaitForReady returned error: %v", err)
	}
	if got := calls.Load(); got < 3 {
		t.Errorf("poll count = %d, want >= 3", got)
	}
}
