// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "arch/ArchProxyAdmin.sol";
import {ArchProxy} from "arch/ArchProxy.sol";

import "./solgen/IActions.sol";
import {Entrypoint} from "./solgen/EntryPoint.sol";
import "./solgen/ICore.sol";

contract Game is Entrypoint, ArchProxyAdmin, Initializable {
    function initialize(address _logic) public initializer {
        address proxyAddress = address(
            new ArchProxy(address(this), _logic, "")
        );
        _setProxy(proxyAddress);
        _initialize();
    }

    function _initialize() internal {
        _addBody(0, 0, 30, 0, 0);
        _addBody(-275, 0, 15, 0, -15);
        _addBody(275, 0, 15, 0, 15);
    }

    function _addBody(int32 x, int32 y, uint32 r, int32 vx, int32 vy) internal {
        ICore(proxy).addBody(
            ActionData_AddBody({x: x, y: y, r: r, vx: vx, vy: vy})
        );
    }

    function addBody(ActionData_AddBody memory action) public override {
        (int32 x, int32 y) = (action.x, action.y);
        uint32 r = action.r;
        (int32 vx, int32 vy) = (action.vx, action.vy);
        // ... do something with x, y, r, vx, vy
        _addBody(x, y, r, vx, vy);
    }

    function tick() public {
        // TODO: implement
        // ICore(proxy).tick();
    }
}
