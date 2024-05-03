package cli

import (
	"encoding/json"
	"errors"

	"github.com/concrete-eth/archetype/example/engine"
	"github.com/concrete-eth/archetype/snapshot"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/spf13/cobra"
)

/*

type SnapshotWriter interface {
	New(query FilterQuery) ([]SnapshotMetadataWithStatus, error)
	Update(query FilterQuery) ([]SnapshotMetadataWithStatus, error)
	Delete(query FilterQuery) error
	Prune() error
	AddSchedule(schedule snapshot_types.Schedule) (snapshot_types.ScheduleResponse, error)
	DeleteSchedule(id uint64) error
}

type SnapshotReader interface {
	Get(address common.Address, blockHash common.Hash) (SnapshotResponse, error)
	Last(address common.Address) (SnapshotMetadataWithStatus, error)
	List(address common.Address) ([]SnapshotMetadataWithStatus, error)
	GetSchedules() (map[uint64]snapshot_types.Schedule, error) // TODO: get schedule by id/address?
}

*/

func methodName(name string) string {
	return engine.SnapshotNamespace + "_" + name
}

func newRpcClient(cmd *cobra.Command) *rpc.Client {
	// TODO: secret
	rpcUrl, err := cmd.Flags().GetString("rpc-url")
	if err != nil {
		logFatal(err)
	}
	rpcClient, err := rpc.Dial(rpcUrl)
	if err != nil {
		logFatal(err)
	}
	return rpcClient
}

func getSnapshotQuery(cmd *cobra.Command) *snapshot.SnapshotQuery {
	blockHashHex, err := cmd.Flags().GetString("block-hash")
	if err != nil {
		logFatal(err)
	}
	addressesHex, err := cmd.Flags().GetStringSlice("address")
	if err != nil {
		logFatal(err)
	}

	blockHash := common.HexToHash(blockHashHex)
	addresses := make([]common.Address, len(addressesHex))

	for _, addressHex := range addressesHex {
		addresses = append(addresses, common.HexToAddress(addressHex))
	}

	return &snapshot.SnapshotQuery{
		BlockHash: blockHash,
		Addresses: addresses,
	}
}

func runNewSnapshot(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	query := getSnapshotQuery(cmd)
	if err := rpcClient.Call(nil, methodName("new"), query); err != nil {
		logFatal(err)
	}
}

func runDeleteSnapshot(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	query := getSnapshotQuery(cmd)
	if err := rpcClient.Call(nil, methodName("delete"), query); err != nil {
		logFatal(err)
	}
}

func runPruneSnapshots(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	if err := rpcClient.Call(nil, methodName("prune")); err != nil {
		logFatal(err)
	}
}

func runAddSchedule(cmd *cobra.Command, args []string) {
	addressesHex, err := cmd.Flags().GetStringSlice("address")
	if err != nil {
		logFatal(err)
	}
	addresses := make([]common.Address, len(addressesHex))
	for _, addressHex := range addressesHex {
		addresses = append(addresses, common.HexToAddress(addressHex))
	}

	blockPeriod, err := cmd.Flags().GetUint("block-period")
	if err != nil {
		logFatal(err)
	}
	if blockPeriod == 0 {
		logFatal(errors.New("block period must be at least 1"))
	}

	replace, err := cmd.Flags().GetBool("replace")
	if err != nil {
		logFatal(err)
	}

	schedule := snapshot.Schedule{
		Addresses:   addresses,
		BlockPeriod: uint64(blockPeriod),
		Replace:     replace,
	}

	rpcClient := newRpcClient(cmd)
	if err := rpcClient.Call(nil, methodName("addSchedule"), schedule); err != nil {
		logFatal(err)
	}
}

func runDeleteSchedule(cmd *cobra.Command, args []string) {
	id, err := cmd.Flags().GetUint("id")
	if err != nil {
		logFatal(err)
	}

	rpcClient := newRpcClient(cmd)
	if err := rpcClient.Call(nil, methodName("deleteSchedule"), id); err != nil {
		logFatal(err)
	}
}

func printSnapshotMetadataWithStatus(metadata ...snapshot.SnapshotMetadataWithStatus) {
	for _, m := range metadata {
		jsonStr, err := json.Marshal(m)
		if err != nil {
			logFatal(err)
		}
		logInfo(string(jsonStr))
	}
}

func printSnapshotResponse(resp snapshot.SnapshotResponse) {
	jsonStr, err := json.Marshal(resp)
	if err != nil {
		logFatal(err)
	}
	logInfo(string(jsonStr))
}

