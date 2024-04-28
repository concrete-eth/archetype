// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "arch/ArchProxyAdmin.sol";
import {ArchProxy} from "arch/ArchProxy.sol";

import "./solgen/IActions.sol";
import {Entrypoint} from "./solgen/EntryPoint.sol";
import "./solgen/ICore.sol";

contract Game is Entrypoint, ArchProxyAdmin, Initializable {
    int32 internal constant scale = 100;

    function initialize(address _logic) public initializer {
        address proxyAddress = address(
            new ArchProxy(address(this), _logic, "")
        );
        _setProxy(proxyAddress);
        _initialize();
    }

    function _initialize() internal {
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

    uint256 public lastTickBlockNumber;

    function tick() public override {
        require(block.number > lastTickBlockNumber, "Game: tick too soon");
        ICore(proxy).tick();
    }
}
