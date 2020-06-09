package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/contributions", getAllContributions).Methods("GET")
	router.HandleFunc("/weights", getContributionWeights).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
