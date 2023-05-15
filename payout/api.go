package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

type PayoutAddresses struct {
	Eth  string `json:"eth"`
	Neo  string `json:"neo"`
	Usdc string `json:"usdc"`
}

type PayoutConfig struct {
	PayoutContractAddresses PayoutAddresses `json:"payoutContractAddresses"`
	ChainId                 int64           `json:"chainId"`
	Env                     string          `json:"env"`
}

type PayoutRequest struct {
	Amount *big.Int  `json:"amount"`
	UserId uuid.UUID `json:"userId"`
}

type PayoutResponse struct {
	Amount        *big.Int `json:"amount"`
	Currency      string   `json:"currency"`
	EncodedUserId string   `json:"encodedUserId"`
	Signature     string   `json:"signature"`
}

type DaoConfigResponse struct {
	Dao        string `json:"dao"`
	Membership string `json:"membership"`
	Wallet     string `json:"wallet"`
	ChainId    int64  `json:"chainId"`
}

func signNeo(w http.ResponseWriter, r *http.Request) {
	var data PayoutRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	signature, err := getNeoSignature(data)
	if err == nil {
		writeJson(w, signature)
	} else {
		writeErr(w, http.StatusBadRequest, "error when generating signature %v", err)
		return
	}
}

func signEth(w http.ResponseWriter, r *http.Request) {
	signForEth(w, r, "ETH")
}

func signUsdc(w http.ResponseWriter, r *http.Request) {
	signForEth(w, r, "USDC")
}

func signForEth(w http.ResponseWriter, r *http.Request, symbol string) {
	var data PayoutRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	signature, err := getEthSignature(data, symbol)
	if err == nil {
		writeJson(w, signature)
	} else {
		writeErr(w, http.StatusBadRequest, "error when generating signature %v", err)
		return
	}
}

func serverTime(w http.ResponseWriter, r *http.Request, email string) {
	currentTime := timeNow()
	writeJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(secondsAdd)+`}`)
}

func serverTimeEth(w http.ResponseWriter, r *http.Request, email string) {
	header, err := ethClient.c.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	currentTime := time.Unix(int64(header.Time), 0)
	writeJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(secondsAdd)+`}`)
}

func timeWarp(w http.ResponseWriter, r *http.Request, _ string) {
	m := mux.Vars(r)
	h := m["hours"]
	if h == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	seconds := hours * 60 * 60
	err = warpChain(seconds, ethClient.rpc)
	//TODO: warp chain on in-memory NEO chain

	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not warp time: %v", m)
		return
	}

	secondsAdd += seconds
	log.Printf("time warp: %v", timeNow())
}

func payoutConfig(w http.ResponseWriter, _ *http.Request) {
	payoutAddresses := PayoutAddresses{
		Eth:  opts.Ethereum.Contract,
		Neo:  opts.NEO.Contract,
		Usdc: opts.Usdc.Contract,
	}

	cfg := PayoutConfig{
		PayoutContractAddresses: payoutAddresses,
		ChainId:                 ethClient.chainId.Int64(),
		Env:                     opts.Env,
	}
	writeJson(w, cfg)
}

func daoConfig(w http.ResponseWriter, _ *http.Request) {
	response := DaoConfigResponse{
		ChainId:    ethClient.chainId.Int64(),
		Dao:        opts.Dao.Dao,
		Membership: opts.Dao.Membership,
		Wallet:     opts.Dao.Wallet,
	}

	writeJson(w, response)
}

//Helpers to respond to the api calls

func writeErr(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Printf(msg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	if debug {
		w.Write([]byte(`{"error":"` + msg + `"}`))
	}
}

func writeJson(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err = json.NewEncoder(w).Encode(obj)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
	}
}

func writeJsonStr(w http.ResponseWriter, obj string) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(obj))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
	}
}
