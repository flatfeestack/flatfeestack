package main

import (
	"bytes"
	"fmt"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

// private key, UUID and amount are the same as we use to test the smart contracts
// the resulting signature (r, s, v) are the same as in the generateSignature function in the test helpers
func TestPostSignEth(t *testing.T) {
	opts = &Opts{}
	opts.Ethereum.PrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	t.Run("should generate a signature for ETH", func(t *testing.T) {
		userId, _ := uuid.Parse("4fed2b83-f968-45cc-8869-a36f844cefdb")
		jsonData, _ := json.Marshal(PayoutRequest2{
			Amount: big.NewInt(10000),
			UserId: userId,
		})

		request, _ := http.NewRequest(http.MethodPost, "/admin/sign/eth", bytes.NewBuffer(jsonData))
		response := httptest.NewRecorder()

		signEth(response, request)

		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		assert.Equal(
			t,
			"{\"r\":\"0x97dc9357711575f0c457b5f0d30754d3fda6e40270cac4de6464fe71b00ff3f7\",\"s\":\"0x7271ce6eb6807bb375751d66a9c4652322ae7b460ceaa36ec1725104c636e463\",\"v\":28}\n",
			string(body),
		)
	})
}

// private key, UUID and amount are the same as we use to test the smart contracts
// the resulting signature (r, s, v) are the same as in the generateSignature function in the test helpers
func TestPostSignUsdc(t *testing.T) {
	opts = &Opts{}
	opts.Ethereum.PrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	t.Run("should generate a signature for USDC", func(t *testing.T) {
		userId, _ := uuid.Parse("4fed2b83-f968-45cc-8869-a36f844cefdb")
		jsonData, _ := json.Marshal(PayoutRequest2{
			Amount: big.NewInt(10),
			UserId: userId,
		})

		request, _ := http.NewRequest(http.MethodPost, "/admin/sign/usdc", bytes.NewBuffer(jsonData))
		response := httptest.NewRecorder()

		signUsdc(response, request)

		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		assert.Equal(
			t,
			"{\"r\":\"0x6aa5ddb34ed10d5e9cf0019c2ea2e6f768c70deb8be91da9a22c575f8af8dbe8\",\"s\":\"0x45708a3ce8cd80b6574b4308ec25e450a29893fd7ecd653792861dc0f4126a2c\",\"v\":28}\n",
			string(body),
		)
	})
}

func TestPostSignNeo(t *testing.T) {
	opts = &Opts{}
	opts.NEO.PrivateKey = "KxyjQ8eUa4FHt3Gvioyt1Wz29cTUrE4eTqX3yFSk1YFCsPL8uNsY"

	t.Run("should generate a signature for NEO", func(t *testing.T) {
		userId, _ := uuid.Parse("4fed2b83-f968-45cc-8869-a36f844cefdb")
		jsonData, _ := json.Marshal(PayoutRequest2{
			Amount: big.NewInt(12345678),
			UserId: userId,
		})

		request, _ := http.NewRequest(http.MethodPost, "/admin/sign/neo", bytes.NewBuffer(jsonData))
		response := httptest.NewRecorder()

		signNeo(response, request)

		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		fmt.Printf(string(body))
		assert.Equal(
			t,
			"{\"raw\":\"YNt4yC7TPasvH2+ywr6jEUt7EnHaVfUUseMglMa8AzmNSrYiiGyRrntbFQCwF7czkrUTG5ME1LNf1muKxz7ujg==\"}\n",
			string(body),
		)
	})
}
