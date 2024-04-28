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

// ActionDataAddBody is an auto generated low-level Go binding around an user-defined struct.
type ActionDataAddBody struct {
	X  int32
	Y  int32
	R  uint32
	Vx int32
	Vy int32
}

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"fallback\",\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addBody\",\"inputs\":[{\"name\":\"action\",\"type\":\"tuple\",\"internalType\":\"structActionData_AddBody\",\"components\":[{\"name\":\"x\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"y\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"r\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"vx\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"vy\",\"type\":\"int32\",\"internalType\":\"int32\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executeMultipleActions\",\"inputs\":[{\"name\":\"actionIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"actionCount\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"},{\"name\":\"actionData\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_logic\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"lastTickBlockNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proxy\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tick\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ActionExecuted\",\"inputs\":[{\"name\":\"actionId\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]}]",
	Bin: "0x608060405234801561001057600080fd5b5061125f806100206000396000f3fe60806040523480156200001157600080fd5b50600436106200006a5760003560e01c806322c5eafe14620000765780633eaf5d9f146200008d578063c4d66de81462000097578063d0b3617114620000ae578063ec55688914620000c5578063ff28019814620000f6575b620000746200010f565b005b62000074620000873660046200084d565b62000187565b62000074620001f0565b62000074620000a8366004620008d4565b62000296565b62000074620000bf36600462000aa9565b62000411565b600054620000d9906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6200010060015481565b604051908152602001620000ed565b6000546001600160a01b03166200016d5760405162461bcd60e51b815260206004820152601d60248201527f4172636850726f787941646d696e3a2070726f7879206e6f742073657400000060448201526064015b60405180910390fd5b60005462000184906001600160a01b0316620004c3565b50565b600054604051631162f57f60e11b81526001600160a01b03909116906322c5eafe90620001b990849060040162000ba0565b600060405180830381600087803b158015620001d457600080fd5b505af1158015620001e9573d6000803e3d6000fd5b5050505050565b6001544311620002395760405162461bcd60e51b815260206004820152601360248201527223b0b6b29d103a34b1b5903a37b79039b7b7b760691b604482015260640162000164565b6000805460408051633eaf5d9f60e01b815290516001600160a01b0390921692633eaf5d9f9260048084019382900301818387803b1580156200027b57600080fd5b505af115801562000290573d6000803e3d6000fd5b50505050565b7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a008054600160401b810460ff16159067ffffffffffffffff16600081158015620002dd5750825b905060008267ffffffffffffffff166001148015620002fb5750303b155b9050811580156200030a575080155b15620003295760405163f92ee8a960e01b815260040160405180910390fd5b845467ffffffffffffffff1916600117855583156200035457845460ff60401b1916600160401b1785555b600030876040516200036690620007a6565b6001600160a01b03928316815291166020820152606060408201819052600090820152608001604051809103906000f080158015620003a9573d6000803e3d6000fd5b509050620003b781620004ea565b620003c1620005d7565b5083156200040957845460ff60401b19168555604051600181527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b505050505050565b6000805b8451811015620001e957600084828151811062000436576200043662000bec565b602002602001015160ff16905060008390505b62000455828562000c18565b811015620004aa57620004a187848151811062000476576200047662000bec565b602002602001015186838151811062000493576200049362000bec565b602002602001015162000663565b60010162000449565b50620004b7818462000c18565b92505060010162000415565b60603660008037600080366000855afa3d6000803e808015620004e5573d6000f35b3d6000fd5b6001600160a01b038116620005505760405162461bcd60e51b815260206004820152602560248201527f4172636850726f787941646d696e3a20696e76616c69642070726f7879206164604482015264647265737360d81b606482015260840162000164565b6000546001600160a01b031615620005b55760405162461bcd60e51b815260206004820152602160248201527f4172636850726f787941646d696e3a2070726f787920616c72656164792073656044820152601d60fa1b606482015260840162000164565b600080546001600160a01b0319166001600160a01b0392909216919091179055565b620005f5600080620005ec6064600662000c34565b60008062000708565b6200062f620006086064603b1962000c34565b6000620006186064600262000c34565b600062000629606460031962000c34565b62000708565b62000661620006416064603c62000c34565b6000620006516064600262000c34565b6000620006296064600462000c34565b565b8163ffffffff16633eaf5d9f0362000683576200067f620001f0565b5050565b8163ffffffff166322c5eafe03620006bf57600081806020019051810190620006ad919062000c5e565b9050620006ba8162000187565b505050565b60405162461bcd60e51b815260206004820152601d60248201527f456e747279706f696e743a20496e76616c696420616374696f6e204944000000604482015260640162000164565b6000546040805160a081018252600388810b825287810b602083015263ffffffff87168284015285810b606083015284900b60808201529051631162f57f60e11b81526001600160a01b03909216916322c5eafe916200076b9160040162000ba0565b600060405180830381600087803b1580156200078657600080fd5b505af11580156200079b573d6000803e3d6000fd5b505050505050505050565b6105508062000cda83390190565b634e487b7160e01b600052604160045260246000fd5b60405160a0810167ffffffffffffffff81118282101715620007f057620007f0620007b4565b60405290565b604051601f8201601f1916810167ffffffffffffffff81118282101715620008225762000822620007b4565b604052919050565b8060030b81146200018457600080fd5b63ffffffff811681146200018457600080fd5b600060a082840312156200086057600080fd5b6200086a620007ca565b823562000877816200082a565b8152602083013562000889816200082a565b602082015260408301356200089e816200083a565b60408201526060830135620008b3816200082a565b60608201526080830135620008c8816200082a565b60808201529392505050565b600060208284031215620008e757600080fd5b81356001600160a01b0381168114620008ff57600080fd5b9392505050565b600067ffffffffffffffff821115620009235762000923620007b4565b5060051b60200190565b600082601f8301126200093f57600080fd5b8135602062000958620009528362000906565b620007f6565b8083825260208201915060208460051b8701019350868411156200097b57600080fd5b602086015b84811015620009aa57803560ff811681146200099c5760008081fd5b835291830191830162000980565b509695505050505050565b6000601f83601f840112620009c957600080fd5b82356020620009dc620009528362000906565b82815260059290921b85018101918181019087841115620009fc57600080fd5b8287015b8481101562000a9d57803567ffffffffffffffff8082111562000a235760008081fd5b818a0191508a603f83011262000a395760008081fd5b8582013560408282111562000a525762000a52620007b4565b62000a65828b01601f19168901620007f6565b92508183528c8183860101111562000a7d5760008081fd5b818185018985013750600090820187015284525091830191830162000a00565b50979650505050505050565b60008060006060848603121562000abf57600080fd5b833567ffffffffffffffff8082111562000ad857600080fd5b818601915086601f83011262000aed57600080fd5b8135602062000b00620009528362000906565b82815260059290921b8401810191818101908a84111562000b2057600080fd5b948201945b8386101562000b4b57853562000b3b816200083a565b8252948201949082019062000b25565b9750508701359250508082111562000b6257600080fd5b62000b70878388016200092d565b9350604086013591508082111562000b8757600080fd5b5062000b9686828701620009b5565b9150509250925092565b600060a082019050825160030b8252602083015160030b602083015263ffffffff6040840151166040830152606083015160030b6060830152608083015160030b608083015292915050565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601160045260246000fd5b8082018082111562000c2e5762000c2e62000c02565b92915050565b60008260030b8260030b028060030b915080821462000c575762000c5762000c02565b5092915050565b600060a0828403121562000c7157600080fd5b62000c7b620007ca565b825162000c88816200082a565b8152602083015162000c9a816200082a565b6020820152604083015162000caf816200083a565b6040820152606083015162000cc4816200082a565b60608201526080830151620008c8816200082a56fe60806040526040516105503803806105508339810160408190526100229161030d565b818161002e8282610042565b5061003a9050836100a1565b5050506103f9565b61004b8261010f565b6040516001600160a01b038316907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a2805115610095576100908282610153565b505050565b61009d6101ca565b5050565b7f7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f6100e1600080516020610530833981519152546001600160a01b031690565b604080516001600160a01b03928316815291841660208301520160405180910390a161010c816101eb565b50565b807f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc5b80546001600160a01b0319166001600160a01b039290921691909117905550565b6060600080846001600160a01b03168460405161017091906103dd565b600060405180830381855af49150503d80600081146101ab576040519150601f19603f3d011682016040523d82523d6000602084013e6101b0565b606091505b5090925090506101c185838361022f565b95945050505050565b34156101e95760405163b398979f60e01b815260040160405180910390fd5b565b6001600160a01b03811661021a57604051633173bdd160e11b8152600060048201526024015b60405180910390fd5b80600080516020610530833981519152610132565b6060826102445761023f8261028e565b610287565b815115801561025b57506001600160a01b0384163b155b1561028457604051639996b31560e01b81526001600160a01b0385166004820152602401610211565b50805b9392505050565b80511561029e5780518082602001fd5b604051630a12f52160e11b815260040160405180910390fd5b80516001600160a01b03811681146102ce57600080fd5b919050565b634e487b7160e01b600052604160045260246000fd5b60005b838110156103045781810151838201526020016102ec565b50506000910152565b60008060006060848603121561032257600080fd5b61032b846102b7565b9250610339602085016102b7565b60408501519092506001600160401b038082111561035657600080fd5b818601915086601f83011261036a57600080fd5b81518181111561037c5761037c6102d3565b604051601f8201601f19908116603f011681019083821181831017156103a4576103a46102d3565b816040528281528960208487010111156103bd57600080fd5b6103ce8360208301602088016102e9565b80955050505050509250925092565b600082516103ef8184602087016102e9565b9190910192915050565b610128806104086000396000f3fe608060405233301480602757506012603a565b6001600160a01b0316336001600160a01b0316145b156033576031606d565b005b603130607b565b60007fb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d61035b546001600160a01b0316919050565b6079607560a0565b60ad565b565b60603660008037600080366000855afa3d6000803e808015609b573d6000f35b3d6000fd5b600060a860cb565b905090565b3660008037600080366000845af43d6000803e808015609b573d6000f35b60007f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc605e56fea2646970667358221220b52b1f3734b6367f6f91f8c1dff8cbb80cdb82f252142a7744a3fc4af904c81c64736f6c63430008180033b53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103a264697066735822122066aa8001cbb97301ed11043557394730d8ee1d39fa682bac3300a4a71d94611b64736f6c63430008180033",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

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

