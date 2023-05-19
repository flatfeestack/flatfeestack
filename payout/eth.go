package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"time"
)

type ClientETH struct {
	c          *ethclient.Client
	rpc        *rpc.Client
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	chainId    *big.Int
	contract   *bind.BoundContract
}

func getEthClient(ethUrl string, hexPrivateKey string) (*ClientETH, error) {
	dialContext, err := rpc.DialContext(context.Background(), ethUrl)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(dialContext)

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
		c:          client,
		rpc:        dialContext,
		privateKey: privateKey,
		publicKey:  publicKeyECDSA,
	}

	chainId, err := c.c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	c.chainId = chainId

	fmt.Println("---------------------------------")
	fmt.Printf("My chain Id is %v\n", chainId)
	fmt.Printf("My address is %v\n", fromAddress)
	fmt.Println("---------------------------------")

	// get time
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	//forward to the time we should be
	t := time.Unix(int64(header.Time), 0)
	diff := timeNow().Sub(t)
	diffSec := diff.Milliseconds() / 1000
	if diffSec > 0 {
		warpChain(int(diffSec), dialContext)
	}

	// show time
	header, err = client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("---------------------------------")
	fmt.Printf("Chain  time is %v\n", time.Unix(int64(header.Time), 0))
	fmt.Printf("Server time is %v\n", timeNow())
	fmt.Println("---------------------------------")

	return c, nil
}

func warpChain(seconds int, rpc *rpc.Client) error {
	//we need to forward the time on the chain, every 15s a block, so now we push a lot of blocks...
	mineNrBlocks := seconds / 15

	if debug {
		for i := 0; i < mineNrBlocks; i++ {
			var result hexutil.Big
			err := rpc.CallContext(context.Background(), &result, "evm_mine")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getEthSignature(data PayoutRequest, symbol string) (PayoutResponse, error) {
	var arguments abi.Arguments
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.StringTy},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{Size: 256, T: abi.UintTy},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.StringTy},
	})

	privateKey, err := crypto.HexToECDSA(opts.Ethereum.PrivateKey)
	if err != nil {
		return PayoutResponse{}, err
	}

	encodedUserId := [32]byte(crypto.Keccak256([]byte(data.UserId.String())))
	packed, err := arguments.Pack(encodedUserId, "#", data.Amount, symbol)
	hashRaw := crypto.Keccak256(packed)

	// Add Ethereum Signed Message prefix to hash
	prefix := []byte("\x19Ethereum Signed Message:\n32")
	prefixedHash := crypto.Keccak256(append(prefix, hashRaw[:]...))

	signature, err := crypto.Sign(prefixedHash[:], privateKey)
	if err != nil {
		return PayoutResponse{}, err
	}

	return PayoutResponse{
		Amount:        data.Amount,
		Currency:      symbol,
		EncodedUserId: hexutil.Encode(encodedUserId[:]),
		Signature:     hexutil.Encode(signature),
	}, nil
}
