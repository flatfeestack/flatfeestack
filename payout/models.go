package main

type Payout struct {
	Uid     string `json:"uid"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type PayoutResponse struct {
	TxHash string `json:"tx_hash"`
}
