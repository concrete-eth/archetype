// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */


struct ActionData_Add {
    int16 summand;
}


interface IActions {
    event ActionExecuted(bytes4 actionId, bytes data);

    function tick() external;


    function add(ActionData_Add memory action) external;
}
