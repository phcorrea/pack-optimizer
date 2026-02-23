package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnknownAPIPath_NotFound(t *testing.T) {
	srv := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/pack-sizes", nil)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.Code)
	}

	if !bytes.Contains(res.Body.Bytes(), []byte("404 page not found")) {
		t.Fatalf("expected file server 404 body, got: %q", res.Body.String())
	}
}

func TestOptimizeEndpoint(t *testing.T) {
	srv := NewHandler()

	body := bytes.NewBufferString(`{"items_ordered":251,"pack_sizes":[250,500,1000,2000,5000]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload struct {
		ItemsOrdered  int   `json:"items_ordered"`
		TotalItems    int   `json:"total_items"`
		TotalPacks    int   `json:"total_packs"`
		PackSizesUsed []int `json:"pack_sizes_used"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ItemsOrdered != 251 || payload.TotalItems != 500 || payload.TotalPacks != 1 {
		t.Fatalf("unexpected optimize response: %+v", payload)
	}
	if len(payload.PackSizesUsed) != 1 || payload.PackSizesUsed[0] != 500 {
		t.Fatalf("unexpected pack_sizes_used: %+v", payload.PackSizesUsed)
	}
}

func TestOptimizeEndpoint_MissingPackSizes(t *testing.T) {
	srv := NewHandler()

	body := bytes.NewBufferString(`{"items_ordered":251}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.Code)
	}
}

func TestOptimizeEndpoint_InvalidItemsOrdered(t *testing.T) {
	srv := NewHandler()

	body := bytes.NewBufferString(`{"items_ordered":0,"pack_sizes":[250,500]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.Code)
	}
}

func TestStaticRootServesIndex(t *testing.T) {
	srv := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	if !bytes.Contains(res.Body.Bytes(), []byte("<!doctype html>")) {
		t.Fatalf("expected html body, got: %q", res.Body.String())
	}
}