func printSnapshotSchedules(schedules ...snapshot.Schedule) {
	for _, schedule := range schedules {
		jsonStr, err := json.Marshal(schedule)
		if err != nil {
			logFatal(err)
		}
		logInfo(string(jsonStr))
	}
}

func runSnapshotGet(cmd *cobra.Command, args []string) {
	addressHex, err := cmd.Flags().GetString("address")
	if err != nil {
		logFatal(err)
	}
	hasBlockHash := cmd.Flags().Changed("block-hash")
	blockHashHex, err := cmd.Flags().GetString("block-hash")
	if err != nil {
		logFatal(err)
	}

	address := common.HexToAddress(addressHex)

	var blockHash common.Hash
	if hasBlockHash {
		blockHash = common.HexToHash(blockHashHex)
	}

	rpcClient := newRpcClient(cmd)

	if hasBlockHash {
		var resp snapshot.SnapshotResponse
		if err := rpcClient.Call(&resp, methodName("get"), address, blockHash); err != nil {
			logFatal(err)
		}
		printSnapshotResponse(resp)
	} else {
		listAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			logFatal(err)
		}
		if listAll {
			var resp []snapshot.SnapshotMetadataWithStatus
			if err := rpcClient.Call(&resp, methodName("list"), address); err != nil {
				logFatal(err)
			}
			printSnapshotMetadataWithStatus(resp...)
		} else {
			var resp snapshot.SnapshotMetadataWithStatus
			if err := rpcClient.Call(&resp, methodName("last"), address); err != nil {
				logFatal(err)
			}
			printSnapshotMetadataWithStatus(resp)
		}
	}
}

func runSnapshotSchedulesGet(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	var resp []snapshot.Schedule
	err := rpcClient.Call(&resp, methodName("getSchedules"))
	if err != nil {
		logFatal(err)
	}
	for _, schedule := range resp {
		printSnapshotSchedules(schedule)
	}
}

// AddSnapshotCommand
func AddSnapshotCommand(parent *cobra.Command) {
	snapshotCmd := &cobra.Command{Use: "snapshot", Short: "manage integrated state snapshots in an execution engine"}
	snapshotCmd.PersistentFlags().StringP("rpc-url", "r", "http://localhost:8545", "rpc endpoint")
	snapshotCmd.PersistentFlags().StringP("jwt-secret", "", "", "jwt secret")

	var (
		newSnapshotCmd    = &cobra.Command{Use: "new", Short: "create a snapshot", Run: runNewSnapshot}
		deleteSnapshotCmd = &cobra.Command{Use: "delete", Short: "delete a snapshot", Run: runDeleteSnapshot}
		pruneSnapshotsCmd = &cobra.Command{Use: "prune", Short: "prune remote dandling snapshot data", Run: runPruneSnapshots}
		addScheduleCmd    = &cobra.Command{Use: "schedule", Short: "create a snapshot schedule", Run: runAddSchedule}
		deleteScheduleCmd = &cobra.Command{Use: "unschedule", Short: "delete a snapshot schedule", Run: runDeleteSchedule}
	)

	for _, cmd := range []*cobra.Command{newSnapshotCmd, deleteSnapshotCmd} {
		cmd.Flags().String("block-hash", "", "block hash")
		cmd.Flags().StringSlice("address", nil, "contract address(es)")
	}

	newSnapshotCmd.Flags().Bool("replace", false, "replace last snapshot")

	addScheduleCmd.Flags().StringSlice("address", nil, "contract address(es)")
	addScheduleCmd.Flags().UintP("block-period", "p", 32, "snapshot period in blocks")
	addScheduleCmd.Flags().Bool("replace", true, "replace last snapshot")

	deleteScheduleCmd.Flags().Uint("id", 0, "schedule id")

	snapshotCmd.AddCommand(newSnapshotCmd)
	snapshotCmd.AddCommand(deleteSnapshotCmd)
	snapshotCmd.AddCommand(pruneSnapshotsCmd)
	snapshotCmd.AddCommand(addScheduleCmd)
	snapshotCmd.AddCommand(deleteScheduleCmd)

	snapshotGetCmd := &cobra.Command{Use: "get", Short: "get a snapshot", Run: runSnapshotGet}
	snapshotGetCmd.PersistentFlags().String("address", "", "contract address")
	snapshotGetCmd.Flags().String("block-hash", "", "block hash")
	snapshotGetCmd.Flags().BoolP("all", "a", false, "list all snapshots")

	snapshotScheduleGet := &cobra.Command{Use: "schedules", Short: "get schedules", Run: runSnapshotSchedulesGet}
	snapshotScheduleGet.Flags().String("address", "", "contract address")
	snapshotGetCmd.AddCommand(snapshotScheduleGet)

	parent.AddCommand(snapshotCmd)
}