// LastTickBlockNumber is a free data retrieval call binding the contract method 0xff280198.
//
// Solidity: function lastTickBlockNumber() view returns(uint256)
func (_Contract *ContractCaller) LastTickBlockNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "lastTickBlockNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastTickBlockNumber is a free data retrieval call binding the contract method 0xff280198.
//
// Solidity: function lastTickBlockNumber() view returns(uint256)
func (_Contract *ContractSession) LastTickBlockNumber() (*big.Int, error) {
	return _Contract.Contract.LastTickBlockNumber(&_Contract.CallOpts)
}

// LastTickBlockNumber is a free data retrieval call binding the contract method 0xff280198.
//
// Solidity: function lastTickBlockNumber() view returns(uint256)
func (_Contract *ContractCallerSession) LastTickBlockNumber() (*big.Int, error) {
	return _Contract.Contract.LastTickBlockNumber(&_Contract.CallOpts)
}

// Proxy is a free data retrieval call binding the contract method 0xec556889.
//
// Solidity: function proxy() view returns(address)
func (_Contract *ContractCaller) Proxy(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proxy")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Proxy is a free data retrieval call binding the contract method 0xec556889.
//
// Solidity: function proxy() view returns(address)
func (_Contract *ContractSession) Proxy() (common.Address, error) {
	return _Contract.Contract.Proxy(&_Contract.CallOpts)
}

// Proxy is a free data retrieval call binding the contract method 0xec556889.
//
// Solidity: function proxy() view returns(address)
func (_Contract *ContractCallerSession) Proxy() (common.Address, error) {
	return _Contract.Contract.Proxy(&_Contract.CallOpts)
}

// AddBody is a paid mutator transaction binding the contract method 0x22c5eafe.
//
// Solidity: function addBody((int32,int32,uint32,int32,int32) action) returns()
func (_Contract *ContractTransactor) AddBody(opts *bind.TransactOpts, action ActionDataAddBody) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "addBody", action)
}

