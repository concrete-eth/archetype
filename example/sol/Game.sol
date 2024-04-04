// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "arch/ArchProxyAdmin.sol";

import "./solgen/IActions.sol";
import {Entrypoint} from "./solgen/EntryPoint.sol";
import {ICore} from "./solgen/ICore.sol";

contract Game is Entrypoint, ArchProxyAdmin, Initializable {
    function initialize(address _logic) public initializer {
        _createArchProxy(_logic);
        _initialize();
    }

    function _initialize() internal initializer {}

    function move(ActionData_Move memory action) public override {}
}
