// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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
	_ = abi.ConvertType
)

// PayoutBaseMetaData contains all meta data concerning the PayoutBase contract.
var PayoutBaseMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"userId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"totalPayOut\",\"type\":\"uint256\"}],\"name\":\"getClaimableAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getContractBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"userId\",\"type\":\"bytes32\"}],\"name\":\"getPayedOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"payedOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"sendRecover\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"dev\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"userId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"totalPayOut\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// PayoutBaseABI is the input ABI used to generate the binding from.
// Deprecated: Use PayoutBaseMetaData.ABI instead.
var PayoutBaseABI = PayoutBaseMetaData.ABI

// PayoutBase is an auto generated Go binding around an Ethereum contract.
type PayoutBase struct {
	PayoutBaseCaller     // Read-only binding to the contract
	PayoutBaseTransactor // Write-only binding to the contract
	PayoutBaseFilterer   // Log filterer for contract events
}

// PayoutBaseCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayoutBaseCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutBaseTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayoutBaseTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutBaseFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayoutBaseFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayoutBaseSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayoutBaseSession struct {
	Contract     *PayoutBase       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PayoutBaseCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayoutBaseCallerSession struct {
	Contract *PayoutBaseCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// PayoutBaseTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayoutBaseTransactorSession struct {
	Contract     *PayoutBaseTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PayoutBaseRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayoutBaseRaw struct {
	Contract *PayoutBase // Generic contract binding to access the raw methods on
}

// PayoutBaseCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayoutBaseCallerRaw struct {
	Contract *PayoutBaseCaller // Generic read-only contract binding to access the raw methods on
}

// PayoutBaseTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayoutBaseTransactorRaw struct {
	Contract *PayoutBaseTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayoutBase creates a new instance of PayoutBase, bound to a specific deployed contract.
