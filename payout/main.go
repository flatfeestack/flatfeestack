package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	if os.Getenv("ENV") == "local" {
		//if run locally get environment file from above docker config file
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatalf("could not find env file. Please add an .env file if you want to run it without docker. %v", err)
		}
	}

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/pay", PaymentRequestHandler).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PaymentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data []Payout
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not decode Webhook Body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("received payout request for %v addresses", len(data))

	//tx, err := fillContract(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	//_ = json.NewEncoder(w).Encode(PayoutResponse{TxHash: tx.String()})
}
