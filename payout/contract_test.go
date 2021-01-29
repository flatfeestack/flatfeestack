// This file is a generated binding and any manual changes will be lost.

package main

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"os"
	"testing"
)

var (
	c                 *ClientETH
	testPrivateKeyHex = "36ae1cf18fa77be025d3668ddcfeb4ab6c227d66b82473db4bd58fec37ce74df"
	testPublicKeyHex  = "0x48Fe4A98911ae45648d9a17aAD5E209DAadF7559"
	testPrivateKey    *ecdsa.PrivateKey
	testPublicKey     *ecdsa.PublicKey
)

func TestMain(m *testing.M) {
	var err error

	testPrivateKey, err = crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	publicKeyGeneric := testPrivateKey.Public()
	testPublicKey, _ = publicKeyGeneric.(*ecdsa.PublicKey)

	testPrivateKeyHex = common.Bytes2Hex(crypto.FromECDSA(testPrivateKey))
	testPublicKeyHex = crypto.PubkeyToAddress(*testPublicKey).Hex()

	//also connect with http://remix.ethereum.org
	c, err = NewClientETH("http://172.17.0.1:8545", "4d5db4107d237df6a3d58ee5f70ae63d73d7658d4026f2eefd2f204c81682cb7")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	c.c.Close()
	os.Exit(code)
}

func TestFill(t *testing.T) {
	var err error
	ContractAddr, err = c.deploy(ContractCode)
	assert.Nil(t, err)
	a := []string{"0xd9145CCE52D386f254917e481eB44e9943F39138", "0xDA0bab807633f07f013f94DD0E6A4F96F8742B53", "0x9D7f74d0C41E726EC95884E0e97Fa6129e3b5E99"}
	v := []*big.Int{big.NewInt(1234), big.NewInt(1235), big.NewInt(1236)}
	_, err = c.fill(a, v)
	assert.Nil(t, err)

	b, err := c.balanceOf("0xd9145CCE52D386f254917e481eB44e9943F39138")
	assert.Nil(t, err)

	z := new(big.Int)
	z.SetBytes(b[0:32])

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(1234), z)
}

func TestFill2(t *testing.T) {
	var err error
	ContractAddr, err = c.deploy(ContractCode)
	assert.Nil(t, err)
	a := []string{"0xd9145CCE52D386f254917e481eB44e9943F39138", "0xDA0bab807633f07f013f94DD0E6A4F96F8742B53"}
	v := []*big.Int{big.NewInt(1234), big.NewInt(1235)}
	_, err = c.fill(a, v)
	assert.Nil(t, err)

	b, err := c.balanceOf("0xd9145CCE52D386f254917e481eB44e9943F39138")
	assert.Nil(t, err)

	z := new(big.Int)
	z.SetBytes(b[0:32])

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(1234), z)
}

func TestRelease(t *testing.T) {
	var err error
	ContractAddr, err = c.deploy(ContractCode)
	assert.Nil(t, err)

	_, err = sendEth(testPublicKeyHex)
	assert.Nil(t, err)

	a := []string{testPublicKeyHex, "0xDA0bab807633f07f013f94DD0E6A4F96F8742B53"}
	v := []*big.Int{big.NewInt(1234), big.NewInt(1235)}
	_, err = c.fill(a, v)
	assert.Nil(t, err)

	_, err = c.release(testPrivateKey)
	assert.Nil(t, err)

	balance, err := c.c.BalanceAt(context.Background(), common.HexToAddress(testPublicKeyHex), nil)

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(1000000000000+1234), balance)
}

func (c ClientETH) release(privateKey *ecdsa.PrivateKey) (string, error) {
	a := common.HexToAddress(testPublicKeyHex)
	nonce, gasPrice, err := c.getNonceGas(&a)
	if err != nil {
		return "", err
	}

	gasLimit := uint64(1000000) //1mio gas
	var data = "86d1a69f"
	tx := types.NewTransaction(nonce, common.HexToAddress(ContractAddr), big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(data))

	return c.singAndSend(tx, privateKey)
}

func sendEth(to string) (string, error) {
	nonce, gasPrice, err := c.getNonceGas(nil)
	if err != nil {
		return "", err
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(to), big.NewInt(1000000000000), 21000, gasPrice, nil)
	return c.singAndSend(tx, nil)
}
