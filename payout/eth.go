package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"time"
)

// PayoutEthMetaData contains all metadata concerning the PayoutEth contract.
var PayoutEthMetaData = &bind.MetaData{
	ABI: "[\n\t{\n\t\t\"inputs\": [],\n\t\t\"stateMutability\": \"nonpayable\",\n\t\t\"type\": \"constructor\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address payable\",\n\t\t\t\t\"name\": \"newOwner\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"changeOwner\",\n\t\t\"outputs\": [],\n\t\t\"stateMutability\": \"nonpayable\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"userId\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"totalPayedOut\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"getClaimableAmount\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"userId\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"getPayedOut\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [],\n\t\t\"name\": \"owner\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address\",\n\t\t\t\t\"name\": \"\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"payedOut\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address payable\",\n\t\t\t\t\"name\": \"receiver\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"amount\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"sndRecoverEth\",\n\t\t\"outputs\": [],\n\t\t\"stateMutability\": \"nonpayable\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address\",\n\t\t\t\t\"name\": \"receiver\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"address\",\n\t\t\t\t\"name\": \"contractAddress\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"amount\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"sndRecoverToken\",\n\t\t\"outputs\": [],\n\t\t\"stateMutability\": \"nonpayable\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address payable\",\n\t\t\t\t\"name\": \"dev\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"userId\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"totalPayedOut\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint8\",\n\t\t\t\t\"name\": \"v\",\n\t\t\t\t\"type\": \"uint8\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"r\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"bytes32\",\n\t\t\t\t\"name\": \"s\",\n\t\t\t\t\"type\": \"bytes32\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"withdraw\",\n\t\t\"outputs\": [],\n\t\t\"stateMutability\": \"nonpayable\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"stateMutability\": \"payable\",\n\t\t\"type\": \"receive\"\n\t}\n]",
	Sigs: map[string]string{
		"a6f9dae1": "changeOwner(address)",
		"db6e81ef": "getClaimableAmount(bytes32,uint256)",
		"8e0fb98d": "getPayedOut(bytes32)",
		"8da5cb5b": "owner()",
		"4c293714": "payedOut(bytes32)",
		"1b31a37f": "sndRecoverEth(address,uint256)",
		"74214d41": "sndRecoverToken(address,address,uint256)",
		"71676bd6": "withdraw(address,bytes32,uint256,uint8,bytes32,bytes32)",
	},
	Bin: "0x608060405234801561001057600080fd5b50600180546001600160a01b03191633179055610772806100326000396000f3fe60806040526004361061007f5760003560e01c80638da5cb5b1161004e5780638da5cb5b1461012d5780638e0fb98d14610165578063a6f9dae114610192578063db6e81ef146101b257600080fd5b80631b31a37f1461008b5780634c293714146100ad57806371676bd6146100ed57806374214d411461010d57600080fd5b3661008657005b600080fd5b34801561009757600080fd5b506100ab6100a636600461060b565b6101d2565b005b3480156100b957600080fd5b506100da6100c836600461069a565b60006020819052908152604090205481565b6040519081526020015b60405180910390f35b3480156100f957600080fd5b506100ab6101083660046105a9565b610240565b34801561011957600080fd5b506100ab610128366004610637565b61041b565b34801561013957600080fd5b5060015461014d906001600160a01b031681565b6040516001600160a01b0390911681526020016100e4565b34801561017157600080fd5b506100da61018036600461069a565b60009081526020819052604090205490565b34801561019e57600080fd5b506100ab6101ad36600461058c565b6104d1565b3480156101be57600080fd5b506100da6101cd3660046106b3565b61051d565b6001546001600160a01b031633146102055760405162461bcd60e51b81526004016101fc906106d5565b60405180910390fd5b6040516001600160a01b0383169082156108fc029083906000818181858888f1935050505015801561023b573d6000803e3d6000fd5b505050565b600085815260208190526040902054841161029d5760405162461bcd60e51b815260206004820152601c60248201527f4e6f206e65772066756e647320746f2062652077697468647261776e0000000060448201526064016101fc565b6001805460408051602081018990529081018790526001600160a01b03909116919060600160408051601f198184030181529082905280516020918201207f19457468657265756d205369676e6564204d6573736167653a0a36360000000091830191909152603c820152605c0160408051601f198184030181528282528051602091820120600084529083018083525260ff871690820152606081018590526080810184905260a0016020604051602081039080840390855afa158015610369573d6000803e3d6000fd5b505050602060405103516001600160a01b0316146103be5760405162461bcd60e51b81526020600482015260126024820152710a6d2cedcc2e8eae4ca40dcde40dac2e8c6d60731b60448201526064016101fc565b60008581526020819052604090208054908590556001600160a01b0387166108fc6103e983886106ff565b6040518115909202916000818181858888f19350505050158015610411573d6000803e3d6000fd5b5050505050505050565b6001546001600160a01b031633146104455760405162461bcd60e51b81526004016101fc906106d5565b60405163a9059cbb60e01b81526001600160a01b0384811660048301526024820183905283919082169063a9059cbb90604401602060405180830381600087803b15801561049257600080fd5b505af11580156104a6573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104ca9190610678565b5050505050565b6001546001600160a01b031633146104fb5760405162461bcd60e51b81526004016101fc906106d5565b600180546001600160a01b0319166001600160a01b0392909216919091179055565b60008281526020819052604081205482101561056c5760405162461bcd60e51b815260206004820152600e60248201526d4e656761746976652066756e647360901b60448201526064016101fc565b60008381526020819052604090205461058590836106ff565b9392505050565b60006020828403121561059e57600080fd5b813561058581610724565b60008060008060008060c087890312156105c257600080fd5b86356105cd81610724565b95506020870135945060408701359350606087013560ff811681146105f157600080fd5b9598949750929560808101359460a0909101359350915050565b6000806040838503121561061e57600080fd5b823561062981610724565b946020939093013593505050565b60008060006060848603121561064c57600080fd5b833561065781610724565b9250602084013561066781610724565b929592945050506040919091013590565b60006020828403121561068a57600080fd5b8151801515811461058557600080fd5b6000602082840312156106ac57600080fd5b5035919050565b600080604083850312156106c657600080fd5b50508035926020909101359150565b60208082526010908201526f27379030baba3437b934bd30ba34b7b760811b604082015260600190565b60008282101561071f57634e487b7160e01b600052601160045260246000fd5b500390565b6001600160a01b038116811461073957600080fd5b5056fea2646970667358221220556a5e5683dc8cbae0a092e52a6ce177fe2c22c6822b1759c143ec87b37a37f164736f6c63430008070033",
}

type ClientETH struct {
	c           *ethclient.Client
	rpc         *rpc.Client
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	fromAddress common.Address
	chainId     *big.Int
	contract    *bind.BoundContract
}

func getEthClient(ethUrl string, hexPrivateKey string, ethContract string) (*ClientETH, error) {
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

	parsed, err := PayoutEthMetaData.GetAbi()
	if err != nil {
		log.Fatal(err)
	}

	c.contract = bind.NewBoundContract(common.HexToAddress(ethContract), *parsed, c.c, c.c, c.c)

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
