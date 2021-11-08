// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

import (
	"errors"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// PayoutEthEvalMetaData contains all meta data concerning the PayoutEthEval contract.
var PayoutEthEvalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"addresspayable[]\",\"name\":\"_devs\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_teas\",\"type\":\"uint256[]\"}],\"name\":\"batchPayout\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable[]\",\"name\":\"_devs\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_teasForPayout\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"}],\"name\":\"batchPayoutServiceFeeWithPayout\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable[]\",\"name\":\"_devs\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_teasToStore\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"}],\"name\":\"batchPayoutServiceFeeWithStore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"changeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"}],\"name\":\"getConcat\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"}],\"name\":\"getConcatHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dev\",\"type\":\"address\"}],\"name\":\"getTotalEarnedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"concatDevTea\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"concatDevTeaHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"recoverSignerHash\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"withdrawHashed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"withdrawHashedRequire\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"withdrawNotHashed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"d91966a7": "batchPayout(address[],uint256[])",
		"3e019839": "batchPayoutServiceFeeWithPayout(address[],uint256[],uint256)",
		"64e0e904": "batchPayoutServiceFeeWithStore(address[],uint256[],uint256)",
		"a6f9dae1": "changeOwner(address)",
		"b6b55f25": "deposit(uint256)",
		"b0b4bfe8": "getConcat(address,uint256)",
		"e41a3766": "getConcatHash(address,uint256)",
		"e6423dcc": "getTotalEarnedAmount(address)",
		"8da5cb5b": "owner()",
		"7c4895fa": "recoverSigner(bytes,uint8,bytes32,bytes32)",
		"6abf4f14": "recoverSignerHash(bytes32,uint8,bytes32,bytes32)",
		"f362a99b": "withdrawHashed(address,uint256,uint8,bytes32,bytes32)",
		"377c91f4": "withdrawHashedRequire(address,uint256,uint8,bytes32,bytes32)",
		"9bb442bd": "withdrawNotHashed(address,uint256,uint8,bytes32,bytes32)",
	},
	Bin: "0x608060405234801561001057600080fd5b50600180546001600160a01b031916331790556112b4806100326000396000f3fe6080604052600436106100dd5760003560e01c8063a6f9dae11161007f578063d91966a711610059578063d91966a714610241578063e41a376614610261578063e6423dcc1461028f578063f362a99b146102c557600080fd5b8063a6f9dae1146101e1578063b0b4bfe814610201578063b6b55f251461022e57600080fd5b80636abf4f14116100bb5780636abf4f14146101445780637c4895fa146101815780638da5cb5b146101a15780639bb442bd146101c157600080fd5b8063377c91f4146100e25780633e0198391461010457806364e0e90414610124575b600080fd5b3480156100ee57600080fd5b506101026100fd366004610cd7565b6102e5565b005b34801561011057600080fd5b5061010261011f366004610e61565b610483565b34801561013057600080fd5b5061010261013f366004610e61565b610624565b34801561015057600080fd5b5061016461015f366004610ece565b6107ca565b6040516001600160a01b0390911681526020015b60405180910390f35b34801561018d57600080fd5b5061016461019c366004610f09565b610898565b3480156101ad57600080fd5b50600154610164906001600160a01b031681565b3480156101cd57600080fd5b506101026101dc366004610cd7565b6108e8565b3480156101ed57600080fd5b506101026101fc366004610fbf565b6109a1565b34801561020d57600080fd5b5061022161021c366004610fe3565b6109ed565b604051610178919061103b565b61010261023c36600461106e565b610a19565b34801561024d57600080fd5b5061010261025c366004611087565b610a8a565b34801561026d57600080fd5b5061028161027c366004610fe3565b610c1f565b604051908152602001610178565b34801561029b57600080fd5b506102816102aa366004610fbf565b6001600160a01b031660009081526020819052604090205490565b3480156102d157600080fd5b506101026102e0366004610cd7565b610c3a565b6001600160a01b03851660009081526020819052604090205484116103255760405162461bcd60e51b815260040161031c906110eb565b60405180910390fd5b6000858560405160200161033a929190611133565b6040516020818303038152906040528051906020012090506000610360828686866107ca565b6001549091506001600160a01b038083169116146103e65760405162461bcd60e51b815260206004820152603760248201527f5369676e617475726520646f6573206e6f74206d61746368206f776e6572206160448201527f6e642070726f766964656420706172616d65746572732e000000000000000000606482015260840161031c565b6001600160a01b0387166000818152602081905260408120805490899055916108fc610412848b61116b565b6040518115909202916000818181858888f193505050509050806104785760405162461bcd60e51b815260206004820152601c60248201527f5472616e7366657220776173206e6f74207375636365737366756c2e00000000604482015260640161031c565b505050505050505050565b6001546001600160a01b031633146104ad5760405162461bcd60e51b815260040161031c90611182565b81518351146104ce5760405162461bcd60e51b815260040161031c906111ad565b60005b835181101561061e5760008060008684815181106104f1576104f16111e4565b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002054905083828151811061052e5761052e6111e4565b60200260200101518110610542575061060c565b82848381518110610555576105556111e4565b602002602001015161056791906111fa565b60008087858151811061057c5761057c6111e4565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020819055508482815181106105ba576105ba6111e4565b60200260200101516001600160a01b03166108fc828685815181106105e1576105e16111e4565b60200260200101516105f3919061116b565b6040518115909202916000818181858888f15050505050505b8061061681611212565b9150506104d1565b50505050565b6001546001600160a01b0316331461064e5760405162461bcd60e51b815260040161031c90611182565b815183511461066f5760405162461bcd60e51b815260040161031c906111ad565b60005b835181101561061e576000806000868481518110610692576106926111e4565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020549050828483815181106106d0576106d06111e4565b60200260200101516106e2919061116b565b81106106ee57506107b8565b838281518110610700576107006111e4565b602002602001015160008087858151811061071d5761071d6111e4565b60200260200101516001600160a01b03166001600160a01b031681526020019081526020016000208190555084828151811061075b5761075b6111e4565b60200260200101516001600160a01b03166108fc8483878681518110610783576107836111e4565b6020026020010151610795919061116b565b61079f919061116b565b6040518115909202916000818181858888f15050505050505b806107c281611212565b915050610672565b6000806040518060400160405280601c81526020017f19457468657265756d205369676e6564204d6573736167653a0a36360000000081525090506000818760405160200161081a92919061122d565b60408051601f1981840301815282825280516020918201206000845290830180835281905260ff8916918301919091526060820187905260808201869052915060019060a0016020604051602081039080840390855afa158015610882573d6000803e3d6000fd5b5050604051601f19015198975050505050505050565b6000806040518060400160405280601d81526020017f19457468657265756d205369676e6564204d6573736167653a0a31303600000081525090506000818760405160200161081a92919061124f565b6001600160a01b038516600090815260208190526040902054841161091f5760405162461bcd60e51b815260040161031c906110eb565b60008585604051602001610934929190611133565b6040516020818303038152906040529050600061095382868686610898565b6001549091506001600160a01b0380831691161415610998576001600160a01b0387166000818152602081905260408120805490899055916108fc610412848b61116b565b50505050505050565b6001546001600160a01b031633146109cb5760405162461bcd60e51b815260040161031c90611182565b600180546001600160a01b0319166001600160a01b0392909216919091179055565b60608282604051602001610a02929190611133565b604051602081830303815290604052905092915050565b803414610a875760405162461bcd60e51b815260206004820152603660248201527f4d6573736167652076616c7565206d757374206d61746368207468652070726f6044820152753b34b232b2103830b930b6b2ba32b9103b30b63ab29760511b606482015260840161031c565b50565b6001546001600160a01b03163314610ab45760405162461bcd60e51b815260040161031c90611182565b8051825114610ad55760405162461bcd60e51b815260040161031c906111ad565b60005b8251811015610c1a576000806000858481518110610af857610af86111e4565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020549050828281518110610b3557610b356111e4565b60200260200101518110610b495750610c08565b828281518110610b5b57610b5b6111e4565b6020026020010151600080868581518110610b7857610b786111e4565b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002081905550838281518110610bb657610bb66111e4565b60200260200101516001600160a01b03166108fc82858581518110610bdd57610bdd6111e4565b6020026020010151610bef919061116b565b6040518115909202916000818181858888f15050505050505b80610c1281611212565b915050610ad8565b505050565b6000610c2b83836109ed565b80519060200120905092915050565b6001600160a01b0385166000908152602081905260409020548411610c715760405162461bcd60e51b815260040161031c906110eb565b60008585604051602001610c86929190611133565b6040516020818303038152906040528051906020012090506000610953828686866107ca565b6001600160a01b0381168114610a8757600080fd5b803560ff81168114610cd257600080fd5b919050565b600080600080600060a08688031215610cef57600080fd5b8535610cfa81610cac565b945060208601359350610d0f60408701610cc1565b94979396509394606081013594506080013592915050565b634e487b7160e01b600052604160045260246000fd5b604051601f8201601f1916810167ffffffffffffffff81118282101715610d6657610d66610d27565b604052919050565b600067ffffffffffffffff821115610d8857610d88610d27565b5060051b60200190565b600082601f830112610da357600080fd5b81356020610db8610db383610d6e565b610d3d565b82815260059290921b84018101918181019086841115610dd757600080fd5b8286015b84811015610dfb578035610dee81610cac565b8352918301918301610ddb565b509695505050505050565b600082601f830112610e1757600080fd5b81356020610e27610db383610d6e565b82815260059290921b84018101918181019086841115610e4657600080fd5b8286015b84811015610dfb5780358352918301918301610e4a565b600080600060608486031215610e7657600080fd5b833567ffffffffffffffff80821115610e8e57600080fd5b610e9a87838801610d92565b94506020860135915080821115610eb057600080fd5b50610ebd86828701610e06565b925050604084013590509250925092565b60008060008060808587031215610ee457600080fd5b84359350610ef460208601610cc1565b93969395505050506040820135916060013590565b60008060008060808587031215610f1f57600080fd5b843567ffffffffffffffff80821115610f3757600080fd5b818701915087601f830112610f4b57600080fd5b8135602082821115610f5f57610f5f610d27565b610f71601f8301601f19168201610d3d565b92508183528981838601011115610f8757600080fd5b81818501828501376000818385010152829750610fa5818a01610cc1565b979a97995050505060408601359560600135949350505050565b600060208284031215610fd157600080fd5b8135610fdc81610cac565b9392505050565b60008060408385031215610ff657600080fd5b823561100181610cac565b946020939093013593505050565b60005b8381101561102a578181015183820152602001611012565b8381111561061e5750506000910152565b602081526000825180602084015261105a81604085016020870161100f565b601f01601f19169190910160400192915050565b60006020828403121561108057600080fd5b5035919050565b6000806040838503121561109a57600080fd5b823567ffffffffffffffff808211156110b257600080fd5b6110be86838701610d92565b935060208501359150808211156110d457600080fd5b506110e185828601610e06565b9150509250929050565b60208082526028908201527f54686573652066756e6473206861766520616c7265616479206265656e2077696040820152673a34323930bbb71760c11b606082015260800190565b60609290921b6bffffffffffffffffffffffff19168252601482015260340190565b634e487b7160e01b600052601160045260246000fd5b60008282101561117d5761117d611155565b500390565b60208082526011908201527027379030baba3437b934bd30ba34b7b71760791b604082015260600190565b6020808252601d908201527f417272617973206d75737420686176652073616d65206c656e6774682e000000604082015260600190565b634e487b7160e01b600052603260045260246000fd5b6000821982111561120d5761120d611155565b500190565b600060001982141561122657611226611155565b5060010190565b6000835161123f81846020880161100f565b9190910191825250602001919050565b6000835161126181846020880161100f565b83519083019061127581836020880161100f565b0194935050505056fea2646970667358221220bcb596b8990f4776221f854483f98332527e6284d6d25881193979adeee3f8ed64736f6c63430008090033",
}

