package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gymshark/internal/service"
)

var testDefaultPackSizes = []int{250, 500, 1000, 2000, 5000}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	packSizeService, err := service.GetPackSizeService()
	if err != nil {
		t.Fatalf("GetPackSizeService returned error: %v", err)
	}
	if err := packSizeService.SetPackSizes(testDefaultPackSizes); err != nil {
		t.Fatalf("SetPackSizes returned error: %v", err)
	}

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	return handler
}

func TestUnknownAPIPath_NotFound(t *testing.T) {
	srv := newTestHandler(t)

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
	srv := newTestHandler(t)

	body := bytes.NewBufferString(`{"items_ordered":251}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload struct {
		ItemsOrdered int `json:"items_ordered"`
		TotalItems   int `json:"total_items"`
		TotalPacks   int `json:"total_packs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ItemsOrdered != 251 || payload.TotalItems != 500 || payload.TotalPacks != 1 {
		t.Fatalf("unexpected optimize response: %+v", payload)
	}
}

func TestOptimizeEndpoint_UsesBackendPackSizesOverRequestPayload(t *testing.T) {
	srv := newTestHandler(t)

	packSizeService, err := service.GetPackSizeService()
	if err != nil {
		t.Fatalf("GetPackSizeService returned error: %v", err)
	}
	if err := packSizeService.SetPackSizes([]int{10, 20}); err != nil {
		t.Fatalf("SetPackSizes returned error: %v", err)
	}

	body := bytes.NewBufferString(`{"items_ordered":21}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload struct {
		ItemsOrdered int `json:"items_ordered"`
		TotalItems   int `json:"total_items"`
		TotalPacks   int `json:"total_packs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ItemsOrdered != 21 || payload.TotalItems != 30 || payload.TotalPacks != 2 {
		t.Fatalf("unexpected optimize response: %+v", payload)
	}
}

func TestOptimizeEndpoint_RejectsPackSizesField(t *testing.T) {
	srv := newTestHandler(t)

	body := bytes.NewBufferString(`{"items_ordered":21,"pack_sizes":[1000]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.Code)
	}
}

func TestOptimizeEndpoint_InvalidItemsOrdered(t *testing.T) {
	srv := newTestHandler(t)

	body := bytes.NewBufferString(`{"items_ordered":0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.Code)
	}
}

func TestPackSizesEndpoint_Get(t *testing.T) {
	srv := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/pack-sizes", nil)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload struct {
		PackSizes []int `json:"pack_sizes"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := []int{5000, 2000, 1000, 500, 250}
	if len(payload.PackSizes) != len(want) {
		t.Fatalf("pack_sizes len = %d, want %d", len(payload.PackSizes), len(want))
	}
	for i := range want {
		if payload.PackSizes[i] != want[i] {
			t.Fatalf("pack_sizes[%d] = %d, want %d", i, payload.PackSizes[i], want[i])
		}
	}
}

func TestPackSizesEndpoint_Update(t *testing.T) {
	srv := newTestHandler(t)

	updateBody := bytes.NewBufferString(`{"pack_sizes":[10,20]}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/pack-sizes", updateBody)
	updateRes := httptest.NewRecorder()
	srv.ServeHTTP(updateRes, updateReq)

	if updateRes.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", updateRes.Code)
	}

	var payload struct {
		PackSizes []int `json:"pack_sizes"`
	}
	if err := json.NewDecoder(updateRes.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := []int{20, 10}
	if len(payload.PackSizes) != len(want) {
		t.Fatalf("pack_sizes len = %d, want %d", len(payload.PackSizes), len(want))
	}
	for i := range want {
		if payload.PackSizes[i] != want[i] {
			t.Fatalf("pack_sizes[%d] = %d, want %d", i, payload.PackSizes[i], want[i])
		}
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/pack-sizes", nil)
	getRes := httptest.NewRecorder()
	srv.ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", getRes.Code)
	}
}

func TestPackSizesEndpoint_UpdateInvalid(t *testing.T) {
	srv := newTestHandler(t)

	updateBody := bytes.NewBufferString(`{"pack_sizes":[0,20]}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/pack-sizes", updateBody)
	updateRes := httptest.NewRecorder()
	srv.ServeHTTP(updateRes, updateReq)

	if updateRes.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", updateRes.Code)
	}
}

func TestStaticRootServesIndex(t *testing.T) {
	srv := newTestHandler(t)

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
