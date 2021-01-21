// This file is a generated binding and any manual changes will be lost.

package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
)

var (
	ContractAddr = "0xD6f396CB541eee419Ca495324d1EDE3EAe857179"
)

// Reference imports to suppress errors if they are not otherwise used.
type ClientETH struct {
	c           *ethclient.Client
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	fromAddress common.Address
}

func NewClientETH(ethUrl string, hexPrivateKey string) (*ClientETH, error) {
	client, err := ethclient.Dial(ethUrl)
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

	return &ClientETH{
		c:           client,
		privateKey:  privateKey,
		publicKey:   publicKeyECDSA,
		fromAddress: fromAddress,
	}, nil
}

//https://ethereum.stackexchange.com/questions/16472/signing-a-raw-transaction-in-go
func (c ClientETH) fill(addresses []string, amounts []*big.Int) error {

	nonce, err := c.c.PendingNonceAt(context.Background(), c.fromAddress)
	if err != nil {
		return err
	}

	if len(addresses) != len(amounts) {
		return errors.New("both arrays must match length")
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

	gasLimit := uint64(8000000) //10mio gas
	gasPrice, err := c.c.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	var data = []byte("158c29c7" + stringAddresses + stringAmounts)
	tx := types.NewTransaction(nonce, common.HexToAddress(ContractAddr), total, gasLimit, gasPrice, data)

	chainID, err := c.c.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return err
	}

	err = c.c.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	return nil
}

func (c ClientETH) deploy(contract string) error {
	tx := types.NewContractCreation(0, big.NewInt(0), 8000000, big.NewInt(1), common.Hex2Bytes(contract))

	chainID, err := c.c.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return err
	}

	err = c.c.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	return nil
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