// PayoutEthEvalABI is the input ABI used to generate the binding from.
// Deprecated: Use PayoutEthEvalMetaData.ABI instead.
var PayoutEthEvalABI = PayoutEthEvalMetaData.ABI

// Deprecated: Use PayoutEthEvalMetaData.Sigs instead.
// PayoutEthEvalFuncSigs maps the 4-byte function signature to its string representation.
var PayoutEthEvalFuncSigs = PayoutEthEvalMetaData.Sigs

// PayoutEthEvalBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PayoutEthEvalMetaData.Bin instead.
var PayoutEthEvalBin = PayoutEthEvalMetaData.Bin

// DeployPayoutEthEval deploys a new Ethereum contract, binding an instance of PayoutEthEval to it.
func DeployPayoutEthEval(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PayoutEthEval, error) {
	parsed, err := PayoutEthEvalMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PayoutEthEvalBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PayoutEthEval{PayoutEthEvalCaller: PayoutEthEvalCaller{contract: contract}, PayoutEthEvalTransactor: PayoutEthEvalTransactor{contract: contract}, PayoutEthEvalFilterer: PayoutEthEvalFilterer{contract: contract}}, nil
}

// PayoutEthEval is an auto generated Go binding around an Ethereum contract.
type PayoutEthEval struct {
	PayoutEthEvalCaller     // Read-only binding to the contract
	PayoutEthEvalTransactor // Write-only binding to the contract
	PayoutEthEvalFilterer   // Log filterer for contract events
}

// PayoutEthEvalCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayoutEthEvalCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthEvalTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayoutEthEvalTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthEvalFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayoutEthEvalFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthEvalSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayoutEthEvalSession struct {
	Contract     *PayoutEthEval    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PayoutEthEvalCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayoutEthEvalCallerSession struct {
	Contract *PayoutEthEvalCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// PayoutEthEvalTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayoutEthEvalTransactorSession struct {
	Contract     *PayoutEthEvalTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// PayoutEthEvalRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayoutEthEvalRaw struct {
	Contract *PayoutEthEval // Generic contract binding to access the raw methods on
}

// PayoutEthEvalCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayoutEthEvalCallerRaw struct {
	Contract *PayoutEthEvalCaller // Generic read-only contract binding to access the raw methods on
}

// PayoutEthEvalTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayoutEthEvalTransactorRaw struct {
	Contract *PayoutEthEvalTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayoutEthEval creates a new instance of PayoutEthEval, bound to a specific deployed contract.
func NewPayoutEthEval(address common.Address, backend bind.ContractBackend) (*PayoutEthEval, error) {
	contract, err := bindPayoutEthEval(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayoutEthEval{PayoutEthEvalCaller: PayoutEthEvalCaller{contract: contract}, PayoutEthEvalTransactor: PayoutEthEvalTransactor{contract: contract}, PayoutEthEvalFilterer: PayoutEthEvalFilterer{contract: contract}}, nil
}

// NewPayoutEthEvalCaller creates a new read-only instance of PayoutEthEval, bound to a specific deployed contract.
func NewPayoutEthEvalCaller(address common.Address, caller bind.ContractCaller) (*PayoutEthEvalCaller, error) {
	contract, err := bindPayoutEthEval(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutEthEvalCaller{contract: contract}, nil
}

// NewPayoutEthEvalTransactor creates a new write-only instance of PayoutEthEval, bound to a specific deployed contract.
func NewPayoutEthEvalTransactor(address common.Address, transactor bind.ContractTransactor) (*PayoutEthEvalTransactor, error) {
	contract, err := bindPayoutEthEval(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutEthEvalTransactor{contract: contract}, nil
}

// NewPayoutEthEvalFilterer creates a new log filterer instance of PayoutEthEval, bound to a specific deployed contract.
func NewPayoutEthEvalFilterer(address common.Address, filterer bind.ContractFilterer) (*PayoutEthEvalFilterer, error) {
	contract, err := bindPayoutEthEval(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayoutEthEvalFilterer{contract: contract}, nil
}

// bindPayoutEthEval binds a generic wrapper to an already deployed contract.
func bindPayoutEthEval(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PayoutEthEvalABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutEthEval *PayoutEthEvalRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutEthEval.Contract.PayoutEthEvalCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutEthEval *PayoutEthEvalRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.PayoutEthEvalTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutEthEval *PayoutEthEvalRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.PayoutEthEvalTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutEthEval *PayoutEthEvalCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutEthEval.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutEthEval *PayoutEthEvalTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutEthEval *PayoutEthEvalTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.contract.Transact(opts, method, params...)
}

// GetConcat is a free data retrieval call binding the contract method 0xb0b4bfe8.
//
// Solidity: function getConcat(address _dev, uint256 _tea) pure returns(bytes)
func (_PayoutEthEval *PayoutEthEvalCaller) GetConcat(opts *bind.CallOpts, _dev common.Address, _tea *big.Int) ([]byte, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "getConcat", _dev, _tea)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetConcat is a free data retrieval call binding the contract method 0xb0b4bfe8.
//
// Solidity: function getConcat(address _dev, uint256 _tea) pure returns(bytes)
func (_PayoutEthEval *PayoutEthEvalSession) GetConcat(_dev common.Address, _tea *big.Int) ([]byte, error) {
	return _PayoutEthEval.Contract.GetConcat(&_PayoutEthEval.CallOpts, _dev, _tea)
}

// GetConcat is a free data retrieval call binding the contract method 0xb0b4bfe8.
//
// Solidity: function getConcat(address _dev, uint256 _tea) pure returns(bytes)
func (_PayoutEthEval *PayoutEthEvalCallerSession) GetConcat(_dev common.Address, _tea *big.Int) ([]byte, error) {
	return _PayoutEthEval.Contract.GetConcat(&_PayoutEthEval.CallOpts, _dev, _tea)
}

// GetConcatHash is a free data retrieval call binding the contract method 0xe41a3766.
//
// Solidity: function getConcatHash(address _dev, uint256 _tea) pure returns(bytes32)
func (_PayoutEthEval *PayoutEthEvalCaller) GetConcatHash(opts *bind.CallOpts, _dev common.Address, _tea *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "getConcatHash", _dev, _tea)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetConcatHash is a free data retrieval call binding the contract method 0xe41a3766.
//
// Solidity: function getConcatHash(address _dev, uint256 _tea) pure returns(bytes32)
func (_PayoutEthEval *PayoutEthEvalSession) GetConcatHash(_dev common.Address, _tea *big.Int) ([32]byte, error) {
	return _PayoutEthEval.Contract.GetConcatHash(&_PayoutEthEval.CallOpts, _dev, _tea)
}

// GetConcatHash is a free data retrieval call binding the contract method 0xe41a3766.
//
// Solidity: function getConcatHash(address _dev, uint256 _tea) pure returns(bytes32)
func (_PayoutEthEval *PayoutEthEvalCallerSession) GetConcatHash(_dev common.Address, _tea *big.Int) ([32]byte, error) {
	return _PayoutEthEval.Contract.GetConcatHash(&_PayoutEthEval.CallOpts, _dev, _tea)
}

// GetTotalEarnedAmount is a free data retrieval call binding the contract method 0xe6423dcc.
//
// Solidity: function getTotalEarnedAmount(address _dev) view returns(uint256)
func (_PayoutEthEval *PayoutEthEvalCaller) GetTotalEarnedAmount(opts *bind.CallOpts, _dev common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "getTotalEarnedAmount", _dev)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalEarnedAmount is a free data retrieval call binding the contract method 0xe6423dcc.
//
// Solidity: function getTotalEarnedAmount(address _dev) view returns(uint256)
func (_PayoutEthEval *PayoutEthEvalSession) GetTotalEarnedAmount(_dev common.Address) (*big.Int, error) {
	return _PayoutEthEval.Contract.GetTotalEarnedAmount(&_PayoutEthEval.CallOpts, _dev)
}

// GetTotalEarnedAmount is a free data retrieval call binding the contract method 0xe6423dcc.
//
// Solidity: function getTotalEarnedAmount(address _dev) view returns(uint256)
func (_PayoutEthEval *PayoutEthEvalCallerSession) GetTotalEarnedAmount(_dev common.Address) (*big.Int, error) {
	return _PayoutEthEval.Contract.GetTotalEarnedAmount(&_PayoutEthEval.CallOpts, _dev)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEthEval *PayoutEthEvalCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEthEval *PayoutEthEvalSession) Owner() (common.Address, error) {
	return _PayoutEthEval.Contract.Owner(&_PayoutEthEval.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEthEval *PayoutEthEvalCallerSession) Owner() (common.Address, error) {
	return _PayoutEthEval.Contract.Owner(&_PayoutEthEval.CallOpts)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x7c4895fa.
//
// Solidity: function recoverSigner(bytes concatDevTea, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalCaller) RecoverSigner(opts *bind.CallOpts, concatDevTea []byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "recoverSigner", concatDevTea, v, r, s)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverSigner is a free data retrieval call binding the contract method 0x7c4895fa.
//
// Solidity: function recoverSigner(bytes concatDevTea, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalSession) RecoverSigner(concatDevTea []byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _PayoutEthEval.Contract.RecoverSigner(&_PayoutEthEval.CallOpts, concatDevTea, v, r, s)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x7c4895fa.
//
// Solidity: function recoverSigner(bytes concatDevTea, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalCallerSession) RecoverSigner(concatDevTea []byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _PayoutEthEval.Contract.RecoverSigner(&_PayoutEthEval.CallOpts, concatDevTea, v, r, s)
}

// RecoverSignerHash is a free data retrieval call binding the contract method 0x6abf4f14.
//
// Solidity: function recoverSignerHash(bytes32 concatDevTeaHash, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalCaller) RecoverSignerHash(opts *bind.CallOpts, concatDevTeaHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	var out []interface{}
	err := _PayoutEthEval.contract.Call(opts, &out, "recoverSignerHash", concatDevTeaHash, v, r, s)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverSignerHash is a free data retrieval call binding the contract method 0x6abf4f14.
//
// Solidity: function recoverSignerHash(bytes32 concatDevTeaHash, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalSession) RecoverSignerHash(concatDevTeaHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _PayoutEthEval.Contract.RecoverSignerHash(&_PayoutEthEval.CallOpts, concatDevTeaHash, v, r, s)
}

// RecoverSignerHash is a free data retrieval call binding the contract method 0x6abf4f14.
//
// Solidity: function recoverSignerHash(bytes32 concatDevTeaHash, uint8 v, bytes32 r, bytes32 s) pure returns(address)
func (_PayoutEthEval *PayoutEthEvalCallerSession) RecoverSignerHash(concatDevTeaHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _PayoutEthEval.Contract.RecoverSignerHash(&_PayoutEthEval.CallOpts, concatDevTeaHash, v, r, s)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) BatchPayout(opts *bind.TransactOpts, _devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "batchPayout", _devs, _teas)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEthEval *PayoutEthEvalSession) BatchPayout(_devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayout(&_PayoutEthEval.TransactOpts, _devs, _teas)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) BatchPayout(_devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayout(&_PayoutEthEval.TransactOpts, _devs, _teas)
}

// BatchPayoutServiceFeeWithPayout is a paid mutator transaction binding the contract method 0x3e019839.
//
// Solidity: function batchPayoutServiceFeeWithPayout(address[] _devs, uint256[] _teasForPayout, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) BatchPayoutServiceFeeWithPayout(opts *bind.TransactOpts, _devs []common.Address, _teasForPayout []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "batchPayoutServiceFeeWithPayout", _devs, _teasForPayout, serviceFee)
}

// BatchPayoutServiceFeeWithPayout is a paid mutator transaction binding the contract method 0x3e019839.
//
// Solidity: function batchPayoutServiceFeeWithPayout(address[] _devs, uint256[] _teasForPayout, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalSession) BatchPayoutServiceFeeWithPayout(_devs []common.Address, _teasForPayout []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayoutServiceFeeWithPayout(&_PayoutEthEval.TransactOpts, _devs, _teasForPayout, serviceFee)
}

// BatchPayoutServiceFeeWithPayout is a paid mutator transaction binding the contract method 0x3e019839.
//
// Solidity: function batchPayoutServiceFeeWithPayout(address[] _devs, uint256[] _teasForPayout, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) BatchPayoutServiceFeeWithPayout(_devs []common.Address, _teasForPayout []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayoutServiceFeeWithPayout(&_PayoutEthEval.TransactOpts, _devs, _teasForPayout, serviceFee)
}

