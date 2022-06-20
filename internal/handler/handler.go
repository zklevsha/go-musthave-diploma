package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/interfaces"
	"github.com/zklevsha/go-musthave-diploma/internal/serializer"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

type Handler struct {
	Storage interfaces.Storage
}

func (h *Handler) sendResponse(w http.ResponseWriter, code int,
	resp structs.Response, compress bool, asText bool) {
	responseBody, err := serializer.EncodeServerResponse(resp, compress, asText)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to encode server response: %s", err.Error())))
		return
	}

	if compress {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.WriteHeader(code)
	w.Write(responseBody)
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	compress :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	asText := !strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	resp := structs.Response{Message: "<html><body><h1>Server is wokring</h1></body></html>"}
	h.sendResponse(w, http.StatusOK, resp, compress, asText)
}

func GetHandler(c config.ServerConfig, ctx context.Context, store interfaces.Storage) http.Handler {
	r := mux.NewRouter()
	h := Handler{Storage: store}
	r.HandleFunc("/", h.rootHandler)
	return r
}
