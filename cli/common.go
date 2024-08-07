package cli

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/cobra"
)

func newRpcClient(cmd *cobra.Command) *rpc.Client {
	opts := make([]rpc.ClientOption, 0)

	jwtSecretHex, err := cmd.Flags().GetString("jwt-secret")
	if err != nil {
		logFatal(err)
	}
	if jwtSecretHex != "" {
		jwtSecret := common.HexToHash(jwtSecretHex)
		opts = append(opts, rpc.WithHTTPAuth(node.NewJWTAuth(jwtSecret)))
	}

	rpcUrl, err := cmd.Flags().GetString("rpc-url")
	if err != nil {
		logFatal(err)
	}

	rpcClient, err := rpc.DialOptions(context.Background(), rpcUrl, opts...)
	if err != nil {
		logFatal(err)
	}
	return rpcClient
}

func getAddress(cmd *cobra.Command) common.Address {
	addressHex, err := cmd.Flags().GetString("address")
	if err != nil {
		logFatal(err)
	}
	if addressHex == "" {
		logFatalNoContext(errors.New("address is required"))
	}
	checkIsHexAddress(addressHex)
	address := common.HexToAddress(addressHex)
	return address
}
