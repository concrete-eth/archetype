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

// RowDataBoard is an auto generated low-level Go binding around an user-defined struct.
type RowDataBoard struct {
	PlayerId uint16
}

// RowDataMeta is an auto generated low-level Go binding around an user-defined struct.
type RowDataMeta struct {
	StartBlock  uint64
	MaxPlayers  uint16
	PlayerCount uint16
}

// RowDataPlayers is an auto generated low-level Go binding around an user-defined struct.
type RowDataPlayers struct {
	X int16
	Y int16
}

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getBoard\",\"inputs\":[{\"name\":\"x\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"y\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structRowData_Board\",\"components\":[{\"name\":\"playerId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMeta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structRowData_Meta\",\"components\":[{\"name\":\"startBlock\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"maxPlayers\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"playerCount\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPlayers\",\"inputs\":[{\"name\":\"playerId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structRowData_Players\",\"components\":[{\"name\":\"x\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"y\",\"type\":\"int16\",\"internalType\":\"int16\"}]}],\"stateMutability\":\"view\"}]",
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

// GetBoard is a free data retrieval call binding the contract method 0x5f65f666.
//
// Solidity: function getBoard(int16 x, int16 y) view returns((uint16))
func (_Contract *ContractCaller) GetBoard(opts *bind.CallOpts, x int16, y int16) (RowDataBoard, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getBoard", x, y)

	if err != nil {
		return *new(RowDataBoard), err
	}

	out0 := *abi.ConvertType(out[0], new(RowDataBoard)).(*RowDataBoard)

	return out0, err

}

// GetBoard is a free data retrieval call binding the contract method 0x5f65f666.
//
// Solidity: function getBoard(int16 x, int16 y) view returns((uint16))
func (_Contract *ContractSession) GetBoard(x int16, y int16) (RowDataBoard, error) {
	return _Contract.Contract.GetBoard(&_Contract.CallOpts, x, y)
}

// GetBoard is a free data retrieval call binding the contract method 0x5f65f666.
//
// Solidity: function getBoard(int16 x, int16 y) view returns((uint16))
func (_Contract *ContractCallerSession) GetBoard(x int16, y int16) (RowDataBoard, error) {
	return _Contract.Contract.GetBoard(&_Contract.CallOpts, x, y)
}

// GetMeta is a free data retrieval call binding the contract method 0xa79af2ce.
//
// Solidity: function getMeta() view returns((uint64,uint16,uint16))
func (_Contract *ContractCaller) GetMeta(opts *bind.CallOpts) (RowDataMeta, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getMeta")

	if err != nil {
		return *new(RowDataMeta), err
	}

	out0 := *abi.ConvertType(out[0], new(RowDataMeta)).(*RowDataMeta)

	return out0, err

}

// GetMeta is a free data retrieval call binding the contract method 0xa79af2ce.
//
// Solidity: function getMeta() view returns((uint64,uint16,uint16))
func (_Contract *ContractSession) GetMeta() (RowDataMeta, error) {
	return _Contract.Contract.GetMeta(&_Contract.CallOpts)
}

// GetMeta is a free data retrieval call binding the contract method 0xa79af2ce.
//
// Solidity: function getMeta() view returns((uint64,uint16,uint16))
func (_Contract *ContractCallerSession) GetMeta() (RowDataMeta, error) {
	return _Contract.Contract.GetMeta(&_Contract.CallOpts)
}

// GetPlayers is a free data retrieval call binding the contract method 0xfa3ffbee.
//
// Solidity: function getPlayers(uint16 playerId) view returns((int16,int16))
func (_Contract *ContractCaller) GetPlayers(opts *bind.CallOpts, playerId uint16) (RowDataPlayers, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getPlayers", playerId)

	if err != nil {
		return *new(RowDataPlayers), err
	}

	out0 := *abi.ConvertType(out[0], new(RowDataPlayers)).(*RowDataPlayers)

	return out0, err

}

// GetPlayers is a free data retrieval call binding the contract method 0xfa3ffbee.
//
// Solidity: function getPlayers(uint16 playerId) view returns((int16,int16))
func (_Contract *ContractSession) GetPlayers(playerId uint16) (RowDataPlayers, error) {
	return _Contract.Contract.GetPlayers(&_Contract.CallOpts, playerId)
}

// GetPlayers is a free data retrieval call binding the contract method 0xfa3ffbee.
//
// Solidity: function getPlayers(uint16 playerId) view returns((int16,int16))
func (_Contract *ContractCallerSession) GetPlayers(playerId uint16) (RowDataPlayers, error) {
	return _Contract.Contract.GetPlayers(&_Contract.CallOpts, playerId)
}
