package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-diploma/internal/archive"
	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/hash"
	"github.com/zklevsha/go-musthave-diploma/internal/interfaces"
	"github.com/zklevsha/go-musthave-diploma/internal/serializer"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

func getFlags(r *http.Request) (bool, bool, bool) {
	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	responseAsText :=
		!strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json")

	return requestCompressed, compressResponse, responseAsText
}

func getErrStatusCode(err error) int {
	switch {
	case errors.Is(err, structs.ErrUserAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

type Handler struct {
	Storage interfaces.Storage
	key     string
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

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed, compressResponse, responseAsText := getFlags(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			h.sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	var creds structs.Credentials
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&creds)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		h.sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	hashedPwd := hash.Sign(h.key, creds.Password)
	err = h.Storage.Register(creds.Login, hashedPwd)
	var respCode int
	resp := structs.Response{}
	if err != nil {
		respCode = getErrStatusCode(err)
		resp.Error = err.Error()
	} else {
		respCode = http.StatusOK
		resp.Message = "user was created"
	}

	h.sendResponse(w, respCode, resp, compressResponse, responseAsText)

}

func GetHandler(c config.ServerConfig, ctx context.Context, store interfaces.Storage) http.Handler {
	r := mux.NewRouter()
	h := Handler{Storage: store, key: c.Key}
	r.HandleFunc("/", h.rootHandler)

	// register
	r.HandleFunc("/api/user/register", h.registerHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return r
}
