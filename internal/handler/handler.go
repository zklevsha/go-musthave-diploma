package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-diploma/internal/archive"
	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/hash"
	"github.com/zklevsha/go-musthave-diploma/internal/interfaces"
	"github.com/zklevsha/go-musthave-diploma/internal/jwt"
	"github.com/zklevsha/go-musthave-diploma/internal/luhn"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

type Handler struct {
	Storage interfaces.Storage
	key     string
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxUserID{} should be set in authentication middleware
	userid := r.Context().Value(structs.RequestCtxUserID{}).(int)
	resp := structs.Response{Message: fmt.Sprintf("Server is working (userid: %d)", userid)}
	sendResponse(w, r, http.StatusOK, resp)
}

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxBody{} should be set in read body middleware
	body := r.Context().Value(structs.RequestCtxBody{}).([]byte)
	var creds structs.Credentials
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&creds)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}

	// Creating user
	hashedPwd := hash.Sign(h.key, creds.Password)
	id, err := h.Storage.Register(creds.Login, hashedPwd)
	if err != nil {
		sendResponse(w, r, getErrStatusCode(err),
			structs.Response{Error: err.Error()})
		return
	}

	// Generating jwt
	token, err := jwt.Generate(id, h.key)
	if err != nil {
		e := fmt.Sprintf("failed to generate jwt token: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	sendResponse(w, r, http.StatusOK, structs.Response{Message: "user was created"})
}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxBody{} should be set in read body middleware
	body := r.Context().Value(structs.RequestCtxBody{}).([]byte)
	var creds structs.Credentials
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&creds)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}

	id, err := h.Storage.GetUserID(creds)
	if err != nil {
		e := fmt.Sprintf("failed to authenticate user: %s", err.Error())
		sendResponse(w, r, getErrStatusCode(err), structs.Response{Error: e})
		return
	}

	token, err := jwt.Generate(id, h.key)
	if err != nil {
		e := fmt.Sprintf("failed to generate jwt token: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	sendResponse(w, r, http.StatusOK, structs.Response{Message: "Authentication successful"})
}

func (h *Handler) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxUserID{} should be set in authentication middleware
	userid := r.Context().Value(structs.RequestCtxUserID{}).(int)
	// RequestCtxBody{} should be set in read body middleware
	body := r.Context().Value(structs.RequestCtxBody{}).([]byte)

	orderid, err := strconv.Atoi(string(body))
	if err != nil {
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: "invalid request format"})
		return
	}

	if !luhn.Valid(orderid) {
		sendResponse(w, r, http.StatusUnprocessableEntity, structs.Response{Error: "invalid orderid value"})
		return
	}

	dbChanged, err := h.Storage.CreateOrder(userid, orderid)
	if err != nil {
		sendResponse(w, r, getErrStatusCode(err), structs.Response{Error: err.Error()})
		return
	}
	if dbChanged {
		sendResponse(w, r, http.StatusAccepted, structs.Response{Message: "order created"})

	} else {
		sendResponse(w, r, http.StatusOK, structs.Response{Message: "order already exists"})
	}

}

func (h *Handler) getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxUserID{} should be set in authentication middleware
	userid := r.Context().Value(structs.RequestCtxUserID{}).(int)
	orders, err := h.Storage.GetOrders(userid)
	if err != nil {
		e := fmt.Sprintf("cant get orders: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}
	if len(orders) == 0 {
		sendResponse(w, r, http.StatusNoContent, structs.Response{Message: "no orders were found"})
	}
	sendResponse(w, r, http.StatusOK, orders)

}

func (h *Handler) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxUserID{} should be set in authentication middleware
	userid := r.Context().Value(structs.RequestCtxUserID{}).(int)
	balance, err := h.Storage.GetUserBalance(userid)
	if err != nil {
		sendResponse(w, r, http.StatusInternalServerError, structs.Response{Error: err.Error()})
	}
	sendResponse(w, r, http.StatusOK, balance)
}

