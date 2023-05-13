package clients

import (
	"bytes"
	"encoding/base64"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/google/uuid"
	"math/big"
	"net/http"
	"strings"
	"time"
)

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

var (
	payoutUrl      string
	payoutPassword string
	payoutUsername string
)

func InitPayout(payoutUrl0 string, payoutPassword0 string, payoutUsername0 string) {
	payoutUrl = payoutUrl0
	payoutPassword = payoutPassword0
	payoutUsername = payoutUsername0
}

func RequestPayout(userId uuid.UUID, amount *big.Int, currency string) (PayoutResponse, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	preq := PayoutRequest{
		amount,
		userId,
	}

	body, err := json.Marshal(preq)
	if err != nil {
		return PayoutResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, payoutUrl+"/admin/sign/"+strings.ToLower(currency), bytes.NewBuffer(body))
	auth := payoutUsername + ":" + payoutPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return PayoutResponse{}, err
	}
	defer resp.Body.Close()

	var presp PayoutResponse
	err = json.NewDecoder(resp.Body).Decode(&presp)
	if err != nil {
		return PayoutResponse{}, err
	}
	return presp, nil
}
