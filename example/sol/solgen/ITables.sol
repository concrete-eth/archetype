// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

struct RowData_Meta {
    uint8 maxBodyCount;
    uint8 bodyCount;
}

struct RowData_Bodies {
    int32 x;
    int32 y;
    uint32 r;
    int32 vx;
    int32 vy;
}

interface ITables {
    function getMeta() external view returns (RowData_Meta memory);
    function getBodies(
        uint8 bodyId
    ) external view returns (RowData_Bodies memory);
}
