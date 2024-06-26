// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// RowDataBodies is an auto generated low-level Go binding around an user-defined struct.
type RowDataBodies struct {
	X  int32
	Y  int32
	R  uint32
	Vx int32
	Vy int32
	Ax int32
	Ay int32
}

// RowDataMeta is an auto generated low-level Go binding around an user-defined struct.
type RowDataMeta struct {
	MaxBodyCount uint8
	BodyCount    uint8
}

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getBodiesRow\",\"inputs\":[{\"name\":\"bodyId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structRowData_Bodies\",\"components\":[{\"name\":\"x\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"y\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"r\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"vx\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"vy\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"ax\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"ay\",\"type\":\"int32\",\"internalType\":\"int32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMetaRow\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structRowData_Meta\",\"components\":[{\"name\":\"maxBodyCount\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"bodyCount\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]}],\"stateMutability\":\"view\"}]",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// GetBodiesRow is a free data retrieval call binding the contract method 0x9cf3a3a9.
//
// Solidity: function getBodiesRow(uint8 bodyId) view returns((int32,int32,uint32,int32,int32,int32,int32))
func (_Contract *ContractCaller) GetBodiesRow(opts *bind.CallOpts, bodyId uint8) (RowDataBodies, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getBodiesRow", bodyId)

	if err != nil {
		return *new(RowDataBodies), err
	}

	out0 := *abi.ConvertType(out[0], new(RowDataBodies)).(*RowDataBodies)

	return out0, err

}

// GetBodiesRow is a free data retrieval call binding the contract method 0x9cf3a3a9.
//
// Solidity: function getBodiesRow(uint8 bodyId) view returns((int32,int32,uint32,int32,int32,int32,int32))
func (_Contract *ContractSession) GetBodiesRow(bodyId uint8) (RowDataBodies, error) {
	return _Contract.Contract.GetBodiesRow(&_Contract.CallOpts, bodyId)
}

// GetBodiesRow is a free data retrieval call binding the contract method 0x9cf3a3a9.
//
// Solidity: function getBodiesRow(uint8 bodyId) view returns((int32,int32,uint32,int32,int32,int32,int32))
func (_Contract *ContractCallerSession) GetBodiesRow(bodyId uint8) (RowDataBodies, error) {
	return _Contract.Contract.GetBodiesRow(&_Contract.CallOpts, bodyId)
}

// GetMetaRow is a free data retrieval call binding the contract method 0x422f7e1d.
//
// Solidity: function getMetaRow() view returns((uint8,uint8))
func (_Contract *ContractCaller) GetMetaRow(opts *bind.CallOpts) (RowDataMeta, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getMetaRow")

	if err != nil {
		return *new(RowDataMeta), err
	}

	out0 := *abi.ConvertType(out[0], new(RowDataMeta)).(*RowDataMeta)

	return out0, err

}

// GetMetaRow is a free data retrieval call binding the contract method 0x422f7e1d.
//
// Solidity: function getMetaRow() view returns((uint8,uint8))
func (_Contract *ContractSession) GetMetaRow() (RowDataMeta, error) {
	return _Contract.Contract.GetMetaRow(&_Contract.CallOpts)
}

// GetMetaRow is a free data retrieval call binding the contract method 0x422f7e1d.
//
// Solidity: function getMetaRow() view returns((uint8,uint8))
func (_Contract *ContractCallerSession) GetMetaRow() (RowDataMeta, error) {
	return _Contract.Contract.GetMetaRow(&_Contract.CallOpts)
}
