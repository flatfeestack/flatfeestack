package main

import (
	"context"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	neo "github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"log"
	"math/big"
)

type NeoSignature struct {
	Raw []byte `json:"raw"`
}

func getNeoClient(endpoint string) (*neo.Client, error) {
	neoClient, err := neo.New(context.Background(), endpoint, neo.Options{})

	if err != nil {
		return nil, err
	}

	err = neoClient.Init()
	if err != nil {
		return nil, err
	}

	return neoClient, nil
}

func payoutNEO(addressValues []string, teas []*big.Int) (string, error) {
	var payoutNeoHash, err = util.Uint160DecodeStringLE(opts.NEO.Contract)
	if err != nil {
		log.Fatalf(err.Error())
		return "", err
	}
	contractOwnerPrivateKey, err := keys.NewPrivateKeyFromWIF(opts.NEO.PrivateKey)
	if err != nil {
		log.Fatalf(err.Error())
		return "", err
	}
	owner := wallet.NewAccountFromPrivateKey(contractOwnerPrivateKey)

	h := CreateBatchPayoutTx(neoClient, payoutNeoHash, owner, addressValues, teas)
	return h, nil
}

func CreateBatchPayoutTx(c *neo.Client, payoutNeoHash util.Uint160, acc *wallet.Account, addressValues []string, teas []*big.Int) string {
	var devP []interface{}
	for _, v := range addressValues {
		add, _ := address.StringToUint160(v)
		devP = append(devP, add)
	}
	var teaP []interface{}
	for _, v := range teas {
		teaP = append(teaP, v)
	}

	w := io.NewBufBinWriter()
	emit.AppCall(w.BinWriter, payoutNeoHash, "batchPayout", callflag.All, devP, teaP)
	script := w.Bytes()
	log.Printf("About to execute job [batchPayout]")
	sender := acc.PrivateKey().GetScriptHash()
	signer := transaction.Signer{
		Account: sender,
		Scopes:  transaction.CalledByEntry,
	}
	tx, err := c.CreateTxFromScript(script, acc, -1, 0, []neo.SignerAccount{{
		Signer: signer,
	}})
	if err != nil {
		log.Fatalf(err.Error())
	}
	net, err := c.GetNetwork()
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = acc.SignTx(net, tx)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hash, err := c.SendRawTransaction(tx)
	if err != nil {
		log.Fatalf("send raw transaction err: %v", err)
	}
	return hash.StringLE()
}

func getNeoSignature(data PayoutRequest2) (NeoSignature, error) {
	privateKey, err := keys.NewPrivateKeyFromWIF(opts.NEO.PrivateKey)
	if err != nil {
		return NeoSignature{}, err
	}

	ownerIdBytes, _ := data.UserId.MarshalBinary()
	teaArray := data.Amount.Bytes()
	for i := 0; i < len(teaArray)/2; i++ {
		opp := len(teaArray) - 1 - i
		teaArray[i], teaArray[opp] = teaArray[opp], teaArray[i]
	}
	message := append(ownerIdBytes, teaArray...)
	signature := privateKey.Sign(message)

	return NeoSignature{
		signature,
	}, nil
}
