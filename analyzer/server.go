package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	err := setEnvs();
	if err != nil {
		fmt.Println(err)
	}
	getPlatformInformation()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/contributions", getAllContributions).Methods("GET")
	router.HandleFunc("/weights", getContributionWeights).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
