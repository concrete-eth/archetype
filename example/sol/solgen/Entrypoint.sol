// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

import "./IActions.sol";

abstract contract Entrypoint is IActions {
    function executeMultipleActions(
        uint32[] memory actionIds,
        uint8[] memory actionCount,
        bytes[] memory actionData
    ) external {
        uint256 actionIdx = 0;
        for (uint256 i = 0; i < actionIds.length; i++) {
            uint256 numActions = uint256(actionCount[i]);
            for (uint256 j = actionIdx; j < actionIdx + numActions; j++) {
                _executeAction(actionIds[i], actionData[j]);
            }
            actionIdx += numActions;
        }
    }

    function _executeAction(uint32 actionId, bytes memory actionData) private {
        if (actionId == 0x3eaf5d9f) {
            tick();
        } else if (actionId == 0x22c5eafe) {
            ActionData_AddBody memory action = abi.decode(
                actionData,
                (ActionData_AddBody)
            );
            addBody(action);
        } else {
            revert("Entrypoint: Invalid action ID");
        }
    }

    function tick() public virtual;

    function addBody(ActionData_AddBody memory action) public virtual;
}
