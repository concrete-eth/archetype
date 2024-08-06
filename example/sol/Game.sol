// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import "./solgen/IActions.sol";
import "./solgen/ICore.sol";
import {Arch} from "./solgen/Arch.sol";

contract Game is Arch {
    int32 internal constant scale = 100;

    function _initialize(bytes memory data) internal override {
        _addBody(0, 0, uint32(6 * scale), 0, 0);
        _addBody(-60 * scale, 0, uint32(2 * scale), 0, -4 * scale);
        _addBody(60 * scale, 0, uint32(2 * scale), 0, 4 * scale);
    }

    function _addBody(int32 x, int32 y, uint32 r, int32 vx, int32 vy) internal {
        ICore(proxy).addBody(
            ActionData_AddBody({x: x, y: y, r: r, vx: vx, vy: vy})
        );
    }

    function addBody(ActionData_AddBody memory action) public override {
        ICore(proxy).addBody(action);
    }
}
