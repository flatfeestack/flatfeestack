package main

import (
	"encoding/json"
	"net/http"
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
