// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {CallUtils} from "./utils/CallUtils.sol";
import {console2} from "forge-std/Test.sol";

contract ArchProxyAdmin {
    address public proxy;

    function _setProxy(address addr) internal {
        require(addr != address(0), "ArchProxyAdmin: invalid proxy address");
        require(
            address(proxy) == address(0),
            "ArchProxyAdmin: proxy already set"
        );
        proxy = addr;
    }

    function _fallback() internal view {
        require(address(proxy) != address(0), "ArchProxyAdmin: proxy not set");
        CallUtils.forwardStaticcall(address(proxy));
    }

    fallback() external {
        _fallback();
    }
}