// AddBody is a paid mutator transaction binding the contract method 0x22c5eafe.
//
// Solidity: function addBody((int32,int32,uint32,int32,int32) action) returns()
func (_Contract *ContractSession) AddBody(action ActionDataAddBody) (*types.Transaction, error) {
	return _Contract.Contract.AddBody(&_Contract.TransactOpts, action)
}

// AddBody is a paid mutator transaction binding the contract method 0x22c5eafe.
//
// Solidity: function addBody((int32,int32,uint32,int32,int32) action) returns()
func (_Contract *ContractTransactorSession) AddBody(action ActionDataAddBody) (*types.Transaction, error) {
	return _Contract.Contract.AddBody(&_Contract.TransactOpts, action)
}

// ExecuteMultipleActions is a paid mutator transaction binding the contract method 0xd0b36171.
//
// Solidity: function executeMultipleActions(uint32[] actionIds, uint8[] actionCount, bytes[] actionData) returns()
func (_Contract *ContractTransactor) ExecuteMultipleActions(opts *bind.TransactOpts, actionIds []uint32, actionCount []uint8, actionData [][]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "executeMultipleActions", actionIds, actionCount, actionData)
}

// ExecuteMultipleActions is a paid mutator transaction binding the contract method 0xd0b36171.
//
// Solidity: function executeMultipleActions(uint32[] actionIds, uint8[] actionCount, bytes[] actionData) returns()
func (_Contract *ContractSession) ExecuteMultipleActions(actionIds []uint32, actionCount []uint8, actionData [][]byte) (*types.Transaction, error) {
	return _Contract.Contract.ExecuteMultipleActions(&_Contract.TransactOpts, actionIds, actionCount, actionData)
}

