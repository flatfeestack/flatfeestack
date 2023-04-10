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

func TestPostSignEth(t *testing.T) {
	opts = &Opts{}
	opts.Ethereum.PrivateKey = "df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e"

	t.Run("should generate a signature for ETH", func(t *testing.T) {
		uuid, _ := uuid.Parse("4fed2b83-f968-45cc-8869-a36f844cefdb")
		jsonData, _ := json.Marshal(PayoutRequest2{
			Amount: big.NewInt(12345678),
			UserId: uuid,
		})

		request, _ := http.NewRequest(http.MethodPost, "/admin/sign/eth", bytes.NewBuffer(jsonData))
		response := httptest.NewRecorder()

		signEth(response, request)

		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		assert.Equal(
			t,
			"{\"raw\":\"LF+KC6UIoWd3GIw7va/36pKET9AosfhQ2TF2vtrvYAkOUbFSoYOeuwRc+T0Duwf5hd6A/zw6AmrjYYcqv1CFhAA=\",\"hash\":\"b50f0a01234bfce2876b2957db6e81f068038b5669f5d70215739c8844ddab12\",\"r\":\"2c5f8a0ba508a16777188c3bbdaff7ea92844fd028b1f850d93176bedaef6009\",\"s\":\"0e51b152a1839ebb045cf93d03bb07f985de80ff3c3a026ae361872abf508584\",\"v\":27}\n",
			string(body),
		)
	})
}

func TestPostSignNeo(t *testing.T) {
	opts = &Opts{}
	opts.NEO.PrivateKey = "KxyjQ8eUa4FHt3Gvioyt1Wz29cTUrE4eTqX3yFSk1YFCsPL8uNsY"

	t.Run("should generate a signature for NEO", func(t *testing.T) {
		uuid, _ := uuid.Parse("4fed2b83-f968-45cc-8869-a36f844cefdb")
		jsonData, _ := json.Marshal(PayoutRequest2{
			Amount: big.NewInt(12345678),
			UserId: uuid,
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
