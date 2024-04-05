// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Address} from "openzeppelin/utils/Address.sol";
import {StorageSlot} from "openzeppelin/utils/StorageSlot.sol";
import {ERC1967Utils} from "openzeppelin/proxy/ERC1967/ERC1967Utils.sol";

library ERC1967PrecompileUtils {
    function _setImplementation(address newImplementation) private {
        StorageSlot
            .getAddressSlot(ERC1967Utils.IMPLEMENTATION_SLOT)
            .value = newImplementation;
    }

    function upgradeToAndCall(
        address newImplementation,
        bytes memory data
    ) internal {
        _setImplementation(newImplementation);
        emit ERC1967Utils.Upgraded(newImplementation);

        if (data.length > 0) {
            Address.functionDelegateCall(newImplementation, data);
        } else {
            _checkNonPayable();
        }
    }

    function _checkNonPayable() private {
        if (msg.value > 0) {
            revert ERC1967Utils.ERC1967NonPayable();
        }
    }
}
