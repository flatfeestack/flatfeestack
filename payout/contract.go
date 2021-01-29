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
	ContractCode = "608060405234801561001057600080fd5b50600180546001600160a01b0319163317905561096a806100326000396000f3fe60806040526004361061003f5760003560e01c8063158c29c71461004457806370a082311461016d57806386d1a69f146101cc578063b2b57708146101d4575b600080fd5b61016b6004803603604081101561005a57600080fd5b81019060208101813564010000000081111561007557600080fd5b82018360208201111561008757600080fd5b803590602001918460208302840111640100000000831117156100a957600080fd5b91908080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525092959493602081019350359150506401000000008111156100f957600080fd5b82018360208201111561010b57600080fd5b8035906020019184602083028401116401000000008311171561012d57600080fd5b919080806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250929550610284945050505050565b005b34801561017957600080fd5b506101a06004803603602081101561019057600080fd5b50356001600160a01b0316610509565b604080516001600160c01b03909316835267ffffffffffffffff90911660208301528051918290030190f35b61016b610541565b3480156101e057600080fd5b5061016b600480360360208110156101f757600080fd5b81019060208101813564010000000081111561021257600080fd5b82018360208201111561022457600080fd5b8035906020019184602083028401116401000000008311171561024657600080fd5b919080806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250929550610647945050505050565b6001546001600160a01b031633146102cd5760405162461bcd60e51b81526004018080602001828103825260228152602001806108e76022913960400191505060405180910390fd5b805182511461030d5760405162461bcd60e51b81526004018080602001828103825260368152602001806108876036913960400191505060405180910390fd5b6000805b83518161ffff1610156104c55761037f838261ffff168151811061033157fe5b6020026020010151600080878561ffff168151811061034c57fe5b6020908102919091018101516001600160a01b03168252810191909152604001600020546001600160c01b0316906107f2565b600080868461ffff168151811061039257fe5b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002060000160006101000a8154816001600160c01b0302191690836001600160c01b03160217905550600080858361ffff16815181106103f457fe5b6020908102919091018101516001600160a01b0316825281019190915260400160002054600160c01b900467ffffffffffffffff166104935742600080868461ffff168151811061044157fe5b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002060000160186101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055505b828161ffff16815181106104a357fe5b60200260200101516001600160c01b0316820191508080600101915050610311565b503481146105045760405162461bcd60e51b815260040180806020018281038252602a8152602001806108bd602a913960400191505060405180910390fd5b505050565b6001600160a01b03166000908152602081905260409020546001600160c01b03811691600160c01b90910467ffffffffffffffff1690565b336000908152602081905260409020546001600160c01b03166105955760405162461bcd60e51b81526004018080602001828103825260278152602001806108606027913960400191505060405180910390fd5b3360008181526020819052604080822080549083905590516001600160c01b03821693600160c01b90920467ffffffffffffffff16926108fc851502918591818181858888f193505050501580156105f1573d6000803e3d6000fd5b50604080513381526001600160c01b038416602082015267ffffffffffffffff83168183015290517fd60024301b81030e9c8bc232c048dd52e4a5ff8fb60493c637091ae17b1580949181900360600190a15050565b6001546001600160a01b031633146106905760405162461bcd60e51b815260040180806020018281038252602c815260200180610909602c913960400191505060405180910390fd5b60005b81518161ffff1610156107ee5742600080848461ffff16815181106106b457fe5b6020908102919091018101516001600160a01b03168252810191909152604001600020546301e1338067ffffffffffffffff600160c01b9092048216011610156107e657610752600080848461ffff168151811061070e57fe5b6020908102919091018101516001600160a01b031682528181019290925260409081016000908120543382529281905220546001600160c01b0390811691166107f2565b33600090815260208190526040812080546001600160c01b0319166001600160c01b039390931692909217909155825181908190859061ffff861690811061079657fe5b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002060000160006101000a8154816001600160c01b0302191690836001600160c01b031602179055505b600101610693565b5050565b60008282016001600160c01b038085169082161015610858576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b939250505056fe5061796d656e7453706c69747465723a206163636f756e7420686173206e6f2062616c616e636541646472657373657320616e642062616c616e636573206172726179206d7573742068617665207468652073616d65206c656e67746853756d206f662062616c616e63657320697320686967686572207468616e207061696420616d6f756e744f6e6c7920746865206f776e65722063616e20616464206e6577207061796f7574734f6e6c7920746865206f776e65722063616e20636f6c6c65637420756e636c61696d6564207061796f757473a26469706673582212205a5eda04cb5ebe9aa433a2b7c2febda189d7e0bf6fa10451a2b8987f9cab4cde64736f6c63430007040033"
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
