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
			"{\"r\":\"0x7343340d4047870048ef076635ef0b7cd54643e89b2869d24be57d6e5bd5463c\",\"s\":\"0x6c38031deeaa94cc448c3c303703f27c492c47387d468e3e8a3b83da5fab8009\",\"v\":27}\n",
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
			"{\"r\":\"0x2244246b407d1974a590de53c7179af6153802e342828dda9ef81a3447de20ae\",\"s\":\"0x3a700bb665f97296bc71f0b586c2952fbbe5150d9416d9c4a99376fbfbd7391a\",\"v\":27}\n",
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
