// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {Test, console2} from "forge-std/Test.sol";

import "../sol/solgen/ICore.sol";
import "../sol/solgen/IActions.sol";
import "../sol/solgen/ITables.sol";
import "../sol/Game.sol";

contract TestLogic is ICore {
    RowData_Bodies[] public bodies;

    function addBody(ActionData_AddBody memory action) external {
        bodies.push(
            RowData_Bodies({
                x: action.x,
                y: action.y,
                r: action.r,
                vx: action.vx,
                vy: action.vy,
                ax: 0,
                ay: 0
            })
        );
    }

    function tick() external {}

    function getMetaRow() external view returns (RowData_Meta memory) {
        return RowData_Meta({maxBodyCount: 0, bodyCount: uint8(bodies.length)});
    }

    function getBodiesRow(
        uint8 bodyId
    ) external view returns (RowData_Bodies memory) {
        return bodies[bodyId - 1];
    }
}

contract GameTest is Test {
    TestLogic logic;
    Game internal game;

    function setUp() public virtual {
        logic = new TestLogic();
        game = new Game();
        game.initialize(address(logic));
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
        assertEq(body.x, action.x);
        assertEq(body.y, action.y);
        assertEq(body.r, action.r);
        assertEq(body.vx, action.vx);
        assertEq(body.vy, action.vy);
    }
}
