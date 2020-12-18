package main

import (
 "github.com/ethereum/go-ethereum/common"
 "github.com/ethereum/go-ethereum/ethclient"
 "log"
)

var (
  conn *ethclient.Client
  contract *Flatfeestack
)

func main() {
 // Ganache URL
 conn, err := ethclient.Dial("http://127.0.0.1:7545")
 if err!= nil {
  log.Fatalf("Could not connect to ethereum client %v", err)
 }

 contract, err = NewFlatfeestack(common.HexToAddress("0xa7a921546dB1C0Ed6bF3BA5E953D8C4f4E8dAFC0"), conn)
 if err!= nil {
  log.Fatalf("Failed to instantiate contract: %v", err)
 }
}