// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

{{ range .Imports }}
import "{{ . }}";
{{- end }}

import {Initializable} from "openzeppelin/proxy/Utils/Initializable.sol";
import {ArchProxyAdmin} from "arch/ArchProxyAdmin.sol";
import {ArchProxy} from "arch/ArchProxy.sol";

uint256 constant NonZeroBoolean_False = 1;
uint256 constant NonZeroBoolean_True = 2;

abstract contract {{$.Name}} is {{ range $i, $v := .Interfaces }}{{ if $i }}, {{ end }}{{ $v }}{{ end }} {
    uint256 internal needsPurge;

    function initialize(address _logic, bytes memory data) public initializer {
        address proxyAddress = address(
            new ArchProxy(address(this), _logic, "")
        );
        _setProxy(proxyAddress);
        _initialize(data);
        needsPurge = NonZeroBoolean_False;
    }

    function _initialize(bytes memory data) internal virtual;

    uint256 public lastTickBlockNumber;

    function tick() public virtual override {
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
        if (gasleft() < 10000+20000) {
            needsPurge = NonZeroBoolean_True;
        } else {
            revert();
        }
    }
}
