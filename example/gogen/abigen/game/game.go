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
	ABI: "[{\"type\":\"fallback\",\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addBody\",\"inputs\":[{\"name\":\"action\",\"type\":\"tuple\",\"internalType\":\"structActionData_AddBody\",\"components\":[{\"name\":\"x\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"y\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"r\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"vx\",\"type\":\"int32\",\"internalType\":\"int32\"},{\"name\":\"vy\",\"type\":\"int32\",\"internalType\":\"int32\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executeMultipleActions\",\"inputs\":[{\"name\":\"actionIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"actionCount\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"},{\"name\":\"actionData\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_logic\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"lastTickBlockNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proxy\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tick\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]}]",
	Bin: "0x6080604052348015600f57600080fd5b506112d98061001f6000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c806322c5eafe1461006c5780633eaf5d9f1461007f578063d0b3617114610087578063d1f578941461009a578063ec556889146100ad578063ff280198146100dd575b61006a6100f4565b005b61006a61007a36600461089c565b610169565b61006a6101ce565b61006a610095366004610aa9565b61032f565b61006a6100a8366004610b8f565b6103ce565b6000546100c0906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100e660025481565b6040519081526020016100d4565b6000546001600160a01b03166101515760405162461bcd60e51b815260206004820152601d60248201527f4172636850726f787941646d696e3a2070726f7879206e6f742073657400000060448201526064015b60405180910390fd5b600054610166906001600160a01b0316610542565b50565b600054604051631162f57f60e11b81526001600160a01b03909116906322c5eafe90610199908490600401610beb565b600060405180830381600087803b1580156101b357600080fd5b505af11580156101c7573d6000803e3d6000fd5b5050505050565b60025443116102105760405162461bcd60e51b815260206004820152600e60248201526d185b1c9958591e481d1a58dad95960921b6044820152606401610148565b6002600154036102795760008054604080516370f0c35160e01b815290516001600160a01b03909216926370f0c3519260048084019382900301818387803b15801561025b57600080fd5b505af115801561026f573d6000803e3d6000fd5b5050600180555050565b600080546001600160a01b03166127105a6102949190610c4d565b60408051600481526024810182526020810180516001600160e01b0316633eaf5d9f60e01b17905290516102c89190610c66565b60006040518083038160008787f1925050503d8060008114610306576040519150601f19603f3d011682016040523d82523d6000602084013e61030b565b606091505b5050905080156103185750565b6175305a101561032a57600260015550565b600080fd5b6000805b84518110156101c757600084828151811061035057610350610c95565b602002602001015160ff16905060008390505b61036d8285610cab565b8110156103b8576103b087848151811061038957610389610c95565b60200260200101518683815181106103a3576103a3610c95565b6020026020010151610568565b600101610363565b506103c38184610cab565b925050600101610333565b7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a008054600160401b810460ff16159067ffffffffffffffff166000811580156104145750825b905060008267ffffffffffffffff1660011480156104315750303b155b90508115801561043f575080155b1561045d5760405163f92ee8a960e01b815260040160405180910390fd5b845467ffffffffffffffff19166001178555831561048757845460ff60401b1916600160401b1785555b60003088604051610497906107fe565b6001600160a01b03928316815291166020820152606060408201819052600090820152608001604051809103906000f0801580156104d9573d6000803e3d6000fd5b5090506104e581610604565b6104ee876106ed565b5060018055831561053957845460ff60401b19168555604051600181527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b50505050505050565b60603660008037600080366000855afa3d6000803e808015610563573d6000f35b3d6000fd5b8163ffffffff16633eaf5d9f03610585576105816101ce565b5050565b8163ffffffff166322c5eafe036105bc576000818060200190518101906105ac9190610cbe565b90506105b781610169565b505050565b60405162461bcd60e51b815260206004820152601d60248201527f456e747279706f696e743a20496e76616c696420616374696f6e2049440000006044820152606401610148565b6001600160a01b0381166106685760405162461bcd60e51b815260206004820152602560248201527f4172636850726f787941646d696e3a20696e76616c69642070726f7879206164604482015264647265737360d81b6064820152608401610148565b6000546001600160a01b0316156106cb5760405162461bcd60e51b815260206004820152602160248201527f4172636850726f787941646d696e3a2070726f787920616c72656164792073656044820152601d60fa1b6064820152608401610148565b600080546001600160a01b0319166001600160a01b0392909216919091179055565b6107076000806106ff60646006610d2c565b600080610764565b6107396107176064603b19610d2c565b600061072560646002610d2c565b60006107346064600319610d2c565b610764565b6101666107486064603c610d2c565b600061075660646002610d2c565b600061073460646004610d2c565b6000546040805160a081018252600388810b825287810b602083015263ffffffff87168284015285810b606083015284900b60808201529051631162f57f60e11b81526001600160a01b03909216916322c5eafe916107c591600401610beb565b600060405180830381600087803b1580156107df57600080fd5b505af11580156107f3573d6000803e3d6000fd5b505050505050505050565b61055080610d5483390190565b634e487b7160e01b600052604160045260246000fd5b60405160a0810167ffffffffffffffff811182821017156108445761084461080b565b60405290565b604051601f8201601f1916810167ffffffffffffffff811182821017156108735761087361080b565b604052919050565b8060030b811461016657600080fd5b63ffffffff8116811461016657600080fd5b600060a082840312156108ae57600080fd5b6108b6610821565b82356108c18161087b565b815260208301356108d18161087b565b602082015260408301356108e48161088a565b604082015260608301356108f78161087b565b6060820152608083013561090a8161087b565b60808201529392505050565b600067ffffffffffffffff8211156109305761093061080b565b5060051b60200190565b600082601f83011261094b57600080fd5b8135602061096061095b83610916565b61084a565b8083825260208201915060208460051b87010193508684111561098257600080fd5b602086015b848110156109ae57803560ff811681146109a15760008081fd5b8352918301918301610987565b509695505050505050565b600082601f8301126109ca57600080fd5b813567ffffffffffffffff8111156109e4576109e461080b565b6109f7601f8201601f191660200161084a565b818152846020838601011115610a0c57600080fd5b816020850160208301376000918101602001919091529392505050565b600082601f830112610a3a57600080fd5b81356020610a4a61095b83610916565b82815260059290921b84018101918181019086841115610a6957600080fd5b8286015b848110156109ae57803567ffffffffffffffff811115610a8d5760008081fd5b610a9b8986838b01016109b9565b845250918301918301610a6d565b600080600060608486031215610abe57600080fd5b833567ffffffffffffffff80821115610ad657600080fd5b818601915086601f830112610aea57600080fd5b81356020610afa61095b83610916565b82815260059290921b8401810191818101908a841115610b1957600080fd5b948201945b83861015610b40578535610b318161088a565b82529482019490820190610b1e565b97505087013592505080821115610b5657600080fd5b610b628783880161093a565b93506040860135915080821115610b7857600080fd5b50610b8586828701610a29565b9150509250925092565b60008060408385031215610ba257600080fd5b82356001600160a01b0381168114610bb957600080fd5b9150602083013567ffffffffffffffff811115610bd557600080fd5b610be1858286016109b9565b9150509250929050565b600060a082019050825160030b8252602083015160030b602083015263ffffffff6040840151166040830152606083015160030b6060830152608083015160030b608083015292915050565b634e487b7160e01b600052601160045260246000fd5b81810381811115610c6057610c60610c37565b92915050565b6000825160005b81811015610c875760208186018101518583015201610c6d565b506000920191825250919050565b634e487b7160e01b600052603260045260246000fd5b80820180821115610c6057610c60610c37565b600060a08284031215610cd057600080fd5b610cd8610821565b8251610ce38161087b565b81526020830151610cf38161087b565b60208201526040830151610d068161088a565b60408201526060830151610d198161087b565b6060820152608083015161090a8161087b565b60008260030b8260030b028060030b9150808214610d4c57610d4c610c37565b509291505056fe60806040526040516105503803806105508339810160408190526100229161030d565b818161002e8282610042565b5061003a9050836100a1565b5050506103f9565b61004b8261010f565b6040516001600160a01b038316907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a2805115610095576100908282610153565b505050565b61009d6101ca565b5050565b7f7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f6100e1600080516020610530833981519152546001600160a01b031690565b604080516001600160a01b03928316815291841660208301520160405180910390a161010c816101eb565b50565b807f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc5b80546001600160a01b0319166001600160a01b039290921691909117905550565b6060600080846001600160a01b03168460405161017091906103dd565b600060405180830381855af49150503d80600081146101ab576040519150601f19603f3d011682016040523d82523d6000602084013e6101b0565b606091505b5090925090506101c185838361022f565b95945050505050565b34156101e95760405163b398979f60e01b815260040160405180910390fd5b565b6001600160a01b03811661021a57604051633173bdd160e11b8152600060048201526024015b60405180910390fd5b80600080516020610530833981519152610132565b6060826102445761023f8261028e565b610287565b815115801561025b57506001600160a01b0384163b155b1561028457604051639996b31560e01b81526001600160a01b0385166004820152602401610211565b50805b9392505050565b80511561029e5780518082602001fd5b604051630a12f52160e11b815260040160405180910390fd5b80516001600160a01b03811681146102ce57600080fd5b919050565b634e487b7160e01b600052604160045260246000fd5b60005b838110156103045781810151838201526020016102ec565b50506000910152565b60008060006060848603121561032257600080fd5b61032b846102b7565b9250610339602085016102b7565b60408501519092506001600160401b038082111561035657600080fd5b818601915086601f83011261036a57600080fd5b81518181111561037c5761037c6102d3565b604051601f8201601f19908116603f011681019083821181831017156103a4576103a46102d3565b816040528281528960208487010111156103bd57600080fd5b6103ce8360208301602088016102e9565b80955050505050509250925092565b600082516103ef8184602087016102e9565b9190910192915050565b610128806104086000396000f3fe608060405233301480602757506012603a565b6001600160a01b0316336001600160a01b0316145b156033576031606d565b005b603130607b565b60007fb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d61035b546001600160a01b0316919050565b6079607560a0565b60ad565b565b60603660008037600080366000855afa3d6000803e808015609b573d6000f35b3d6000fd5b600060a860cb565b905090565b3660008037600080366000845af43d6000803e808015609b573d6000f35b60007f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc605e56fea264697066735822122080fba3c7bee25cb2cde5ad6be21f260b752c8d5921f88a4ef4f96c6f73b8f06464736f6c63430008190033b53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103a264697066735822122089eb685184c7bf9c9c1da4e160a302e77bea793633026f24a4cc8562f1f9f3cf64736f6c63430008190033",
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

// Initialize is a paid mutator transaction binding the contract method 0xd1f57894.
//
// Solidity: function initialize(address _logic, bytes data) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _logic common.Address, data []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _logic, data)
}

// Initialize is a paid mutator transaction binding the contract method 0xd1f57894.
//
// Solidity: function initialize(address _logic, bytes data) returns()
func (_Contract *ContractSession) Initialize(_logic common.Address, data []byte) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _logic, data)
}

// Initialize is a paid mutator transaction binding the contract method 0xd1f57894.
//
// Solidity: function initialize(address _logic, bytes data) returns()
func (_Contract *ContractTransactorSession) Initialize(_logic common.Address, data []byte) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _logic, data)
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
