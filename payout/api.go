package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func sign(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	params := mux.Vars(r)
	userId := params["userId"]
	totalPayedOut := params["totalPayedOut"]

	if userId == "" {
		writeErr(w, http.StatusBadRequest, "user id is empty")
		return
	}

	if totalPayedOut == "" {
		writeErr(w, http.StatusBadRequest, "totalPayedOut is empty")
		return
	}

	privateKey, err := crypto.HexToECDSA(opts.Ethereum.PrivateKey)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "private key error %v", err)
		return
	}

	hashRaw := crypto.Keccak256([]byte(userId + "#" + totalPayedOut))
	signature, err := crypto.Sign(hashRaw, privateKey)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "private key error %v", err)
		return
	}

	//https://ethereum.stackexchange.com/questions/45580/validating-go-ethereum-key-signature-with-ecrecover
	sig := Signature{
		signature,
		bytes32(hashRaw),
		bytes32(signature[:32]),
		bytes32(signature[32:64]),
		uint8(int(signature[65])) + 27, // Yes add 27, weird Ethereum quirk
	}

	writeJson(w, sig)
}

func timeWarpOffset(w http.ResponseWriter, _ *http.Request, _ *TokenClaims) {
	tw := Timewarp{
		Offset: secondsAdd,
	}
	writeJson(w, tw)
}

func timeWarp(w http.ResponseWriter, r *http.Request, _ *TokenClaims) {
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

//Generic helpers

func bytes32(b []byte) [32]byte {
	var ret [32]byte
	copy(ret[:], b)
	return ret
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