func NewPayoutBase(address common.Address, backend bind.ContractBackend) (*PayoutBase, error) {
	contract, err := bindPayoutBase(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayoutBase{PayoutBaseCaller: PayoutBaseCaller{contract: contract}, PayoutBaseTransactor: PayoutBaseTransactor{contract: contract}, PayoutBaseFilterer: PayoutBaseFilterer{contract: contract}}, nil
}

// NewPayoutBaseCaller creates a new read-only instance of PayoutBase, bound to a specific deployed contract.
func NewPayoutBaseCaller(address common.Address, caller bind.ContractCaller) (*PayoutBaseCaller, error) {
	contract, err := bindPayoutBase(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutBaseCaller{contract: contract}, nil
}

// NewPayoutBaseTransactor creates a new write-only instance of PayoutBase, bound to a specific deployed contract.
func NewPayoutBaseTransactor(address common.Address, transactor bind.ContractTransactor) (*PayoutBaseTransactor, error) {
	contract, err := bindPayoutBase(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayoutBaseTransactor{contract: contract}, nil
}

// NewPayoutBaseFilterer creates a new log filterer instance of PayoutBase, bound to a specific deployed contract.
func NewPayoutBaseFilterer(address common.Address, filterer bind.ContractFilterer) (*PayoutBaseFilterer, error) {
	contract, err := bindPayoutBase(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayoutBaseFilterer{contract: contract}, nil
}

// bindPayoutBase binds a generic wrapper to an already deployed contract.
func bindPayoutBase(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PayoutBaseMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutBase *PayoutBaseRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutBase.Contract.PayoutBaseCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutBase *PayoutBaseRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutBase.Contract.PayoutBaseTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutBase *PayoutBaseRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutBase.Contract.PayoutBaseTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayoutBase *PayoutBaseCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayoutBase.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayoutBase *PayoutBaseTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutBase.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayoutBase *PayoutBaseTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayoutBase.Contract.contract.Transact(opts, method, params...)
}

// GetClaimableAmount is a free data retrieval call binding the contract method 0xdb6e81ef.
//
// Solidity: function getClaimableAmount(bytes32 userId, uint256 totalPayOut) view returns(uint256)
func (_PayoutBase *PayoutBaseCaller) GetClaimableAmount(opts *bind.CallOpts, userId [32]byte, totalPayOut *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PayoutBase.contract.Call(opts, &out, "getClaimableAmount", userId, totalPayOut)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetClaimableAmount is a free data retrieval call binding the contract method 0xdb6e81ef.
//
// Solidity: function getClaimableAmount(bytes32 userId, uint256 totalPayOut) view returns(uint256)
func (_PayoutBase *PayoutBaseSession) GetClaimableAmount(userId [32]byte, totalPayOut *big.Int) (*big.Int, error) {
	return _PayoutBase.Contract.GetClaimableAmount(&_PayoutBase.CallOpts, userId, totalPayOut)
}

// GetClaimableAmount is a free data retrieval call binding the contract method 0xdb6e81ef.
//
// Solidity: function getClaimableAmount(bytes32 userId, uint256 totalPayOut) view returns(uint256)
func (_PayoutBase *PayoutBaseCallerSession) GetClaimableAmount(userId [32]byte, totalPayOut *big.Int) (*big.Int, error) {
	return _PayoutBase.Contract.GetClaimableAmount(&_PayoutBase.CallOpts, userId, totalPayOut)
}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_PayoutBase *PayoutBaseCaller) GetContractBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PayoutBase.contract.Call(opts, &out, "getContractBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_PayoutBase *PayoutBaseSession) GetContractBalance() (*big.Int, error) {
	return _PayoutBase.Contract.GetContractBalance(&_PayoutBase.CallOpts)
}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_PayoutBase *PayoutBaseCallerSession) GetContractBalance() (*big.Int, error) {
	return _PayoutBase.Contract.GetContractBalance(&_PayoutBase.CallOpts)
}

// GetPayedOut is a free data retrieval call binding the contract method 0x8e0fb98d.
//
// Solidity: function getPayedOut(bytes32 userId) view returns(uint256)
func (_PayoutBase *PayoutBaseCaller) GetPayedOut(opts *bind.CallOpts, userId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _PayoutBase.contract.Call(opts, &out, "getPayedOut", userId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPayedOut is a free data retrieval call binding the contract method 0x8e0fb98d.
//
// Solidity: function getPayedOut(bytes32 userId) view returns(uint256)
func (_PayoutBase *PayoutBaseSession) GetPayedOut(userId [32]byte) (*big.Int, error) {
	return _PayoutBase.Contract.GetPayedOut(&_PayoutBase.CallOpts, userId)
}

// GetPayedOut is a free data retrieval call binding the contract method 0x8e0fb98d.
//
// Solidity: function getPayedOut(bytes32 userId) view returns(uint256)
func (_PayoutBase *PayoutBaseCallerSession) GetPayedOut(userId [32]byte) (*big.Int, error) {
	return _PayoutBase.Contract.GetPayedOut(&_PayoutBase.CallOpts, userId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutBase *PayoutBaseCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PayoutBase.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutBase *PayoutBaseSession) Owner() (common.Address, error) {
	return _PayoutBase.Contract.Owner(&_PayoutBase.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PayoutBase *PayoutBaseCallerSession) Owner() (common.Address, error) {
	return _PayoutBase.Contract.Owner(&_PayoutBase.CallOpts)
}

// PayedOut is a free data retrieval call binding the contract method 0x4c293714.
//
// Solidity: function payedOut(bytes32 ) view returns(uint256)
func (_PayoutBase *PayoutBaseCaller) PayedOut(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _PayoutBase.contract.Call(opts, &out, "payedOut", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PayedOut is a free data retrieval call binding the contract method 0x4c293714.
//
// Solidity: function payedOut(bytes32 ) view returns(uint256)
func (_PayoutBase *PayoutBaseSession) PayedOut(arg0 [32]byte) (*big.Int, error) {
	return _PayoutBase.Contract.PayedOut(&_PayoutBase.CallOpts, arg0)
}

// PayedOut is a free data retrieval call binding the contract method 0x4c293714.
//
// Solidity: function payedOut(bytes32 ) view returns(uint256)
func (_PayoutBase *PayoutBaseCallerSession) PayedOut(arg0 [32]byte) (*big.Int, error) {
	return _PayoutBase.Contract.PayedOut(&_PayoutBase.CallOpts, arg0)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PayoutBase *PayoutBaseTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayoutBase.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PayoutBase *PayoutBaseSession) RenounceOwnership() (*types.Transaction, error) {
	return _PayoutBase.Contract.RenounceOwnership(&_PayoutBase.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PayoutBase *PayoutBaseTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _PayoutBase.Contract.RenounceOwnership(&_PayoutBase.TransactOpts)
}

// SendRecover is a paid mutator transaction binding the contract method 0xd2fc98ea.
//
// Solidity: function sendRecover(address receiver, uint256 amount) returns()
func (_PayoutBase *PayoutBaseTransactor) SendRecover(opts *bind.TransactOpts, receiver common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PayoutBase.contract.Transact(opts, "sendRecover", receiver, amount)
}

// SendRecover is a paid mutator transaction binding the contract method 0xd2fc98ea.
//
// Solidity: function sendRecover(address receiver, uint256 amount) returns()
func (_PayoutBase *PayoutBaseSession) SendRecover(receiver common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PayoutBase.Contract.SendRecover(&_PayoutBase.TransactOpts, receiver, amount)
}

// SendRecover is a paid mutator transaction binding the contract method 0xd2fc98ea.
//
// Solidity: function sendRecover(address receiver, uint256 amount) returns()
func (_PayoutBase *PayoutBaseTransactorSession) SendRecover(receiver common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PayoutBase.Contract.SendRecover(&_PayoutBase.TransactOpts, receiver, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PayoutBase *PayoutBaseTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PayoutBase.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PayoutBase *PayoutBaseSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutBase.Contract.TransferOwnership(&_PayoutBase.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PayoutBase *PayoutBaseTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PayoutBase.Contract.TransferOwnership(&_PayoutBase.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x71676bd6.
//
// Solidity: function withdraw(address dev, bytes32 userId, uint256 totalPayOut, uint8 v, bytes32 r, bytes32 s) returns()
func (_PayoutBase *PayoutBaseTransactor) Withdraw(opts *bind.TransactOpts, dev common.Address, userId [32]byte, totalPayOut *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _PayoutBase.contract.Transact(opts, "withdraw", dev, userId, totalPayOut, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x71676bd6.
//
// Solidity: function withdraw(address dev, bytes32 userId, uint256 totalPayOut, uint8 v, bytes32 r, bytes32 s) returns()
func (_PayoutBase *PayoutBaseSession) Withdraw(dev common.Address, userId [32]byte, totalPayOut *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _PayoutBase.Contract.Withdraw(&_PayoutBase.TransactOpts, dev, userId, totalPayOut, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x71676bd6.
//
// Solidity: function withdraw(address dev, bytes32 userId, uint256 totalPayOut, uint8 v, bytes32 r, bytes32 s) returns()
func (_PayoutBase *PayoutBaseTransactorSession) Withdraw(dev common.Address, userId [32]byte, totalPayOut *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _PayoutBase.Contract.Withdraw(&_PayoutBase.TransactOpts, dev, userId, totalPayOut, v, r, s)
}

// PayoutBaseInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the PayoutBase contract.
type PayoutBaseInitializedIterator struct {
	Event *PayoutBaseInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PayoutBaseInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PayoutBaseInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PayoutBaseInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PayoutBaseInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PayoutBaseInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PayoutBaseInitialized represents a Initialized event raised by the PayoutBase contract.
type PayoutBaseInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_PayoutBase *PayoutBaseFilterer) FilterInitialized(opts *bind.FilterOpts) (*PayoutBaseInitializedIterator, error) {

	logs, sub, err := _PayoutBase.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PayoutBaseInitializedIterator{contract: _PayoutBase.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_PayoutBase *PayoutBaseFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PayoutBaseInitialized) (event.Subscription, error) {

	logs, sub, err := _PayoutBase.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PayoutBaseInitialized)
				if err := _PayoutBase.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_PayoutBase *PayoutBaseFilterer) ParseInitialized(log types.Log) (*PayoutBaseInitialized, error) {
	event := new(PayoutBaseInitialized)
	if err := _PayoutBase.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PayoutBaseOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the PayoutBase contract.
type PayoutBaseOwnershipTransferredIterator struct {
	Event *PayoutBaseOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PayoutBaseOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PayoutBaseOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PayoutBaseOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PayoutBaseOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PayoutBaseOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PayoutBaseOwnershipTransferred represents a OwnershipTransferred event raised by the PayoutBase contract.
type PayoutBaseOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PayoutBase *PayoutBaseFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PayoutBaseOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PayoutBase.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PayoutBaseOwnershipTransferredIterator{contract: _PayoutBase.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PayoutBase *PayoutBaseFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PayoutBaseOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PayoutBase.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PayoutBaseOwnershipTransferred)
				if err := _PayoutBase.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PayoutBase *PayoutBaseFilterer) ParseOwnershipTransferred(log types.Log) (*PayoutBaseOwnershipTransferred, error) {
	event := new(PayoutBaseOwnershipTransferred)
	if err := _PayoutBase.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
