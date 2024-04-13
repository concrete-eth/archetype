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

    function _initialize() internal initializer {}

    function _addBody(int16 x, int16 y, uint16 m, int16 vx, int16 vy) internal {
        ICore(proxy).addBody(
            ActionData_AddBody({x: x, y: y, m: m, vx: vx, vy: vy})
        );
    }

    function addBody(ActionData_AddBody memory action) public override {
        (int16 x, int16 y) = (action.x, action.y);
        uint16 m = action.m;
        (int16 vx, int16 vy) = (action.vx, action.vy);
        _addBody(x, y, m, vx, vy);
    }
}
