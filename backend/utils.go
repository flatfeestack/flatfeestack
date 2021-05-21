package main

import (
	"encoding/json"
	"net/http"
)

func writeJsonStr(w http.ResponseWriter, obj string) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(obj))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
	}
}

func writeJson(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err = json.NewEncoder(w).Encode(obj)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
	}
}
