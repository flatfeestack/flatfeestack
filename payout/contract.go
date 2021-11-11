// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
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

// PayoutEthMetaData contains all meta data concerning the PayoutEth contract.
var PayoutEthMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"addresspayable[]\",\"name\":\"_devs\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_teas\",\"type\":\"uint256[]\"}],\"name\":\"batchPayout\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"changeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dev\",\"type\":\"address\"}],\"name\":\"getTea\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"oldTea\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newTea\",\"type\":\"uint256\"}],\"name\":\"setTea\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_devs\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"oldTeas\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newTeas\",\"type\":\"uint256[]\"}],\"name\":\"setTeas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"teaMap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_dev\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tea\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"d91966a7": "batchPayout(address[],uint256[])",
		"a6f9dae1": "changeOwner(address)",
		"eb3b46a4": "getTea(address)",
		"8da5cb5b": "owner()",
		"c78b6f4b": "setTea(address,uint256,uint256)",
		"25aafce8": "setTeas(address[],uint256[],uint256[])",
		"b2cb37f0": "teaMap(address)",
		"7078002f": "withdraw(address,uint256,uint8,bytes32,bytes32)",
	},
	Bin: "0x608060405234801561001057600080fd5b50600180546001600160a01b03191633179055610c6b806100326000396000f3fe60806040526004361061007f5760003560e01c8063b2cb37f01161004e578063b2cb37f01461012a578063c78b6f4b14610165578063d91966a714610185578063eb3b46a4146101a557600080fd5b806325aafce81461008b5780637078002f146100ad5780638da5cb5b146100cd578063a6f9dae11461010a57600080fd5b3661008657005b600080fd5b34801561009757600080fd5b506100ab6100a63660046108b1565b6101db565b005b3480156100b957600080fd5b506100ab6100c8366004610963565b610338565b3480156100d957600080fd5b506001546100ed906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b34801561011657600080fd5b506100ab6101253660046109bb565b610568565b34801561013657600080fd5b506101576101453660046109bb565b60006020819052908152604090205481565b604051908152602001610101565b34801561017157600080fd5b506100ab6101803660046109df565b6105b4565b34801561019157600080fd5b506100ab6101a0366004610aea565b6106fe565b3480156101b157600080fd5b506101576101c03660046109bb565b6001600160a01b031660009081526020819052604090205490565b6001546001600160a01b0316331461020e5760405162461bcd60e51b815260040161020590610bac565b60405180910390fd5b8481146102675760405162461bcd60e51b815260206004820152602160248201527f506172616d6574657273206d75737420686176652073616d65206c656e6774686044820152601760f91b6064820152608401610205565b60005b8581101561032f57600087878381811061028657610286610bd7565b905060200201602081019061029b91906109bb565b6001600160a01b0381166000908152602081905260408120549192508585858181106102c9576102c9610bd7565b905060200201359050818888868181106102e5576102e5610bd7565b905060200201351480156102f857508181115b15610319576001600160a01b03831660009081526020819052604090208190555b505050808061032790610c03565b91505061026a565b50505050505050565b6001600160a01b03851660009081526020819052604090205484116103b05760405162461bcd60e51b815260206004820152602860248201527f54686573652066756e6473206861766520616c7265616479206265656e2077696044820152673a34323930bbb71760c11b6064820152608401610205565b600180546040516bffffffffffffffffffffffff19606089901b166020820152603481018790526001600160a01b03909116919060540160408051601f198184030181529082905280516020918201207f19457468657265756d205369676e6564204d6573736167653a0a36360000000091830191909152603c820152605c0160408051601f198184030181528282528051602091820120600084529083018083525260ff871690820152606081018590526080810184905260a0016020604051602081039080840390855afa15801561048e573d6000803e3d6000fd5b505050602060405103516001600160a01b0316146105145760405162461bcd60e51b815260206004820152603760248201527f5369676e617475726520646f6573206e6f74206d61746368206f776e6572206160448201527f6e642070726f766964656420706172616d65746572732e0000000000000000006064820152608401610205565b6001600160a01b0385166000818152602081905260409020805490869055906108fc6105408388610c1e565b6040518115909202916000818181858888f1935050505015801561032f573d6000803e3d6000fd5b6001546001600160a01b031633146105925760405162461bcd60e51b815260040161020590610bac565b600180546001600160a01b0319166001600160a01b0392909216919091179055565b6001546001600160a01b031633146105de5760405162461bcd60e51b815260040161020590610bac565b6001600160a01b038316600090815260208190526040902054821461065d5760405162461bcd60e51b815260206004820152602f60248201527f53746f72656420746561206973206e6f7420657175616c20746f20746865207060448201526e3937bb34b232b21037b6322a32b09760891b6064820152608401610205565b6001600160a01b03831660009081526020819052604090205481116106de5760405162461bcd60e51b815260206004820152603160248201527f43616e6e6f74207365742061206c6f7765722076616c75652064756520746f2060448201527039b2b1bab934ba3c903932b0b9b7b7399760791b6064820152608401610205565b6001600160a01b0390921660009081526020819052604090209190915550565b6001546001600160a01b031633146107285760405162461bcd60e51b815260040161020590610bac565b80518251146107795760405162461bcd60e51b815260206004820152601d60248201527f417272617973206d75737420686176652073616d65206c656e6774682e0000006044820152606401610205565b60005b825181101561086057600083828151811061079957610799610bd7565b602002602001015190506000806000836001600160a01b03166001600160a01b0316815260200190815260200160002054905060008484815181106107e0576107e0610bd7565b602002602001015190508181116107f95750505061084e565b6001600160a01b03831660008181526020819052604090208290556108fc6108218484610c1e565b6040518115909202916000818181858888f19350505050158015610849573d6000803e3d6000fd5b505050505b8061085881610c03565b91505061077c565b505050565b60008083601f84011261087757600080fd5b50813567ffffffffffffffff81111561088f57600080fd5b6020830191508360208260051b85010111156108aa57600080fd5b9250929050565b600080600080600080606087890312156108ca57600080fd5b863567ffffffffffffffff808211156108e257600080fd5b6108ee8a838b01610865565b9098509650602089013591508082111561090757600080fd5b6109138a838b01610865565b9096509450604089013591508082111561092c57600080fd5b5061093989828a01610865565b979a9699509497509295939492505050565b6001600160a01b038116811461096057600080fd5b50565b600080600080600060a0868803121561097b57600080fd5b85356109868161094b565b945060208601359350604086013560ff811681146109a357600080fd5b94979396509394606081013594506080013592915050565b6000602082840312156109cd57600080fd5b81356109d88161094b565b9392505050565b6000806000606084860312156109f457600080fd5b83356109ff8161094b565b95602085013595506040909401359392505050565b634e487b7160e01b600052604160045260246000fd5b604051601f8201601f1916810167ffffffffffffffff81118282101715610a5357610a53610a14565b604052919050565b600067ffffffffffffffff821115610a7557610a75610a14565b5060051b60200190565b600082601f830112610a9057600080fd5b81356020610aa5610aa083610a5b565b610a2a565b82815260059290921b84018101918181019086841115610ac457600080fd5b8286015b84811015610adf5780358352918301918301610ac8565b509695505050505050565b60008060408385031215610afd57600080fd5b823567ffffffffffffffff80821115610b1557600080fd5b818501915085601f830112610b2957600080fd5b81356020610b39610aa083610a5b565b82815260059290921b84018101918181019089841115610b5857600080fd5b948201945b83861015610b7f578535610b708161094b565b82529482019490820190610b5d565b96505086013592505080821115610b9557600080fd5b50610ba285828601610a7f565b9150509250929050565b60208082526011908201527027379030baba3437b934bd30ba34b7b71760791b604082015260600190565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601160045260246000fd5b6000600019821415610c1757610c17610bed565b5060010190565b600082821015610c3057610c30610bed565b50039056fea264697066735822122088996e827c5dfdde5dcecb6c92e7585fbb01e32b763f3b4a2a5e4421d4d9a91d64736f6c63430008090033",
}

