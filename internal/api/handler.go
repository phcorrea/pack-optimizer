package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"

	"gymshark/internal/service"
	"gymshark/internal/webassets"
)

type optimizeRequest struct {
	ItemsOrdered int   `json:"items_ordered"`
	PackSizes    []int `json:"pack_sizes"`
}

type handler struct {
	static http.Handler
}

func NewHandler() http.Handler {
	staticFiles, err := fs.Sub(webassets.FS, "static")
	if err != nil {
		panic(err)
	}

	h := &handler{
		static: http.FileServer(http.FS(staticFiles)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/api/optimize", h.handleOptimize)
	mux.HandleFunc("/", h.handleStatic)
	return mux
}

func (h *handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) handleOptimize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req optimizeRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	plan, err := service.Optimize(req.ItemsOrdered, req.PackSizes)
	if err != nil {
		if errors.Is(err, service.ErrInvalidItemsOrdered) || errors.Is(err, service.ErrInvalidPackSizes) || errors.Is(err, service.ErrOptimizationTooLarge) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to optimize pack breakdown")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func (h *handler) handleStatic(w http.ResponseWriter, r *http.Request) {
	h.static.ServeHTTP(w, r)
}

func decodeJSON(body io.ReadCloser, dst any) error {
	defer body.Close()

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
