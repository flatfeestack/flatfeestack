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
	//curl --data '{"method":"eth_getBalance","params":["<ETH>"],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST localhost:8545
	//https://openethereum.github.io/JSONRPC-eth-module#eth_gettransactionreceipt
	ContractCode = "608060405234801561001057600080fd5b50600180546001600160a01b03191633179055610ab8806100326000396000f3fe60806040526004361061004a5760003560e01c8063158c29c71461004f57806370a082311461006457806386d1a69f146100c0578063b2b57708146100c8578063d0b48509146100e8575b600080fd5b61006261005d3660046108ad565b610146565b005b34801561007057600080fd5b506100a361007f366004610852565b6001600160a01b03166000908152602081905260409020546001600160c01b031690565b6040516001600160c01b0390911681526020015b60405180910390f35b610062610479565b3480156100d457600080fd5b506100626100e3366004610873565b61059e565b3480156100f457600080fd5b5061012e610103366004610852565b6001600160a01b0316600090815260208190526040902054600160c01b90046001600160401b031690565b6040516001600160401b0390911681526020016100b7565b6001546001600160a01b031633146101b05760405162461bcd60e51b815260206004820152602260248201527f4f6e6c7920746865206f776e65722063616e20616464206e6577207061796f75604482015261747360f01b60648201526084015b60405180910390fd5b80518251146102205760405162461bcd60e51b815260206004820152603660248201527f41646472657373657320616e642062616c616e636573206172726179206d75736044820152750e840d0c2ecca40e8d0ca40e6c2daca40d8cadccee8d60531b60648201526084016101a7565b6000805b83518161ffff16101561041157828161ffff168151811061025557634e487b7160e01b600052603260045260246000fd5b6020026020010151600080868461ffff168151811061028457634e487b7160e01b600052603260045260246000fd5b6020908102919091018101516001600160a01b03168252810191909152604001600090812080549091906102c29084906001600160c01b03166109cf565b92506101000a8154816001600160c01b0302191690836001600160c01b03160217905550600080858361ffff168151811061030d57634e487b7160e01b600052603260045260246000fd5b6020908102919091018101516001600160a01b0316825281019190915260400160002054600160c01b90046001600160401b03166103bd5742600080868461ffff168151811061036d57634e487b7160e01b600052603260045260246000fd5b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002060000160186101000a8154816001600160401b0302191690836001600160401b031602179055505b828161ffff16815181106103e157634e487b7160e01b600052603260045260246000fd5b60200260200101516001600160c01b0316826103fd91906109fa565b91508061040981610a34565b915050610224565b503481146104745760405162461bcd60e51b815260206004820152602a60248201527f53756d206f662062616c616e63657320697320686967686572207468616e2070604482015269185a5908185b5bdd5b9d60b21b60648201526084016101a7565b505050565b336000908152602081905260409020546001600160c01b03166104ee5760405162461bcd60e51b815260206004820152602760248201527f5061796d656e7453706c69747465723a206163636f756e7420686173206e6f2060448201526662616c616e636560c81b60648201526084016101a7565b3360008181526020819052604080822080549083905590516001600160c01b03821693600160c01b9092046001600160401b0316926108fc851502918591818181858888f19350505050158015610549573d6000803e3d6000fd5b50604080513381526001600160c01b03841660208201526001600160401b0383168183015290517fd60024301b81030e9c8bc232c048dd52e4a5ff8fb60493c637091ae17b1580949181900360600190a15050565b6001546001600160a01b0316331461060d5760405162461bcd60e51b815260206004820152602c60248201527f4f6e6c7920746865206f776e65722063616e20636f6c6c65637420756e636c6160448201526b696d6564207061796f75747360a01b60648201526084016101a7565b60005b81518161ffff1610156107bc5742600080848461ffff168151811061064557634e487b7160e01b600052603260045260246000fd5b6020908102919091018101516001600160a01b031682528101919091526040016000205461068790600160c01b90046001600160401b03166301e13380610a12565b6001600160401b031610156107aa57600080838361ffff16815181106106bd57634e487b7160e01b600052603260045260246000fd5b6020908102919091018101516001600160a01b031682528181019290925260409081016000908120543382529281905290812080546001600160c01b039384169391929161070d918591166109cf565b92506101000a8154816001600160c01b0302191690836001600160c01b031602179055506000806000848461ffff168151811061075a57634e487b7160e01b600052603260045260246000fd5b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002060000160006101000a8154816001600160c01b0302191690836001600160c01b031602179055505b806107b481610a34565b915050610610565b5050565b80356001600160a01b03811681146107d757600080fd5b919050565b600082601f8301126107ec578081fd5b813560206108016107fc836109ac565b61097c565b80838252828201915082860187848660051b8901011115610820578586fd5b855b8581101561084557610833826107c0565b84529284019290840190600101610822565b5090979650505050505050565b600060208284031215610863578081fd5b61086c826107c0565b9392505050565b600060208284031215610884578081fd5b81356001600160401b03811115610899578182fd5b6108a5848285016107dc565b949350505050565b600080604083850312156108bf578081fd5b82356001600160401b03808211156108d5578283fd5b6108e1868387016107dc565b93506020915081850135818111156108f7578384fd5b85019050601f81018613610909578283fd5b80356109176107fc826109ac565b80828252848201915084840189868560051b8701011115610936578687fd5b8694505b8385101561096c5780356001600160c01b0381168114610958578788fd5b83526001949094019391850191850161093a565b5080955050505050509250929050565b604051601f8201601f191681016001600160401b03811182821017156109a4576109a4610a6c565b604052919050565b60006001600160401b038211156109c5576109c5610a6c565b5060051b60200190565b60006001600160c01b038281168482168083038211156109f1576109f1610a56565b01949350505050565b60008219821115610a0d57610a0d610a56565b500190565b60006001600160401b038083168185168083038211156109f1576109f1610a56565b600061ffff80831681811415610a4c57610a4c610a56565b6001019392505050565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052604160045260246000fdfea2646970667358221220039ba2a3f7d7e3e2335ed6fea1f5854dc305b9ffac1ace5ead46a9a4544e62be64736f6c63430008040033"
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

// TransactionReceipt the struct for a transaction receipt on Ethereum
type TransactionReceipt struct {
	TransactionHash   string
	TransactionIndex  string
	BlockHash         string
	BlockNumber       string
	CumulativeGasUsed string
	GasUsed           string
	ContractAddress   string
	Logs              []Log
	LogsBloom         string
	Root              string
	Status            string
}

type Log struct {
	Removed          bool
	LogIndex         string
	TransactionIndex string
	TransactionHash  string
	BlockNumber      string
	BlockHash        string
	Address          string
	Data             string
	Topics           []string
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

func (c ClientETH) deploy(contract string) (string, error) {
	nonce, gasPrice, err := c.getNonceGas(nil)
	if err != nil {
		return "", err
	}

	tx := types.NewContractCreation(nonce, big.NewInt(0), 867749, gasPrice, common.Hex2Bytes(contract))
	addr, err := c.singAndSend(tx, nil)
	if err != nil {
		return "", err
	}
	return c.txReceipt(addr)
}

// GetTransactionReceipt returns the transaction receipt, noce the tx is mined
func (c ClientETH) txReceipt(receipt string) (string, error) {
	var raw TransactionReceipt
	err := c.rpc.Call(&raw, "eth_getTransactionReceipt", receipt)

	if err != nil {
		return "", err
	}

	if raw.Status == "0x1" {
		return raw.ContractAddress, nil
	}
	return "", errors.New("tx not fonud")
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
