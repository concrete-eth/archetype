// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {ERC1967Utils} from "openzeppelin/proxy/ERC1967/ERC1967Utils.sol";
import {Proxy} from "openzeppelin/proxy/Proxy.sol";

import {ERC1967PrecompileUtils} from "./ERC1967PrecompileUtils.sol";

contract ERC1967PrecompileProxy is Proxy {
    constructor(address implementation, bytes memory _data) payable {
        ERC1967PrecompileUtils.upgradeToAndCall(implementation, _data);
    }

    function _implementation()
        internal
        view
        virtual
        override
        returns (address)
    {
        return ERC1967Utils.getImplementation();
    }
}
