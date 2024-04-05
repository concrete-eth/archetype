// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

enum TableId {
    Config,
    Players
}

struct RowData_Config {
    uint64 startBlock;
    uint8 maxPlayers;
}

struct RowData_Players {
    int16 x;
    int16 y;
    uint8 health;
}

interface ITableGetter {
    function getConfigRow() external view returns (RowData_Config memory);
    function getPlayersRow(uint8 playerId) external view returns (RowData_Players memory);
}