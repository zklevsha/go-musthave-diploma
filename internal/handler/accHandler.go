package handler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/luhn"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

type accHandler struct{}

func (a *accHandler) rootHandler(w http.ResponseWriter, r *http.Request) {
	_, compress := getFlags(r)
	tooManyReq(w, 10, compress)
	resp := structs.Response{Message: "Server is working"}
	sendResponse(w, http.StatusOK, resp, compress)
}

func (a *accHandler) orderHandler(w http.ResponseWriter, r *http.Request) {
	_, compress := getFlags(r)
	tooManyReq(w, 10, compress)
	v := mux.Vars(r)
	order, err := strconv.Atoi(v["order"])
	if err != nil {
		sendResponse(w, http.StatusBadRequest,
			structs.Response{Error: fmt.Sprintf("cant convert %s to int", v["order"])},
			compress)
		return
	}
	if !luhn.Valid(order) {
		sendResponse(w, http.StatusUnprocessableEntity, structs.Response{Error: "invalid order number"},
			compress)
		return
	}

	if rand.Intn(10)%2 == 0 {
		accrual := float64(order % 359)
		sendResponse(w, http.StatusOK,
			structs.Order{Order: v["order"], Status: "PROCESSED", Accrual: &accrual},
			compress)
		return
	}

	switch rand.Intn(2) {
	case 0:
		sendResponse(w, http.StatusOK,
			structs.Order{Order: v["order"], Status: "REGISTERED"}, compress)
	case 1:
		sendResponse(w, http.StatusOK,
			structs.Order{Order: v["order"], Status: "PROCESSING"}, compress)
	case 2:
		sendResponse(w, http.StatusOK,
			structs.Order{Order: v["order"], Status: "INVALID"}, compress)
	}
}

func AccGetHandler(c config.AccrualConfig, ctx context.Context) http.Handler {
	r := mux.NewRouter()
	a := accHandler{}
	r.HandleFunc("/", a.rootHandler)
	//
	r.HandleFunc("/api/orders/{order}", a.orderHandler).
		Methods("GET")

	return r
}
