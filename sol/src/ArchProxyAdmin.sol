// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {ArchProxy} from "./ArchProxy.sol";
import {CallUtils} from "./utils/CallUtils.sol";

contract ArchProxyAdmin {
    ArchProxy public proxy;

    function _createArchProxy(address _logic) internal {
        proxy = new ArchProxy(_logic, "");
    }

    function _fallback() internal view {
        CallUtils.forwardStaticcall(address(proxy));
    }

    fallback() external {
        _fallback();
    }
}
