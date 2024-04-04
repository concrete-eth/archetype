// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

import "./IActions.sol";

abstract contract Entrypoint is IActionExecutor {
    function executeMultipleActions(
        uint8[] memory actionIds,
        uint8[] memory actionNumber,
        bytes[] memory actionData
    ) external {
        uint256 actionIdx = 0;
        for (uint256 i = 0; i < actionIds.length; i++) {
            uint256 numActions = uint256(actionNumber[i]);
            for (uint256 j = actionIdx; j < actionIdx + numActions; j++) {
                _executeAction(actionIds[i], actionData[j]);
            }
            actionIdx += numActions;
        }
    }

    function _executeAction(uint8 actionId, bytes memory actionData) private {
        if (actionId == 0) {
            ActionData_Move memory action = abi.decode(
                actionData,
                (ActionData_Move)
            );
            move(action);
        } else {
            revert("Entrypoint: Invalid action ID");
        }
    }
    
    function move(ActionData_Move memory action) public virtual;
}