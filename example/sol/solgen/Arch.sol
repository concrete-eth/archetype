// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

import "./ICore.sol";
import "./Entrypoint.sol";

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "arch/ArchProxyAdmin.sol";
import {ArchProxy} from "arch/ArchProxy.sol";

abstract contract Arch is Entrypoint, ArchProxyAdmin, Initializable {
    function initialize(address _logic, bytes memory data) public initializer {
        address proxyAddress = address(
            new ArchProxy(address(this), _logic, "")
        );
        _setProxy(proxyAddress);
        _initialize(data);
    }

    function _initialize(bytes memory data) internal virtual;

    uint256 public lastTickBlockNumber;

    function tick() public virtual override {
        require(block.number > lastTickBlockNumber, "already ticked");
        ICore(proxy).tick();
    }
}
