package handler

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/zklevsha/go-musthave-diploma/internal/jwt"
	"github.com/zklevsha/go-musthave-diploma/internal/serializer"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

func getFlags(r *http.Request) (bool, bool, bool) {
	requestCompressed :=
		strings.Contains(strings.Join(r.Header["Content-Encoding"], ","), "gzip")
	compressResponse :=
		strings.Contains(strings.Join(r.Header["Accept-Encoding"], ","), "gzip")
	responseAsText :=
		!strings.Contains(strings.Join(r.Header["Accept"], ","), "application/json") &&
			!strings.Contains(strings.Join(r.Header["Accept"], ","), "*/*")

	return requestCompressed, compressResponse, responseAsText
}

func getErrStatusCode(err error) int {
	switch {
	case errors.Is(err, structs.ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, structs.ErrUserAuth):
		return http.StatusUnauthorized
	case errors.Is(err, structs.ErrOrderIdAlreadyUsed):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func TokenGetUserId(r *http.Request, key string) (int, error) {
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

func sendOrdersResponse(w http.ResponseWriter, code int,
	orders []structs.Order, compress bool) {
	responseBody, err := serializer.EncodeOrdersResponse(orders, compress)
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

func sendOrderResponse(w http.ResponseWriter, code int,
	order structs.Order, compress bool) {
	responseBody, err := serializer.EncodeOrderResponse(order, compress)
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

func tooManyReq(w http.ResponseWriter, chance int, compress bool, asText bool) {
	if chance < rand.Intn(100) {
		return
	}
	w.Header().Set("Retry-After", "60")
	sendResponse(w, http.StatusTooManyRequests,
		structs.Response{Error: "No more than N requests per minute allowed"},
		compress, asText)
}
