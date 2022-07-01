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
	_, compress, asText := getFlags(r)

	_, err := TokenGetUserId(r, h.key)
	if err != nil {
		e := fmt.Sprintf("Authentication failure: %s", err.Error())
		sendResponse(w, http.StatusUnauthorized, structs.Response{Error: e},
			compress, asText)
		return
	}

	resp := structs.Response{Message: "Server is working"}
	sendResponse(w, http.StatusOK, resp, compress, asText)
}

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed, compressResponse, responseAsText := getFlags(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	var creds structs.Credentials
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&creds)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
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
	sendResponse(w, respCode, resp, compressResponse, responseAsText)

}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed, compressResponse, responseAsText := getFlags(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	var creds structs.Credentials
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&creds)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	id, err := h.Storage.GetUserID(creds)
	if err != nil {
		e := fmt.Sprintf("failed to authenticate user: %s", err.Error())
		sendResponse(w, getErrStatusCode(err), structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	token, err := jwt.Generate(id, h.key)
	if err != nil {
		e := fmt.Sprintf("failed to generate jwt token: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	sendResponse(w, http.StatusOK, structs.Response{Token: token},
		compressResponse, responseAsText)
}

func (h *Handler) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed, compressResponse, responseAsText := getFlags(r)

	userid, err := TokenGetUserId(r, h.key)
	if err != nil {
		e := fmt.Sprintf("Authentication failure: %s", err.Error())
		sendResponse(w, http.StatusUnauthorized, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	orderid, err := strconv.Atoi(string(b))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: "invalid request format"},
			compressResponse, responseAsText)
		return
	}

	if !luhn.Valid(orderid) {
		sendResponse(w, http.StatusUnprocessableEntity, structs.Response{Error: "invalid orderid value"},
			compressResponse, responseAsText)
		return
	}

	dbChanged, err := h.Storage.CreateOrder(userid, orderid)
	if err != nil {
		sendResponse(w, getErrStatusCode(err), structs.Response{Error: err.Error()},
			compressResponse, responseAsText)
		return
	}
	if dbChanged {
		sendResponse(w, http.StatusAccepted, structs.Response{Message: "order created"},
			compressResponse, responseAsText)

	} else {
		sendResponse(w, http.StatusOK, structs.Response{Message: "order already exists"},
			compressResponse, responseAsText)
	}

}

func (h *Handler) getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	_, compressResponse, responseAsText := getFlags(r)
	userid, err := TokenGetUserId(r, h.key)
	if err != nil {
		e := fmt.Sprintf("authentication failure: %s", err.Error())
		sendResponse(w, http.StatusUnauthorized, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}
	orders, err := h.Storage.GetOrders(userid)
	if err != nil {
		e := fmt.Sprintf("cant get orders: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}
	if len(orders) == 0 {
		sendResponse(w, http.StatusNoContent, structs.Response{Message: "no orders were found"},
			compressResponse, responseAsText)
	}
	sendOrdersResponse(w, http.StatusOK, orders, compressResponse)

}

func (h *Handler) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	_, compressResponse, responseAsText := getFlags(r)
	userid, err := TokenGetUserId(r, h.key)
	if err != nil {
		e := fmt.Sprintf("authentication failure: %s", err.Error())
		sendResponse(w, http.StatusUnauthorized, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	balance, err := h.Storage.GetUserBalance(userid)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, structs.Response{Error: err.Error()},
			compressResponse, responseAsText)
	}
	sendResponse(w, http.StatusOK, balance, compressResponse, false)
}

func (h *Handler) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	requestCompressed, compressResponse, responseAsText := getFlags(r)

	userid, err := TokenGetUserId(r, h.key)
	if err != nil {
		e := fmt.Sprintf("authentication failure: %s", err.Error())
		sendResponse(w, http.StatusUnauthorized, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Sprintf("failed to read body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
	}

	if requestCompressed {
		b, err = archive.Decompress(b)
		if err != nil {
			e := fmt.Sprintf("Failed to decompress request body: %s", err.Error())
			sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
				compressResponse, responseAsText)
			return
		}
	}

	var withdraw structs.Withdraw
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&withdraw)
	if err != nil {
		e := fmt.Sprintf("failed to decode request body: %s", err.Error())
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	orderid, err := strconv.Atoi(string(withdraw.Order))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, structs.Response{Error: "invalid order value"},
			compressResponse, responseAsText)
	}

	if !luhn.Valid(orderid) {
		sendResponse(w, http.StatusUnprocessableEntity, structs.Response{Error: "invalid order value"},
			compressResponse, responseAsText)
		return
	}

	balance, err := h.Storage.GetUserBalance(userid)
	if err != nil {
		e := fmt.Sprintf("failed to get users`s balance: %s", err.Error())
		sendResponse(w, http.StatusInternalServerError,
			structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	if balance.Current < withdraw.Sum {
		e := fmt.Sprintf("winthdraw sum exceeds current balance (%d)", balance.Current)
		sendResponse(w, http.StatusPaymentRequired,
			structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	err = h.Storage.Withdraw(userid, withdraw)
	if err != nil {
		e := fmt.Sprintf("failed to withdraw: %s", err.Error())
		sendResponse(w, http.StatusInternalServerError, structs.Response{Error: e},
			compressResponse, responseAsText)
		return
	}

	sendResponse(w, http.StatusOK,
		structs.Response{Message: "withdraw reqest was proccessed"},
		compressResponse, responseAsText)
}

func GetHandler(c config.ServerConfig, ctx context.Context, store interfaces.Storage) http.Handler {
	r := mux.NewRouter()
	h := Handler{Storage: store, key: c.Key}
	r.HandleFunc("/", h.rootHandler)

	// register
	r.HandleFunc("/api/user/register", h.registerHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// login
	r.HandleFunc("/api/user/login", h.loginHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// create order
	r.HandleFunc("/api/user/orders", h.createOrderHandler).
		Methods("POST").
		Headers("Content-Type", "text/plain")

	// get orders
	r.HandleFunc("/api/user/orders", h.getOrdersHandler).
		Methods("GET")

	// get balance
	r.HandleFunc("/api/user/balance", h.getBalanceHandler).
		Methods("GET")

	// withdraw
	r.HandleFunc("/api/user/balance/withdraw", h.withdrawHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")
	return r
}
