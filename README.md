Todo

- Clean concrete CLI and more
- Port from ark: evm, gas, rpc, headless, precompile, snapshot, kvstore, sol proxy and tick master, snapshot sidecar
- Port lib from ark: ebiten renderer, decren, utils, simulated, web play

# How it works

# Cli

## Codegen

Given a definition for tables where data is stored and pre-defined actions 

Generate Solidity and Golang bindings for a Concrete Datamod definition

Takes a tables.json and actions.json and generates Solidity and Golang.

tables.json: Defines one or more tables e.g.,

```json
{
    "config": {
        "schema": {
            "startBlock": "uint64",
            "maxPlayers": "uint8"
        }
    },
    "players": {
        "keySchema": {
            "playerId": "uint8"
        },
        "schema": {
            "x": "int16",
            "y": "int16",
            "health": "uint8"
        }
    }
}
```

actions.json: Defines one or more actions e.g.,

```json
{
    "tick": {
        "schema": {}
    },
    "move": {
        "schema": {
            "playerId": "uint8",
            "direction": "uint8"
        }
    }
}
```

Built in actions
Tick

Built in tables
Tick config
- TicksPerBlock
