// This file is a generated binding and any manual changes will be lost.

package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"strconv"
	"strings"
)

var (
	//curl --data '{"method":"eth_getTransactionReceipt","params":["0x25f6a4f512fa76621533e0f93c418be46269f87b8f2e3b751d47d4b6dd1d301f"],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST localhost:8545
	//https://openethereum.github.io/JSONRPC-eth-module#eth_gettransactionreceipt
	ContractAddr = "0x731a10897d267e19b34503ad902d0a29173ba4b1"
)

// Reference imports to suppress errors if they are not otherwise used.
type ClientETH struct {
	c           *ethclient.Client
	rpc         *rpc.Client
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	fromAddress common.Address
	chainId     *big.Int
}

func NewClientETH(ethUrl string, hexPrivateKey string) (*ClientETH, error) {
	rpc, err := rpc.DialContext(context.Background(), ethUrl)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpc)

	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	c := &ClientETH{
		c:           client,
		rpc:         rpc,
		privateKey:  privateKey,
		publicKey:   publicKeyECDSA,
		fromAddress: fromAddress,
	}

	chainId, err := c.c.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	c.chainId = chainId

	return c, nil
}

func (c ClientETH) balanceOf(addresses string) ([]byte, error) {
	to := common.HexToAddress(ContractAddr)
	data := "70a08231" + padString32(addresses)
	callMsg := ethereum.CallMsg{
		From:     c.fromAddress,
		To:       &to,
		Gas:      0,
		GasPrice: nil,
		Value:    nil,
		Data:     common.Hex2Bytes(data),
	}
	ret, err := c.c.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (c ClientETH) unclaimed(addresses []string) (string, error) {
	nonce, gasPrice, err := c.getNonceGas(nil)
	if err != nil {
		return "", err
	}
	stringAddresses := ""
	for _, v := range addresses {
		stringAddresses = stringAddresses + padString32(v)
	}
	gasLimit := uint64(8000000) //8mio gas

	var data = "b2b57708" + padInt64(int64(len(addresses))) + stringAddresses
	tx := types.NewTransaction(nonce, common.HexToAddress(ContractAddr), big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(data))
	return c.singAndSend(tx, nil)
}

//https://ethereum.stackexchange.com/questions/16472/signing-a-raw-transaction-in-go
func (c ClientETH) fill(addresses []string, amounts []*big.Int) (string, error) {
	nonce, gasPrice, err := c.getNonceGas(nil)
	if err != nil {
		return "", err
	}

	l := len(addresses)
	if l != len(amounts) {
		return "", errors.New("both arrays must match length")
	}

	stringAmounts := ""
	total := big.NewInt(0)
	for _, v := range amounts {
		total = total.Add(total, v)
		stringAmounts = stringAmounts + padBigInt(v)
	}

	stringAddresses := ""
	for _, v := range addresses {
		stringAddresses = stringAddresses + padString32(v)
	}

	gasLimit := uint64(8000000) //8mio gas

	//64 is the index of the first array, its always 64
	//second is 3 fix and 3 for the address array
	var data = "158c29c7" + padInt64(64) + padInt64(int64(32*(3+l))) + padInt64(int64(l)) + stringAddresses + padInt64(int64(l)) + stringAmounts
	tx := types.NewTransaction(nonce, common.HexToAddress(ContractAddr), total, gasLimit, gasPrice, common.Hex2Bytes(data))

	return c.singAndSend(tx, nil)
}

func (c ClientETH) getNonceGas(from *common.Address) (uint64, *big.Int, error) {
	if from == nil {
		from = &c.fromAddress
	}
	nonce, err := c.c.PendingNonceAt(context.Background(), *from)
	if err != nil {
		return 0, nil, err
	}

	gasPrice, err := c.c.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, nil, err
	}

	return nonce, gasPrice, nil
}

func (c ClientETH) singAndSend(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (string, error) {
	if privateKey == nil {
		privateKey = c.privateKey
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainId), privateKey)
	if err != nil {
		return "", err
	}

	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", err
	}
	var ret string
	err = c.rpc.CallContext(context.Background(), &ret, "eth_sendRawTransaction", hexutil.Encode(data))

	if err != nil {
		return "", err
	}

	return ret, nil
}

func padString32(hex string) string {
	if strings.Index(hex, "0x") == 0 {
		hex = hex[2:]
	}
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	padded := common.LeftPadBytes(common.Hex2Bytes(hex), 32)
	return common.Bytes2Hex(padded)
}

func padBigInt(b *big.Int) string {
	hex := b.Text(16)
	return padString32(hex)
}

func padInt64(i int64) string {
	hex := strconv.FormatInt(i, 16)
	return padString32(hex)
}