// ExecuteMultipleActions is a paid mutator transaction binding the contract method 0xd0b36171.
//
// Solidity: function executeMultipleActions(uint32[] actionIds, uint8[] actionCount, bytes[] actionData) returns()
func (_Contract *ContractTransactorSession) ExecuteMultipleActions(actionIds []uint32, actionCount []uint8, actionData [][]byte) (*types.Transaction, error) {
	return _Contract.Contract.ExecuteMultipleActions(&_Contract.TransactOpts, actionIds, actionCount, actionData)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _logic) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _logic common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _logic)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _logic) returns()
func (_Contract *ContractSession) Initialize(_logic common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _logic)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _logic) returns()
func (_Contract *ContractTransactorSession) Initialize(_logic common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _logic)
}

// Tick is a paid mutator transaction binding the contract method 0x3eaf5d9f.
//
// Solidity: function tick() returns()
func (_Contract *ContractTransactor) Tick(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "tick")
}

// Tick is a paid mutator transaction binding the contract method 0x3eaf5d9f.
//
// Solidity: function tick() returns()
func (_Contract *ContractSession) Tick() (*types.Transaction, error) {
	return _Contract.Contract.Tick(&_Contract.TransactOpts)
}

// Tick is a paid mutator transaction binding the contract method 0x3eaf5d9f.
//
// Solidity: function tick() returns()
func (_Contract *ContractTransactorSession) Tick() (*types.Transaction, error) {
	return _Contract.Contract.Tick(&_Contract.TransactOpts)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Contract *ContractTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Contract.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Contract *ContractSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Contract.Contract.Fallback(&_Contract.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Contract *ContractTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Contract.Contract.Fallback(&_Contract.TransactOpts, calldata)
}

// ContractActionExecutedIterator is returned from FilterActionExecuted and is used to iterate over the raw logs and unpacked data for ActionExecuted events raised by the Contract contract.
type ContractActionExecutedIterator struct {
	Event *ContractActionExecuted // Event containing the contract specifics and raw log

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
func (it *ContractActionExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractActionExecuted)
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
		it.Event = new(ContractActionExecuted)
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
func (it *ContractActionExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractActionExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractActionExecuted represents a ActionExecuted event raised by the Contract contract.
type ContractActionExecuted struct {
	ActionId [4]byte
	Data     []byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterActionExecuted is a free log retrieval operation binding the contract event 0x45065f461aede1b904079823f6d858e465fa8c25fcf1654bb4a89e6dee320a1a.
//
// Solidity: event ActionExecuted(bytes4 actionId, bytes data)
func (_Contract *ContractFilterer) FilterActionExecuted(opts *bind.FilterOpts) (*ContractActionExecutedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ActionExecuted")
	if err != nil {
		return nil, err
	}
	return &ContractActionExecutedIterator{contract: _Contract.contract, event: "ActionExecuted", logs: logs, sub: sub}, nil
}

// WatchActionExecuted is a free log subscription operation binding the contract event 0x45065f461aede1b904079823f6d858e465fa8c25fcf1654bb4a89e6dee320a1a.
//
// Solidity: event ActionExecuted(bytes4 actionId, bytes data)
func (_Contract *ContractFilterer) WatchActionExecuted(opts *bind.WatchOpts, sink chan<- *ContractActionExecuted) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ActionExecuted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractActionExecuted)
				if err := _Contract.contract.UnpackLog(event, "ActionExecuted", log); err != nil {
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

// ParseActionExecuted is a log parse operation binding the contract event 0x45065f461aede1b904079823f6d858e465fa8c25fcf1654bb4a89e6dee320a1a.
//
// Solidity: event ActionExecuted(bytes4 actionId, bytes data)
func (_Contract *ContractFilterer) ParseActionExecuted(log types.Log) (*ContractActionExecuted, error) {
	event := new(ContractActionExecuted)
	if err := _Contract.contract.UnpackLog(event, "ActionExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Contract contract.
type ContractInitializedIterator struct {
	Event *ContractInitialized // Event containing the contract specifics and raw log

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
func (it *ContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractInitialized)
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
		it.Event = new(ContractInitialized)
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
func (it *ContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractInitialized represents a Initialized event raised by the Contract contract.
type ContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Contract *ContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractInitializedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractInitializedIterator{contract: _Contract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Contract *ContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractInitialized) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractInitialized)
				if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Contract *ContractFilterer) ParseInitialized(log types.Log) (*ContractInitialized, error) {
	event := new(ContractInitialized)
	if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