// BatchPayoutServiceFeeWithStore is a paid mutator transaction binding the contract method 0x64e0e904.
//
// Solidity: function batchPayoutServiceFeeWithStore(address[] _devs, uint256[] _teasToStore, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) BatchPayoutServiceFeeWithStore(opts *bind.TransactOpts, _devs []common.Address, _teasToStore []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "batchPayoutServiceFeeWithStore", _devs, _teasToStore, serviceFee)
}

// BatchPayoutServiceFeeWithStore is a paid mutator transaction binding the contract method 0x64e0e904.
//
// Solidity: function batchPayoutServiceFeeWithStore(address[] _devs, uint256[] _teasToStore, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalSession) BatchPayoutServiceFeeWithStore(_devs []common.Address, _teasToStore []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayoutServiceFeeWithStore(&_PayoutEthEval.TransactOpts, _devs, _teasToStore, serviceFee)
}

// BatchPayoutServiceFeeWithStore is a paid mutator transaction binding the contract method 0x64e0e904.
//
// Solidity: function batchPayoutServiceFeeWithStore(address[] _devs, uint256[] _teasToStore, uint256 serviceFee) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) BatchPayoutServiceFeeWithStore(_devs []common.Address, _teasToStore []*big.Int, serviceFee *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.BatchPayoutServiceFeeWithStore(&_PayoutEthEval.TransactOpts, _devs, _teasToStore, serviceFee)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) ChangeOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "changeOwner", newOwner)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEthEval *PayoutEthEvalSession) ChangeOwner(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.ChangeOwner(&_PayoutEthEval.TransactOpts, newOwner)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) ChangeOwner(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.ChangeOwner(&_PayoutEthEval.TransactOpts, newOwner)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) payable returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) Deposit(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "deposit", amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) payable returns()
func (_PayoutEthEval *PayoutEthEvalSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.Deposit(&_PayoutEthEval.TransactOpts, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) payable returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.Deposit(&_PayoutEthEval.TransactOpts, amount)
}

