package rpc

import (
	"context"
	"errors"
	"math"
	"math/big"
	"sync"
	"time"

	archcodec "github.com/concrete-eth/archetype/codec"
	"github.com/concrete-eth/archetype/params"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

/*

- Subscribe to action batches: must be cancelable
- Send action batches in transactions

*/

var (
	StandardTimeout        = 5 * time.Second // Standard timeout for RPC requests
	BlockQueryLimit uint64 = 256             // Maximum number of blocks to query in a single request
	HeaderChanSize         = 4               // Size of the header channel
)

func getBlockNumber(ethcli EthCli) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	return ethcli.BlockNumber(ctx)
}

func getPendingNonce(ethcli EthCli, addr common.Address) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	return ethcli.PendingNonceAt(ctx, addr)
}

func suggestGasTipCap(ethcli EthCli) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	return ethcli.SuggestGasTipCap(ctx)
}

func getHeadHeader(ethcli EthCli) (*types.Header, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	return ethcli.HeaderByNumber(ctx, nil)
}

func getGasPrice(ethcli EthCli) (gasFeeCap, gasTipCap *big.Int, err error) {
	errChan := make(chan error, 2)
	headerChan := make(chan *types.Header, 1)
	gasTipCapChan := make(chan *big.Int, 1)

	go func() {
		header, err := getHeadHeader(ethcli)
		if err != nil {
			errChan <- err
			return
		}
		headerChan <- header
	}()

	go func() {
		gasTipCap, err := suggestGasTipCap(ethcli)
		if err != nil {
			errChan <- err
			return
		}
		gasTipCapChan <- gasTipCap
	}()

	// Wait for both goroutines to send a value or an error
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			return nil, nil, err
		case header := <-headerChan:
			gasFeeCap = new(big.Int).Add(header.BaseFee, <-gasTipCapChan)
		case gasTipCap = <-gasTipCapChan:
			// Wait to calculate gasFeeCap until header is received
		}
	}

	return gasFeeCap, gasTipCap, nil
}

type ActionBatchSubscription struct {
	ethcli          EthCli
	actionMap       archtypes.ActionMap
	actionAbi       abi.ABI
	coreAddress     common.Address
	actionBatchChan chan<- archtypes.ActionBatch
	unsubChan       chan struct{}
	errChan         chan error
	closeUnsubOnce  sync.Once
	closeErrOnce    sync.Once
	unsubscribed    bool
}

// SubscribeActionBatches subscribes to action batches emitted by the core contract at coreAddress.
func SubscribeActionBatches(
	ethcli EthCli,
	actionMap archtypes.ActionMap,
	actionAbi abi.ABI,
	coreAddress common.Address,
	startingBlockNumber uint64,
	actionBatchesChan chan<- archtypes.ActionBatch,
) *ActionBatchSubscription {
	sub := &ActionBatchSubscription{
		ethcli:          ethcli,
		actionMap:       actionMap,
		actionAbi:       actionAbi,
		coreAddress:     coreAddress,
		actionBatchChan: actionBatchesChan,
		unsubChan:       make(chan struct{}),
		errChan:         make(chan error, 1),
	}
	go sub.runSubscription(startingBlockNumber)
	return sub
}

func (s *ActionBatchSubscription) tryCloseUnsub() {
	s.closeUnsubOnce.Do(func() {
		close(s.unsubChan)
	})
}

func (s *ActionBatchSubscription) tryCloseErr() {
	s.closeErrOnce.Do(func() {
		close(s.errChan)
	})
}

func (s *ActionBatchSubscription) hasUnsubscribed() bool {
	if s.unsubscribed {
		return true
	}
	select {
	case <-s.unsubChan:
		s.unsubscribed = true
		s.tryCloseUnsub()
		s.tryCloseErr()
		return true
	default:
		return false
	}
}

func (s *ActionBatchSubscription) sendErr(err error) {
	s.errChan <- err
	s.tryCloseUnsub()
	s.tryCloseErr()
}

