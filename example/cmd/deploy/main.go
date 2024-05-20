package main

import (
	"fmt"

	"github.com/concrete-eth/archetype/example/cmd"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	privateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	pcAddr        = common.HexToAddress("0x80")
	rpcUrl        = "ws://localhost:9546"
)

func main() {
	// Connect to rpc
	fmt.Println("Connecting to", rpcUrl)
	ethcli, chainId, err := cmd.NewEthClient(rpcUrl)
	if err != nil {
		panic(err)
	}

	// Load tx opts
	fmt.Println("Loading private key")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		panic(err)
	}

	// Deploy game
	fmt.Println("Deploying game")
	gameAddr, coreAddr, err := cmd.DeployGame(ethcli, auth, pcAddr)
	if err != nil {
		panic(err)
	}

	// Print addresses
	fmt.Println("Game:", gameAddr.Hex())
	fmt.Println("Core:", coreAddr.Hex())
}
