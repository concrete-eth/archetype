// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import "../sol/solgen/ICore.sol";
import "../sol/solgen/IActions.sol";
import "../sol/solgen/ITables.sol";
import "../sol/Game.sol";

contract Test {
    Game internal game;

    function setUp() public virtual {
        game = new Game();
        game.initialize(address(0x80), "");
    }

    function assertEq(string memory title, uint256 a, uint256 b) internal pure {
        if (a == b) {
            return;
        }
        string memory message = string(
            abi.encodePacked(title, ": Expected ", a, " but got ", b)
        );
        revert(message);
    }

    function assertEq(uint256 a, uint256 b) internal pure {
        assertEq("", a, b);
    }

    function assertEq(string memory title, int32 a, int32 b) internal pure {
        assertEq(title, uint256(int256(a)), uint256(int256(b)));
    }

    function assertEq(int32 a, int32 b) internal pure {
        assertEq("", uint256(int256(a)), uint256(int256(b)));
    }

    function testAddBody() public {
        ActionData_AddBody memory action = ActionData_AddBody({
            x: 1,
            y: 2,
            r: 3,
            vx: 4,
            vy: 5
        });
        game.addBody(action);
        uint8 bodyId = ICore(address(game)).getMetaRow().bodyCount;
        RowData_Bodies memory body = ICore(address(game)).getBodiesRow(bodyId);
        assertEq("X", body.x, action.x);
        assertEq("Y", body.y, action.y);
        assertEq("R", body.r, action.r);
        assertEq("Vx", body.vx, action.vx);
        assertEq("Vy", body.vy, action.vy);
    }
}