func (s *ActionBatchSubscription) runSubscription(startingBlock uint64) {
	if _, err := s.sync(startingBlock); err != nil {
		s.sendErr(err)
		return
	}
}

func (s *ActionBatchSubscription) getLogs(fromBlock, toBlock uint64) ([]types.Log, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{s.coreAddress},
		Topics:    [][]common.Hash{{s.actionAbi.Events[params.ActionExecutedEventName].ID}},
	}
	return s.ethcli.FilterLogs(ctx, query)
}

// sync sends an action batch for every block from startingBlock to the head block.
// When the head block is reached, a new batch is sent for every new block.
// sync will only return when the subscription is unsubscribed or an error occurs.
func (s *ActionBatchSubscription) sync(startingBlock uint64) (uint64, error) {
	var oldestUnsyncedBN uint64
	var err error
	if oldestUnsyncedBN, err = s.syncToHead(startingBlock); err != nil {
		return oldestUnsyncedBN, err
	}
	if oldestUnsyncedBN, err = s.syncAtHead(oldestUnsyncedBN); err != nil {
		return oldestUnsyncedBN, err
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) syncToHead(startingBlock uint64) (uint64, error) {
	if s.hasUnsubscribed() {
		return startingBlock, nil
	}
	oldestUnsyncedBN := startingBlock
	headBN, err := getBlockNumber(s.ethcli)
	if err != nil {
		return startingBlock, err
	}

	for oldestUnsyncedBN < headBN {
		if s.hasUnsubscribed() {
			return oldestUnsyncedBN, nil
		}
		var fromBN, toBN uint64
		{
			// Sync from oldestUnsyncedBN to oldestUnsyncedBN + BlockQueryLimit
			// If toBN > headBN, refetch headBN and set toBN = min(headBN, toBN)
			fromBN = oldestUnsyncedBN
			toBN = oldestUnsyncedBN + BlockQueryLimit
			if toBN > headBN {
				headBN, err := getBlockNumber(s.ethcli)
				if err != nil {
					return oldestUnsyncedBN, err
				}
				toBN = utils.Min(headBN, toBN)
			}
		}
		// Fetch logs from fromBN to toBN
		logs, err := s.getLogs(fromBN, toBN)
		if err != nil {
			return oldestUnsyncedBN, err
		}
		// Process logs
		if oldestUnsyncedBN, err = s.processLogs(logs, oldestUnsyncedBN, toBN); err != nil {
			return oldestUnsyncedBN, err
		}
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) syncAtHead(startingBlock uint64) (uint64, error) {
	if s.hasUnsubscribed() {
		return startingBlock, nil
	}
	oldestUnsyncedBN := startingBlock

	headerChan := make(chan *types.Header, HeaderChanSize)
	headersSub, err := s.ethcli.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		return oldestUnsyncedBN, err
	}
	defer headersSub.Unsubscribe()

	for {
		select {
		case err := <-headersSub.Err():
			return oldestUnsyncedBN, err
		case header := <-headerChan:
			if s.hasUnsubscribed() {
				return oldestUnsyncedBN, nil
			}
			if header.Number.Uint64() < oldestUnsyncedBN {
				continue
			}
			// Fetch logs from oldestUnsyncedBN to head
			logs, err := s.getLogs(oldestUnsyncedBN, header.Number.Uint64())
			if err != nil {
				return oldestUnsyncedBN, err
			}
			// Process logs
			if oldestUnsyncedBN, err = s.processLogs(logs, oldestUnsyncedBN, header.Number.Uint64()); err != nil {
				return oldestUnsyncedBN, err
			}
		}
	}
}

func (s *ActionBatchSubscription) processLogs(logs []types.Log, from, to uint64) (uint64, error) {
	oldestUnsyncedBN := from
	logBatch := make([]types.Log, 0)
	for _, log := range logs {
		for oldestUnsyncedBN < log.BlockNumber {
			if s.hasUnsubscribed() {
				return oldestUnsyncedBN, nil
			}
			if err := s.sendLogBatch(oldestUnsyncedBN, logBatch); err != nil {
				return oldestUnsyncedBN, err
			}
			logBatch = make([]types.Log, 0)
			oldestUnsyncedBN++
		}
		logBatch = append(logBatch, log)
	}
	for oldestUnsyncedBN <= to {
		if err := s.sendLogBatch(oldestUnsyncedBN, logBatch); err != nil {
			return oldestUnsyncedBN, err
		}
		logBatch = make([]types.Log, 0)
		oldestUnsyncedBN++
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) sendLogBatch(blockNumber uint64, logBatch []types.Log) error {
	// Process logBatch into action batch and send
	actions := make([]archtypes.Action, 0, len(logBatch))
	for _, log := range logBatch {
		action, err := archcodec.LogToAction(s.actionAbi, s.actionMap, log)
		if err != nil {
			return err
		}
		actions = append(actions, action)
	}
	actionBatch := archtypes.NewActionBatch(blockNumber, actions)
	s.actionBatchChan <- actionBatch
	return nil
}

func (s *ActionBatchSubscription) Unsubscribe() {
	s.unsubChan <- struct{}{}
}

func (s *ActionBatchSubscription) Err() <-chan error {
	return s.errChan
}

var _ ethereum.Subscription = (*ActionBatchSubscription)(nil)

type ActionSender struct {
	ethcli             EthCli
	actionMap          archtypes.ActionMap
	actionAbi          abi.ABI
	actionIdFromAction func(action interface{}) (archtypes.RawIdType, bool)
	gasEstimator       ethereum.GasEstimator
	coreAddress        common.Address
	from               common.Address
	nonce              uint64
	signerFn           bind.SignerFn
}

func NewActionSender(
	ethcli EthCli,
	actionMap archtypes.ActionMap,
	actionAbi abi.ABI,
	actionIdFromAction func(action interface{}) (archtypes.RawIdType, bool),
	gasEstimator ethereum.GasEstimator,
	coreAddress common.Address,
	from common.Address,
	nonce uint64,
	signerFn bind.SignerFn,
) *ActionSender {
	return &ActionSender{
		ethcli:             ethcli,
		actionMap:          actionMap,
		actionAbi:          actionAbi,
		actionIdFromAction: actionIdFromAction,
		gasEstimator:       gasEstimator,
		coreAddress:        coreAddress,
		from:               from,
		nonce:              nonce,
		signerFn:           signerFn,
	}
}

func (a *ActionSender) encodeAction(action archtypes.Action) (archtypes.RawIdType, []byte, error) {
	actionId, ok := a.actionIdFromAction(action)
	if !ok {
		return 0, nil, errors.New("unknown action ID")
	}
	actionMetadata := a.actionMap[actionId]
	method := a.actionAbi.Methods[actionMetadata.MethodName]
	data, err := method.Inputs.Pack(action)
	if err != nil {
		return 0, nil, err
	}
	return actionId, data, nil
}

func (a *ActionSender) packActionCall(action archtypes.Action) ([]byte, error) {
	actionId, ok := a.actionIdFromAction(action)
	if !ok {
		return nil, errors.New("unknown action ID")
	}
	actionMetadata := a.actionMap[actionId]
	data, err := a.actionAbi.Pack(actionMetadata.MethodName, action)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *ActionSender) packMultiActionCall(actions []archtypes.Action) ([]byte, error) {
	var (
		actionIds   = make([]archtypes.RawIdType, 0)
		actionCount = make([]uint8, 0)
		actionData  = make([]interface{}, 0, len(actions))
	)
	if len(actions) == 0 {
		return a.actionAbi.Pack(params.MultiActionMethodName, actionIds, actionCount, actionData)
	}

	firstActionId, firstData, err := a.encodeAction(actions[0])
	if err != nil {
		return nil, err
	}
	actionIds = append(actionIds, firstActionId)
	actionCount = append(actionCount, 1)
	actionData = append(actionData, firstData)

	for _, action := range actions[1:] {
		actionId, data, err := a.encodeAction(action)
		if err != nil {
			return nil, err
		}
		actionData = append(actionData, data)
		if actionId == actionIds[len(actionIds)-1] {
			actionCount[len(actionCount)-1]++
		} else {
			actionIds = append(actionIds, actionId)
			actionCount = append(actionCount, 1)
		}
	}

	return a.actionAbi.Pack(params.MultiActionMethodName, actionIds, actionCount, actionData)
}

func (a *ActionSender) sendData(data []byte) error {
	errChan := make(chan error, 2)
	gasPriceChan := make(chan [2]*big.Int, 1)
	estGasCostChan := make(chan uint64, 1)

	// Get gas price concurrently
	go func() {
		gasFeeCap, gasTipCap, err := getGasPrice(a.ethcli)
		if err != nil {
			errChan <- err
			return
		}
		gasPriceChan <- [2]*big.Int{gasFeeCap, gasTipCap}
	}()

	// Use provisional gas price to estimate gas
	gasEstGasFeeCap := new(big.Int).SetUint64(math.MaxUint64)
	gasEstTipCap := common.Big0

	msg := ethereum.CallMsg{
		From:      a.from,
		To:        &a.coreAddress,
		Value:     common.Big0,
		GasFeeCap: gasEstGasFeeCap,
		GasTipCap: gasEstTipCap,
		Data:      data,
	}

	// Estimate gas concurrently
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
		defer cancel()
		estimatedGas, err := a.gasEstimator.EstimateGas(ctx, msg)
		if err != nil {
			errChan <- err
			return
		}
		estGasCostChan <- estimatedGas
	}()

	// Wait for gas price response
	var gasFeeCap, gasTipCap *big.Int
	var estGasCost uint64

	// Wait for both goroutines to send a value or an error
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			return err
		case gasPrices := <-gasPriceChan:
			gasFeeCap, gasTipCap = gasPrices[0], gasPrices[1]
		case estGasCost = <-estGasCostChan:
		}
	}

	gasLimit := estGasCost + estGasCost/4

	// Send transaction
	txData := &types.DynamicFeeTx{
		Nonce:     a.nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		To:        msg.To,
		Value:     msg.Value,
		Data:      msg.Data,
	}
	err := a.signAndSend(txData)

	// Retry if nonce too low or too high
	if err.Error() == "nonce too low" || err.Error() == "nonce too high" {
		a.nonce, err = getPendingNonce(a.ethcli, a.from)
		if err != nil {
			return err
		}
		txData.Nonce = a.nonce
		err = a.signAndSend(txData)
	}

	// Increment nonce
	a.nonce++

	return err
}