// WithdrawHashed is a paid mutator transaction binding the contract method 0xf362a99b.
//
// Solidity: function withdrawHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) WithdrawHashed(opts *bind.TransactOpts, _dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "withdrawHashed", _dev, _tea, _v, _r, _s)
}

// WithdrawHashed is a paid mutator transaction binding the contract method 0xf362a99b.
//
// Solidity: function withdrawHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalSession) WithdrawHashed(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawHashed(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}

// WithdrawHashed is a paid mutator transaction binding the contract method 0xf362a99b.
//
// Solidity: function withdrawHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) WithdrawHashed(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawHashed(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}

// WithdrawHashedRequire is a paid mutator transaction binding the contract method 0x377c91f4.
//
// Solidity: function withdrawHashedRequire(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) WithdrawHashedRequire(opts *bind.TransactOpts, _dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "withdrawHashedRequire", _dev, _tea, _v, _r, _s)
}

// WithdrawHashedRequire is a paid mutator transaction binding the contract method 0x377c91f4.
//
// Solidity: function withdrawHashedRequire(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalSession) WithdrawHashedRequire(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawHashedRequire(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}

// WithdrawHashedRequire is a paid mutator transaction binding the contract method 0x377c91f4.
//
// Solidity: function withdrawHashedRequire(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) WithdrawHashedRequire(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawHashedRequire(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}

// WithdrawNotHashed is a paid mutator transaction binding the contract method 0x9bb442bd.
//
// Solidity: function withdrawNotHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactor) WithdrawNotHashed(opts *bind.TransactOpts, _dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.contract.Transact(opts, "withdrawNotHashed", _dev, _tea, _v, _r, _s)
}

// WithdrawNotHashed is a paid mutator transaction binding the contract method 0x9bb442bd.
//
// Solidity: function withdrawNotHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalSession) WithdrawNotHashed(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawNotHashed(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}

// WithdrawNotHashed is a paid mutator transaction binding the contract method 0x9bb442bd.
//
// Solidity: function withdrawNotHashed(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEthEval *PayoutEthEvalTransactorSession) WithdrawNotHashed(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEthEval.Contract.WithdrawNotHashed(&_PayoutEthEval.TransactOpts, _dev, _tea, _v, _r, _s)
}
