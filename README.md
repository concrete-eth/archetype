- gas estimation -- later

```
Core
- client     : headless client keeping local state in sync
- rpc        : objects that interact with the chain (action subs, tx sender, tx hinter)
- arch       : the arch interface
- precompile : wrapper to put core logic in a precompile

Libs
- kvstore : key-value store
- utils   : math (max, min, abs, etc) and channel (fork, probe) utils
- sol     : solidity contracts (proxy and proxy admin)

CLI
- cli     : cli tool
- codegen : code generation

Testing
- e2e       : end-to-end test
- simulated : extended geth simulated backend for testing
- testutils : counter app for testing

Example
- example : example app
```
