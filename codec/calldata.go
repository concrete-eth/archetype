package codec

import (
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func ActionToCalldata(actionAbi abi.ABI, actionMap archtypes.ActionMap, action archtypes.Action) ([]byte, error) {
	return nil, nil
}

func CalldataToAction(actionAbi abi.ABI, actionMap archtypes.ActionMap, data []byte) (archtypes.Action, error) {
	return nil, nil
}

func ArgsToAction(method *abi.Method, args []interface{}) (archtypes.Action, error) {
	return nil, nil
}
