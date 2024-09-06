package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/concrete-eth/archetype/example/engine"
	"github.com/concrete-eth/archetype/snapshot"
	"github.com/ethereum/go-ethereum/common"

	"github.com/spf13/cobra"
)

func checkIsHexAddress(addressHex string) {
	if !common.IsHexAddress(addressHex) {
		err := fmt.Errorf("'%s' is not a valid address", addressHex)
		logFatalNoContext(err)
	}
}

func methodName(name string) string {
	return engine.SnapshotNamespace + "_" + name
}

func getSnapshotQuery(cmd *cobra.Command) *snapshot.SnapshotQuery {
	blockHashHex, err := cmd.Flags().GetString("block-hash")
	if err != nil {
		logFatal(err)
	}
	if blockHashHex == "" {
		logFatalNoContext(errors.New("block hash is required"))
	}
	addressesHex, err := cmd.Flags().GetStringSlice("address")
	if err != nil {
		logFatal(err)
	}
	if len(addressesHex) == 0 {
		logFatalNoContext(errors.New("at least one address is required"))
	}

	blockHash := common.HexToHash(blockHashHex)
	addresses := make([]common.Address, len(addressesHex))

	for idx, addressHex := range addressesHex {
		checkIsHexAddress(addressHex)
		addresses[idx] = common.HexToAddress(addressHex)
	}

	return &snapshot.SnapshotQuery{
		BlockHash: blockHash,
		Addresses: addresses,
	}
}

func runNewSnapshot(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	query := getSnapshotQuery(cmd)
	replace, err := cmd.Flags().GetBool("replace")
	if err != nil {
		logFatal(err)
	}
	var resp []snapshot.SnapshotMetadataWithStatus
	if replace {
		if err := rpcClient.Call(&resp, methodName("update"), query); err != nil {
			logFatalNoContext(err)
		}
	} else {
		if err := rpcClient.Call(&resp, methodName("new"), query); err != nil {
			logFatalNoContext(err)
		}
	}
	printSnapshotMetadataWithStatus(resp...)
}

func runDeleteSnapshot(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	query := getSnapshotQuery(cmd)
	if err := rpcClient.Call(nil, methodName("delete"), query); err != nil {
		logFatalNoContext(err)
	}
}

func runPruneSnapshots(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	if err := rpcClient.Call(nil, methodName("prune")); err != nil {
		logFatalNoContext(err)
	}
}

func runAddSchedule(cmd *cobra.Command, args []string) {
	addressesHex, err := cmd.Flags().GetStringSlice("address")
	if err != nil {
		logFatal(err)
	}

	if len(addressesHex) == 0 {
		logFatalNoContext(errors.New("at least one address is required"))
	}

	addresses := make([]common.Address, len(addressesHex))
	for _, addressHex := range addressesHex {
		checkIsHexAddress(addressHex)
		addresses = append(addresses, common.HexToAddress(addressHex))
	}

	blockPeriod, err := cmd.Flags().GetUint("block-period")
	if err != nil {
		logFatal(err)
	}
	if blockPeriod == 0 {
		logFatalNoContext(errors.New("block period must be at least 1"))
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
		logFatalNoContext(err)
	}
}

func runDeleteSchedule(cmd *cobra.Command, args []string) {
	id, err := cmd.Flags().GetUint("id")
	if err != nil {
		logFatal(err)
	}

	rpcClient := newRpcClient(cmd)
	if err := rpcClient.Call(nil, methodName("deleteSchedule"), id); err != nil {
		logFatalNoContext(err)
	}
}

func printSnapshotMetadataWithStatus(metadata ...snapshot.SnapshotMetadataWithStatus) {
	var data any
	if len(metadata) == 1 {
		data = metadata[0]
	} else {
		data = metadata
	}
	jsonStr, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		logFatal(err)
	}
	logInfo(string(jsonStr))
}

func printSnapshotResponse(resp snapshot.SnapshotResponse) {
	jsonStr, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		logFatal(err)
	}
	logInfo(string(jsonStr))
}

func printSnapshotSchedules(schedules map[uint64]snapshot.Schedule) {
	data := schedules
	jsonStr, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		logFatal(err)
	}
	logInfo(string(jsonStr))
}

func runGetSnapshot(cmd *cobra.Command, args []string) {
	address := getAddress(cmd)
	hasBlockHash := cmd.Flags().Changed("block-hash")

	blockHashHex, err := cmd.Flags().GetString("block-hash")
	if err != nil {
		logFatal(err)
	}
	blockHash := common.HexToHash(blockHashHex)

	rpcClient := newRpcClient(cmd)

	if hasBlockHash {
		var resp snapshot.SnapshotResponse
		if err := rpcClient.Call(&resp, methodName("get"), address, blockHash); err != nil {
			logFatalNoContext(err)
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
				logFatalNoContext(err)
			}
			printSnapshotMetadataWithStatus(resp...)
		} else {
			var resp snapshot.SnapshotMetadataWithStatus
			if err := rpcClient.Call(&resp, methodName("last"), address); err != nil {
				logFatalNoContext(err)
			}
			printSnapshotMetadataWithStatus(resp)
		}
	}
}

func runGetSnapshotSchedules(cmd *cobra.Command, args []string) {
	rpcClient := newRpcClient(cmd)
	var resp map[uint64]snapshot.Schedule
	if err := rpcClient.Call(&resp, methodName("getSchedules")); err != nil {
		logFatalNoContext(err)
	}
	printSnapshotSchedules(resp)
}

// AddSnapshotCommand
func AddSnapshotCommand(parent *cobra.Command) {
	snapshotCmd := &cobra.Command{Use: "snapshot", Short: "Manage integrated state snapshots in an execution engine"}
	addRpcFlags(snapshotCmd)

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

	snapshotGetCmd := &cobra.Command{Use: "get", Short: "get a snapshot", Run: runGetSnapshot}
	snapshotGetCmd.PersistentFlags().String("address", "", "contract address")
	snapshotGetCmd.Flags().String("block-hash", "", "block hash")
	snapshotGetCmd.Flags().BoolP("all", "a", false, "list all snapshots")

	snapshotScheduleGet := &cobra.Command{Use: "schedules", Short: "get schedules", Run: runGetSnapshotSchedules}
	snapshotGetCmd.AddCommand(snapshotScheduleGet)

	snapshotCmd.AddCommand(snapshotGetCmd)

	parent.AddCommand(snapshotCmd)
}
