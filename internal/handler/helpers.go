package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/zklevsha/go-musthave-diploma/internal/archive"
	"github.com/zklevsha/go-musthave-diploma/internal/jwt"
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

func encodeResponse(str interface{}, compress bool) ([]byte, error) {
	resp, err := json.Marshal(str)
	if err != nil {
		return nil, fmt.Errorf("failed to encode server response to json %s", err.Error())
	}

	if !compress {
		return resp, nil
	}

	compressed, err := archive.Compress(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to compress server response %s", err.Error())
	}
	return compressed, nil
}

func sendResponse(w http.ResponseWriter, r *http.Request, code int,
	resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_, compress := getFlags(r)
	responseBody, err := encodeResponse(resp, compress)
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

func tooManyReq(w http.ResponseWriter, r *http.Request, chance int) {
	if chance < rand.Intn(100) {
		return
	}
	w.Header().Set("Retry-After", "60")
	sendResponse(w, r, http.StatusTooManyRequests,
		structs.Response{Error: "No more than N requests per minute allowed"})
}