// PayoutEthABI is the input ABI used to generate the binding from.
// Deprecated: Use PayoutEthMetaData.ABI instead.
var PayoutEthABI = PayoutEthMetaData.ABI

// Deprecated: Use PayoutEthMetaData.Sigs instead.
// PayoutEthFuncSigs maps the 4-byte function signature to its string representation.
var PayoutEthFuncSigs = PayoutEthMetaData.Sigs

// PayoutEthBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PayoutEthMetaData.Bin instead.
var PayoutEthBin = PayoutEthMetaData.Bin

// DeployPayoutEth deploys a new Ethereum contract, binding an instance of PayoutEth to it.
func DeployPayoutEth(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PayoutEth, error) {
	parsed, err := PayoutEthMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PayoutEthBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PayoutEth{PayoutEthCaller: PayoutEthCaller{contract: contract}, PayoutEthTransactor: PayoutEthTransactor{contract: contract}, PayoutEthFilterer: PayoutEthFilterer{contract: contract}}, nil
}

// PayoutEth is an auto generated Go binding around an Ethereum contract.
type PayoutEth struct {
	PayoutEthCaller     // Read-only binding to the contract
	PayoutEthTransactor // Write-only binding to the contract
	PayoutEthFilterer   // Log filterer for contract events
}

// PayoutEthCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayoutEthCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayoutEthTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayoutEthFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutEthSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayoutEthSession struct {
	Contract     *PayoutEth        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PayoutEthCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayoutEthCallerSession struct {
	Contract *PayoutEthCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// PayoutEthTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayoutEthTransactorSession struct {
	Contract     *PayoutEthTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// PayoutEthRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayoutEthRaw struct {
	Contract *PayoutEth // Generic contract binding to access the raw methods on
}

// PayoutEthCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayoutEthCallerRaw struct {
	Contract *PayoutEthCaller // Generic read-only contract binding to access the raw methods on
}

// PayoutEthTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayoutEthTransactorRaw struct {
	Contract *PayoutEthTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayoutEth creates a new instance of PayoutEth, bound to a specific deployed contract.
func NewPayoutEth(address common.Address, backend bind.ContractBackend) (*PayoutEth, error) {
	contract, err := bindPayoutEth(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayoutEth{PayoutEthCaller: PayoutEthCaller{contract: contract}, PayoutEthTransactor: PayoutEthTransactor{contract: contract}, PayoutEthFilterer: PayoutEthFilterer{contract: contract}}, nil
}

// NewPayoutEthCaller creates a new read-only instance of PayoutEth, bound to a specific deployed contract.
func NewPayoutEthCaller(address common.Address, caller bind.ContractCaller) (*PayoutEthCaller, error) {
	contract, err := bindPayoutEth(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutEthCaller{contract: contract}, nil
}

// NewPayoutEthTransactor creates a new write-only instance of PayoutEth, bound to a specific deployed contract.
func NewPayoutEthTransactor(address common.Address, transactor bind.ContractTransactor) (*PayoutEthTransactor, error) {
	contract, err := bindPayoutEth(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutEthTransactor{contract: contract}, nil
}

// NewPayoutEthFilterer creates a new log filterer instance of PayoutEth, bound to a specific deployed contract.
func NewPayoutEthFilterer(address common.Address, filterer bind.ContractFilterer) (*PayoutEthFilterer, error) {
	contract, err := bindPayoutEth(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayoutEthFilterer{contract: contract}, nil
}

// bindPayoutEth binds a generic wrapper to an already deployed contract.
func bindPayoutEth(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PayoutEthABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutEth *PayoutEthRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutEth.Contract.PayoutEthCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutEth *PayoutEthRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutEth.Contract.PayoutEthTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutEth *PayoutEthRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutEth.Contract.PayoutEthTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutEth *PayoutEthCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutEth.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutEth *PayoutEthTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutEth.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutEth *PayoutEthTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutEth.Contract.contract.Transact(opts, method, params...)
}

// GetTea is a free data retrieval call binding the contract method 0xeb3b46a4.
//
// Solidity: function getTea(address _dev) view returns(uint256)
func (_PayoutEth *PayoutEthCaller) GetTea(opts *bind.CallOpts, _dev common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PayoutEth.contract.Call(opts, &out, "getTea", _dev)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTea is a free data retrieval call binding the contract method 0xeb3b46a4.
//
// Solidity: function getTea(address _dev) view returns(uint256)
func (_PayoutEth *PayoutEthSession) GetTea(_dev common.Address) (*big.Int, error) {
	return _PayoutEth.Contract.GetTea(&_PayoutEth.CallOpts, _dev)
}

// GetTea is a free data retrieval call binding the contract method 0xeb3b46a4.
//
// Solidity: function getTea(address _dev) view returns(uint256)
func (_PayoutEth *PayoutEthCallerSession) GetTea(_dev common.Address) (*big.Int, error) {
	return _PayoutEth.Contract.GetTea(&_PayoutEth.CallOpts, _dev)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEth *PayoutEthCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PayoutEth.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEth *PayoutEthSession) Owner() (common.Address, error) {
	return _PayoutEth.Contract.Owner(&_PayoutEth.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutEth *PayoutEthCallerSession) Owner() (common.Address, error) {
	return _PayoutEth.Contract.Owner(&_PayoutEth.CallOpts)
}

// TeaMap is a free data retrieval call binding the contract method 0xb2cb37f0.
//
// Solidity: function teaMap(address ) view returns(uint256)
func (_PayoutEth *PayoutEthCaller) TeaMap(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PayoutEth.contract.Call(opts, &out, "teaMap", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TeaMap is a free data retrieval call binding the contract method 0xb2cb37f0.
//
// Solidity: function teaMap(address ) view returns(uint256)
func (_PayoutEth *PayoutEthSession) TeaMap(arg0 common.Address) (*big.Int, error) {
	return _PayoutEth.Contract.TeaMap(&_PayoutEth.CallOpts, arg0)
}

// TeaMap is a free data retrieval call binding the contract method 0xb2cb37f0.
//
// Solidity: function teaMap(address ) view returns(uint256)
func (_PayoutEth *PayoutEthCallerSession) TeaMap(arg0 common.Address) (*big.Int, error) {
	return _PayoutEth.Contract.TeaMap(&_PayoutEth.CallOpts, arg0)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEth *PayoutEthTransactor) BatchPayout(opts *bind.TransactOpts, _devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.contract.Transact(opts, "batchPayout", _devs, _teas)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEth *PayoutEthSession) BatchPayout(_devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.BatchPayout(&_PayoutEth.TransactOpts, _devs, _teas)
}

// BatchPayout is a paid mutator transaction binding the contract method 0xd91966a7.
//
// Solidity: function batchPayout(address[] _devs, uint256[] _teas) returns()
func (_PayoutEth *PayoutEthTransactorSession) BatchPayout(_devs []common.Address, _teas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.BatchPayout(&_PayoutEth.TransactOpts, _devs, _teas)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEth *PayoutEthTransactor) ChangeOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEth.contract.Transact(opts, "changeOwner", newOwner)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEth *PayoutEthSession) ChangeOwner(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEth.Contract.ChangeOwner(&_PayoutEth.TransactOpts, newOwner)
}

// ChangeOwner is a paid mutator transaction binding the contract method 0xa6f9dae1.
//
// Solidity: function changeOwner(address newOwner) returns()
func (_PayoutEth *PayoutEthTransactorSession) ChangeOwner(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutEth.Contract.ChangeOwner(&_PayoutEth.TransactOpts, newOwner)
}

// SetTea is a paid mutator transaction binding the contract method 0xc78b6f4b.
//
// Solidity: function setTea(address _dev, uint256 oldTea, uint256 newTea) returns()
func (_PayoutEth *PayoutEthTransactor) SetTea(opts *bind.TransactOpts, _dev common.Address, oldTea *big.Int, newTea *big.Int) (*types.Transaction, error) {
	return _PayoutEth.contract.Transact(opts, "setTea", _dev, oldTea, newTea)
}

// SetTea is a paid mutator transaction binding the contract method 0xc78b6f4b.
//
// Solidity: function setTea(address _dev, uint256 oldTea, uint256 newTea) returns()
func (_PayoutEth *PayoutEthSession) SetTea(_dev common.Address, oldTea *big.Int, newTea *big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.SetTea(&_PayoutEth.TransactOpts, _dev, oldTea, newTea)
}

// SetTea is a paid mutator transaction binding the contract method 0xc78b6f4b.
//
// Solidity: function setTea(address _dev, uint256 oldTea, uint256 newTea) returns()
func (_PayoutEth *PayoutEthTransactorSession) SetTea(_dev common.Address, oldTea *big.Int, newTea *big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.SetTea(&_PayoutEth.TransactOpts, _dev, oldTea, newTea)
}

// SetTeas is a paid mutator transaction binding the contract method 0x25aafce8.
//
// Solidity: function setTeas(address[] _devs, uint256[] oldTeas, uint256[] newTeas) returns()
func (_PayoutEth *PayoutEthTransactor) SetTeas(opts *bind.TransactOpts, _devs []common.Address, oldTeas []*big.Int, newTeas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.contract.Transact(opts, "setTeas", _devs, oldTeas, newTeas)
}

// SetTeas is a paid mutator transaction binding the contract method 0x25aafce8.
//
// Solidity: function setTeas(address[] _devs, uint256[] oldTeas, uint256[] newTeas) returns()
func (_PayoutEth *PayoutEthSession) SetTeas(_devs []common.Address, oldTeas []*big.Int, newTeas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.SetTeas(&_PayoutEth.TransactOpts, _devs, oldTeas, newTeas)
}

// SetTeas is a paid mutator transaction binding the contract method 0x25aafce8.
//
// Solidity: function setTeas(address[] _devs, uint256[] oldTeas, uint256[] newTeas) returns()
func (_PayoutEth *PayoutEthTransactorSession) SetTeas(_devs []common.Address, oldTeas []*big.Int, newTeas []*big.Int) (*types.Transaction, error) {
	return _PayoutEth.Contract.SetTeas(&_PayoutEth.TransactOpts, _devs, oldTeas, newTeas)
}

// Withdraw is a paid mutator transaction binding the contract method 0x7078002f.
//
// Solidity: function withdraw(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEth *PayoutEthTransactor) Withdraw(opts *bind.TransactOpts, _dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEth.contract.Transact(opts, "withdraw", _dev, _tea, _v, _r, _s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x7078002f.
//
// Solidity: function withdraw(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEth *PayoutEthSession) Withdraw(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEth.Contract.Withdraw(&_PayoutEth.TransactOpts, _dev, _tea, _v, _r, _s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x7078002f.
//
// Solidity: function withdraw(address _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_PayoutEth *PayoutEthTransactorSession) Withdraw(_dev common.Address, _tea *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _PayoutEth.Contract.Withdraw(&_PayoutEth.TransactOpts, _dev, _tea, _v, _r, _s)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayoutEth *PayoutEthTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutEth.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayoutEth *PayoutEthSession) Receive() (*types.Transaction, error) {
	return _PayoutEth.Contract.Receive(&_PayoutEth.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayoutEth *PayoutEthTransactorSession) Receive() (*types.Transaction, error) {
	return _PayoutEth.Contract.Receive(&_PayoutEth.TransactOpts)
}