func (a *ActionSender) signAndSend(txData types.TxData) error {
	tx := types.NewTx(txData)

	// Sign transaction
	signedTx, err := a.signerFn(a.from, tx)
	if err != nil {
		return err
	}

	// Send transaction
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	if err := a.ethcli.SendTransaction(ctx, signedTx); err != nil {
		return err
	}

	return nil
}

func (a *ActionSender) SendAction(action archtypes.Action) error {
	data, err := a.packActionCall(action)
	if err != nil {
		return err
	}
	return a.sendData(data)
}

func (a *ActionSender) SendActions(actionBatch []archtypes.Action) error {
	if len(actionBatch) == 0 {
		return nil
	} else if len(actionBatch) == 1 {
		return a.SendAction(actionBatch[0])
	} else {
		data, err := a.packMultiActionCall(actionBatch)
		if err != nil {
			return err
		}
		return a.sendData(data)
	}
}

func (a *ActionSender) StartSendingActions(actionsChan <-chan []archtypes.Action) (<-chan error, func()) {
	stopChan := make(chan struct{})
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case actions := <-actionsChan:
				if err := a.SendActions(actions); err != nil {
					select {
					case errChan <- err:
					default:
					}
				}
			}
		}
	}()
	cancel := func() {
		close(stopChan)
	}
	return errChan, cancel
}
