// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "./ArchProxyAdmin.sol";
import {ArchProxy} from "./ArchProxy.sol";

interface ICore {
    function tick() external;
    function purge() external;
}

uint256 constant NonZeroBoolean_False = 1;
uint256 constant NonZeroBoolean_True = 2;

abstract contract ArchBase is ArchProxyAdmin, Initializable {
    uint256 internal needsPurge;
    uint256 public lastTickBlockNumber;

    function tick() public virtual {
        require(block.number > lastTickBlockNumber, "already ticked");
        if (needsPurge == NonZeroBoolean_True) {
            ICore(proxy).purge();
            needsPurge = NonZeroBoolean_False;
            return;
        }
        (bool success, ) = proxy.call{gas: gasleft() - 10000}(
            abi.encodeWithSignature("tick()")
        );
        if (success) {
            return;
        }
        // The tick method SHOULD NEVER FAIL for reasons other than out-of-gas, so we can be very
        // aggressive when determining whether the method ran out of gas.
        if (gasleft() < 10000 + 20000) {
            needsPurge = NonZeroBoolean_True;
        } else {
            revert();
        }
    }

    function initialize(address _logic, bytes memory data) public initializer {
        address proxyAddress = address(
            new ArchProxy(address(this), _logic, "")
        );
        _setProxy(proxyAddress);
        _initialize(data);
        needsPurge = NonZeroBoolean_False;
    }

    function _initialize(bytes memory data) internal virtual;
}
