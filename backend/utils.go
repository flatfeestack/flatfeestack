package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"net/http"
	"net/url"
	"strings"
)

func writeJsonStr(w http.ResponseWriter, obj string) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(obj))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
	}
}

func writeJson(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err = json.NewEncoder(w).Encode(obj)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could encode json: %v", err)
	}
}

func IntPow(n int64, m int64) int64 {
	if m == 0 {
		return 1
	}
	result := n
	for i := int64(2); i <= m; i++ {
		result *= n
	}
	return result
}

func isUUIDZero(id *uuid.UUID) bool {
	if id == nil {
		return true
	}
	for x := 0; x < 16; x++ {
		if id[x] != 0 {
			return false
		}
	}
	return true
}

func genRnd(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func isValidUrl(s string) *string {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return nil
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}

	if u.Path == "" {
		return stringPointer(u.Host)
	}
	ts := strings.TrimPrefix(u.Path, "/")
	return stringPointer(ts)
}

func printMap(balanceMap map[string]*big.Int) string {
	s := ""
	for k, c := range supportedCurrencies {
		v := balanceMap[k]
		if v != nil {
			vf := new(big.Float).SetInt(v)
			fi := new(big.Int).Exp(big.NewInt(10), big.NewInt(c.FactorPow), nil)
			ff := new(big.Float).SetInt(fi)
			rs := new(big.Float).Quo(vf, ff)
			s += fmt.Sprintf("%v", rs) + " " + k
		}
	}
	return s
}
