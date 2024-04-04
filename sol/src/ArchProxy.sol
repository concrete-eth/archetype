// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {ERC1967Utils} from "openzeppelin/proxy/ERC1967/ERC1967Utils.sol";

import {ERC1967PrecompileProxy} from "./lib/ERC1967PrecompileProxy.sol";
import {CallUtils} from "./lib/CallUtils.sol";

contract ArchProxy is ERC1967PrecompileProxy {
    constructor(
        address _logic,
        bytes memory _data
    ) payable ERC1967PrecompileProxy(_logic, _data) {
        ERC1967Utils.changeAdmin(msg.sender);
    }

    fallback() external payable virtual override {
        if (
            msg.sender == address(this) || msg.sender == ERC1967Utils.getAdmin()
        ) {
            _fallback();
        } else {
            // Static call to self -> delegate call to implementation to achieve
            // static-delegate call behavior
            CallUtils.forwardStaticcall(address(this));
        }
    }
}
