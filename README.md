# Archetype

*Archetype is an engine for real-time onchain games on the OPStack.*

## Run run the demo locally

```bash
git clone https://github.com/concrete-eth/archetype.git && \
  cd archetype && \
  go run example/cmd/local/main.go
```

![arch_example_demo](https://github.com/user-attachments/assets/f4c064a7-7f19-4e49-bac3-6e5cfd52281e)

## Directories

```
Core
- client     : headless client for keeping local state in sync
- rpc        : objects that interact with the chain
- arch       : the arch interface
- precompile : wrapper to put core logic in a precompile

Libs
- kvstore  : key-value store
- utils    : math (max, min, abs, etc) and channel (fork, probe) utils
- sol      : solidity contracts (proxy and proxy admin)
- deploy   : helpers for local and remote deployment
- snapshot : geth rpc api for taking storage snapshots of smart contracts

CLI
- cli     : cli tool
- codegen : code generation

Testing
- e2e       : end-to-end test
- simulated : extended geth simulated backend
- testutils : counter app for testing

Example
- example : example game
```
