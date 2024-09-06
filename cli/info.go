package cli

import (
	"context"
	"encoding/json"

	admin_contract "github.com/concrete-eth/archetype/abigen/arch_proxy_admin"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

var implementationSlot = common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")

func runInfo(cmd *cobra.Command, args []string) {
	address := getAddress(cmd)
	rpcClient := newRpcClient(cmd)
	ethcli := ethclient.NewClient(rpcClient)

	adminContract, err := admin_contract.NewContract(address, ethcli)
	if err != nil {
		logFatal(err)
	}

	proxyAddr, err := adminContract.Proxy(nil)
	if err != nil {
		logFatal(err)
	}

	logicAddrBytes, err := ethcli.StorageAt(context.Background(), proxyAddr, implementationSlot, nil)
	if err != nil {
		logFatal(err)
	}
	logicAddr := common.BytesToAddress(logicAddrBytes)

	info := struct {
		Game  string `json:"game"`
		Proxy string `json:"proxy"`
		Logic string `json:"logic"`
	}{address.Hex(), proxyAddr.Hex(), logicAddr.Hex()}

	jsonStr, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		logFatal(err)
	}
	logInfo(string(jsonStr))
}

func AddInfoCommand(parent *cobra.Command) {
	infoCommand := &cobra.Command{Use: "info", Short: "Get game contract info", Run: runInfo}
	infoCommand.PersistentFlags().StringP("address", "a", "", "game contract address")
	addRpcFlags(infoCommand)
	parent.AddCommand(infoCommand)
}
