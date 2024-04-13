// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Test, console2} from "forge-std/Test.sol";
import {ArchProxy} from "../src/ArchProxy.sol";
import {ArchProxyAdmin} from "../src/ArchProxyAdmin.sol";

contract TestLogic {
    uint256 public value;
    function setValue(uint256 _value) public {
        value = _value;
    }
}

contract TestAdmin is ArchProxyAdmin {
    function setProxy(address addr) public {
        _setProxy(addr);
    }

    function setValue(uint256 _value) public {
        TestLogic(address(proxy)).setValue(_value);
    }
}

contract ArchProxyAdminTest is Test {
    TestLogic internal logic;
    TestAdmin internal admin;
    ArchProxy internal proxy;

    function setUp() public virtual {
        admin = new TestAdmin();
        logic = new TestLogic();
        proxy = new ArchProxy(address(admin), address(logic), "");
        admin.setProxy(address(proxy));
    }

    function testProxyAdmin() public {
        // Set value through the admin
        admin.setValue(42);

        // Get value from the proxy
        assertEq(
            TestLogic(address(proxy)).value(),
            42,
            "admin-set proxy-get failed"
        );
        // Get value from the admin (t/fallback)
        assertEq(
            TestLogic(address(admin)).value(),
            42,
            "admin-set admin-get failed"
        );
        // Get value from the logic
        assertEq(
            logic.value(),
            0,
            "logic should not be affected by proxy write"
        );
    }
}
