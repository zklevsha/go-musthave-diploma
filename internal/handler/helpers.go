package handler

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/zklevsha/go-musthave-diploma/internal/interfaces"
	"github.com/zklevsha/go-musthave-diploma/internal/jwt"
	"github.com/zklevsha/go-musthave-diploma/internal/serializer"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

func getFlags(r *http.Request) (bool, bool) {
	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")

	return requestCompressed, compressResponse
}

func getErrStatusCode(err error) int {
	switch {
	case errors.Is(err, structs.ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, structs.ErrUserAuth):
		return http.StatusUnauthorized
	case errors.Is(err, structs.ErrOrderIDAlreadyUsed):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func TokenGetUserID(r *http.Request, key string) (int, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return -1, errors.New("authorization header is not set")
	}
	splitToken := strings.Split(auth, "Bearer")
	if len(splitToken) != 2 {
		return -1, errors.New("bad format Authorization header: expect Bearer <jwt-token>")
	}
	token := strings.TrimSpace(splitToken[1])
	return jwt.GetUserID(token, key)
}

func sendResponse(w http.ResponseWriter, code int,
	resp interfaces.ServerResponse, compress bool) {
	w.Header().Set("Content-Type", "application/json")

	responseBody, err := serializer.EncodeServerResponse(resp, compress, false)
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

func sendResponseJSON(w http.ResponseWriter, code int,
	str interface{}, compress bool) {
	w.Header().Set("Content-Type", "application/json")

	responseBody, err := serializer.EncodeResponse(str, compress)
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

func tooManyReq(w http.ResponseWriter, chance int, compress bool) {
	if chance < rand.Intn(100) {
		return
	}
	w.Header().Set("Retry-After", "60")
	sendResponse(w, http.StatusTooManyRequests,
		structs.Response{Error: "No more than N requests per minute allowed"},
		compress)
}