func (h *Handler) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	// RequestCtxUserID{} should be set in authentication middleware
	userid := r.Context().Value(structs.RequestCtxUserID{}).(int)
	// RequestCtxBody{} should be set in read body middleware
	body := r.Context().Value(structs.RequestCtxBody{}).([]byte)

	var withdraw structs.Withdraw
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&withdraw)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}

	orderid, err := strconv.Atoi(string(withdraw.Order))
	if err != nil {
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: "invalid order value"})
	}

	if !luhn.Valid(orderid) {
		sendResponse(w, r, http.StatusUnprocessableEntity, structs.Response{Error: "invalid order value"})
		return
	}

	balance, err := h.Storage.GetUserBalance(userid)
	if err != nil {
		e := fmt.Sprintf("failed to get users`s balance: %s", err.Error())
		sendResponse(w, r, http.StatusInternalServerError, structs.Response{Error: e})
		return
	}

	if balance.Current < withdraw.Sum {
		e := fmt.Sprintf("winthdraw amount exceeds current balance (%f)", balance.Current)
		sendResponse(w, r, http.StatusPaymentRequired, structs.Response{Error: e})
		return
	}

	err = h.Storage.Withdraw(userid, withdraw)
	if err != nil {
		e := fmt.Sprintf("failed to withdraw: %s", err.Error())
		sendResponse(w, r, http.StatusInternalServerError, structs.Response{Error: e})
		return
	}

	sendResponse(w, r, http.StatusOK,
		structs.Response{Message: "withdraw reqest was proccessed"})
}

func (h *Handler) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userid, err := TokenGetUserID(r, h.key)
	if err != nil {
		e := fmt.Sprintf("authentication failure: %s", err.Error())
		sendResponse(w, r, http.StatusUnauthorized, structs.Response{Error: e})
		return
	}
	withdrawals, err := h.Storage.GetWithdrawls(userid)
	if err != nil {
		e := fmt.Sprintf("cant get withdrawals: %s", err.Error())
		sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
		return
	}
	if len(withdrawals) == 0 {
		sendResponse(w, r, http.StatusNoContent,
			structs.Response{Message: "no withdrawals were found"})
	}
	sendResponse(w, r, http.StatusOK, withdrawals)

}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, err := TokenGetUserID(r, h.key)
		if err != nil {
			e := fmt.Sprintf("authentication failure: %s", err.Error())
			sendResponse(w, r, http.StatusUnauthorized, structs.Response{Error: e})
			return
		}
		ctx := context.WithValue(r.Context(), structs.RequestCtxUserID{}, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) readBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCompressed, _ := getFlags(r)
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			e := fmt.Sprintf("failed to read body: %s", err.Error())
			sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
			return
		}
		if requestCompressed {
			b, err = archive.Decompress(b)
			if err != nil {
				e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
				sendResponse(w, r, http.StatusBadRequest, structs.Response{Error: e})
				return
			}
		}
		ctx := context.WithValue(r.Context(), structs.RequestCtxBody{}, b)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetHandler(c config.ServerConfig, ctx context.Context, store interfaces.Storage) http.Handler {
	r := mux.NewRouter()
	h := Handler{Storage: store, key: c.Key}

	// root
	chain := h.authMiddleware(http.HandlerFunc(h.rootHandler))
	r.Handle("/", chain).Methods("GET")

	// register
	chain = h.readBodyMiddleware(http.HandlerFunc(h.registerHandler))
	r.Handle("/api/user/register", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// login
	chain = h.readBodyMiddleware(http.HandlerFunc(h.loginHandler))
	r.Handle("/api/user/login", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// create order
	chain = h.authMiddleware(
		h.readBodyMiddleware(http.HandlerFunc(h.createOrderHandler)))
	r.Handle("/api/user/orders", chain).
		Methods("POST").
		Headers("Content-Type", "text/plain")

	// get orders
	chain = h.authMiddleware(http.HandlerFunc(h.getOrdersHandler))
	r.Handle("/api/user/orders", chain).
		Methods("GET")

	// get balance
	chain = h.authMiddleware(http.HandlerFunc(h.getBalanceHandler))
	r.Handle("/api/user/balance", chain).
		Methods("GET")

	// withdraw
	chain = h.authMiddleware(h.readBodyMiddleware(
		http.HandlerFunc(h.withdrawHandler)))
	r.Handle("/api/user/balance/withdraw", chain).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// get withdrawals
	chain = h.authMiddleware(http.HandlerFunc(h.getWithdrawalsHandler))
	r.Handle("/api/user/withdrawals", chain).
		Methods("GET")
	return r
}
