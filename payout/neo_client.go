package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/manifest"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/nef"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"io/ioutil"
	"log"
	"math/big"
)

func payoutNEO(addressValues []string, teas []*big.Int) string {
	var neo = opts.Blockchains["neo"]
	// Contract hash of deployed contract on testnet

	var payoutNeoHash, _ = util.Uint160DecodeStringLE(neo.Contract) //Own
	contractOwnerPrivateKey, _ := keys.NewPrivateKeyFromWIF(neo.PrivateKey)
	// signatureBytes := signature_provider.NewSignatureNeo(dev, tea, contractOwnerPrivateKey)

	// Following the steps on the developer's side after receiving the signature bytes:
	// Create and initialize client
	// Developer received the signature bytes and can now create the transaction to withdraw funds
	owner := wallet.NewAccountFromPrivateKey(contractOwnerPrivateKey)

	h := CreateBatchPayoutTx(neoClient, payoutNeoHash, owner, 0, owner.PrivateKey().GetScriptHash(), addressValues, teas)
	return h
}

//CreateWithdrawTx creates a transaction to withdraw funds for the provided dev, tea and the signature bytes.
func CreateBatchPayoutTx(c *client.Client, payoutNeoHash util.Uint160, acc *wallet.Account, additionalNetworkFee int64,
	dev util.Uint160, addressValues []string, teas []*big.Int) string {
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
	log.Printf(hex.EncodeToString(script))
	tx, err := c.CreateTxFromScript(script, acc, -1, 0, []client.SignerAccount{{
		Signer: transaction.Signer{
			Account: acc.PrivateKey().GetScriptHash(),
			Scopes:  transaction.CalledByEntry,
		},
	}})
	if err != nil {
		log.Fatalf(err.Error())
	}
	acc.SignTx(c.GetNetwork(), tx)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hash, err := c.SendRawTransaction(tx)
	if err != nil {
		fmt.Errorf("send raw transaction err: %v", err)
	}
	return hash.StringLE()
}

//SignTransaction Signs the transaction with the provided signer account.
func SignTransaction(c *client.Client, signer *wallet.Account, transaction *transaction.Transaction) error {
	return signer.SignTx(c.GetNetwork(), transaction)
}

func readNEFFile(filename string) (*nef.File, []byte, error) {
	if len(filename) == 0 {
		return nil, nil, errors.New("no nef file was provided")
	}

	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	nefFile, err := nef.FileFromBytes(f)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse NEF file: %w", err)
	}

	return &nefFile, f, nil
}

func readManifest(filename string) (*manifest.Manifest, []byte, error) {
	if len(filename) == 0 {
		return nil, nil, errors.New("no manifest file was provided")
	}

	manifestBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	m := new(manifest.Manifest)
	err = json.Unmarshal(manifestBytes, m)
	if err != nil {
		return nil, nil, err
	}
	return m, manifestBytes, nil
}

func deploy(c *client.Client, acc *wallet.Account) (util.Uint160, error) {
	nativeManagementContractHash, err := c.GetNativeContractHash(nativenames.Management)
	if err != nil {
		log.Fatalf("Couldn't get native management contract hash")
	}
	ne, nefB, err := readNEFFile("./PayoutNeo.nef")
	_, mfB, err := readManifest("./PayoutNeo.manifest.json")
	sender := acc.PrivateKey().GetScriptHash()
	pk := acc.PrivateKey().PublicKey().Bytes()
	appCallParams := []smartcontract.Parameter{
		{
			Type:  smartcontract.ByteArrayType,
			Value: nefB,
		},
		{
			Type:  smartcontract.ByteArrayType,
			Value: mfB,
		},
		{
			Type:  smartcontract.PublicKeyType,
			Value: pk,
		},
	}

	contractHash := state.CreateContractHash(sender, ne.Checksum, "PayoutNeo")
	signer := transaction.Signer{
		Account: sender,
		Scopes:  transaction.Global,
		// CustomContracts do not work with neo-go, if that scope is used for the sender when using the method CreateTxFromScript.
		// Same holds for CustomGroups...
		//Scopes:           transaction.CustomContracts,
		//AllowedContracts: []util.Uint160{contractHash},
	}
	resp, _ := c.InvokeFunction(nativeManagementContractHash, "deploy", appCallParams, []transaction.Signer{signer})
	tx, err := c.CreateTxFromScript(resp.Script, acc, -1, 0, []client.SignerAccount{{Signer: signer}})
	if err != nil {
		log.Fatalf(err.Error())
	}
	txHash, err := c.SignAndPushTx(tx, acc, nil)
	if err != nil {
		fmt.Errorf("failed to sign and push transaction: %w", err)
	}
	fmt.Println("---------------------------------")
	fmt.Println("NEO smart contract deployed: " + txHash.StringLE())
	fmt.Println("---------------------------------")
	return contractHash, err
}
