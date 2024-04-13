// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Test, console2} from "forge-std/Test.sol";
import {ArchProxy} from "../src/ArchProxy.sol";

contract TestLogic {
    uint256 public value;
    function setValue(uint256 _value) public {
        value = _value;
    }
}

contract ArchProxyTest is Test {
    TestLogic internal logic;
    ArchProxy internal proxy;

    function setUp() public virtual {
        logic = new TestLogic();
        proxy = new ArchProxy(address(this), address(logic), "");
    }

    function testProxyAdmin() public {
        // Set value in the proxy
        TestLogic(address(proxy)).setValue(42);

        // Get value from the proxy
        assertEq(TestLogic(address(proxy)).value(), 42, "proxy set-get failed");
        // Get value from the logic
        assertEq(
            logic.value(),
            0,
            "logic should not be affected by proxy write"
        );

        // Set value in the logic
        logic.setValue(53);

        // Get value from the logic
        assertEq(logic.value(), 53, "logic set-get failed");
        // Get value from the proxy
        assertEq(
            TestLogic(address(proxy)).value(),
            42,
            "proxy should not be affected by logic write"
        );
    }

    function testReadFromNonAdmin() public {
        TestLogic(address(proxy)).setValue(42);
        vm.prank(address(0));
        assertEq(TestLogic(address(proxy)).value(), 42, "proxy get failed");
    }

    function testFailWriteFromNonAdmin() public {
        vm.prank(address(0));
        TestLogic(address(proxy)).setValue(42);
    }
}
