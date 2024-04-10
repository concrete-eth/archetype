package rpc

import (
	"context"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/params"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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
	// Start two goroutines to get the head header and suggested gas tip cap concurrently

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

// ActionBatchSubscription is a subscription to action batches emitted by a core contract.
type ActionBatchSubscription struct {
	ethcli          EthCli
	actionSpecs     archtypes.ActionSpecs
	coreAddress     common.Address
	actionBatchChan chan<- archtypes.ActionBatch
	unsubChan       chan struct{}
	errChan         chan error
	closeUnsubOnce  sync.Once
	closeErrOnce    sync.Once
	unsubscribed    bool
}

var _ ethereum.Subscription = (*ActionBatchSubscription)(nil)

// SubscribeActionBatches subscribes to action batches emitted by the core contract at coreAddress.
func SubscribeActionBatches(
	ethcli EthCli,
	actionSpecs archtypes.ActionSpecs,
	coreAddress common.Address,
	startingBlockNumber uint64,
	actionBatchesChan chan<- archtypes.ActionBatch,
) *ActionBatchSubscription {
	sub := &ActionBatchSubscription{
		ethcli:          ethcli,
		actionSpecs:     actionSpecs,
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
		Topics:    [][]common.Hash{{archtypes.ActionExecutedEvent.ID}},
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
		action, err := s.actionSpecs.LogToAction(log)
		if err != nil {
			return err
		}
		actions = append(actions, action)
	}
	actionBatch := archtypes.NewActionBatch(blockNumber, actions)
	s.actionBatchChan <- actionBatch
	return nil
}

// Unsubscribe unsubscribes from the action batch subscription and closes the error channel.
// It does not close the action batch channel.
func (s *ActionBatchSubscription) Unsubscribe() {
	s.unsubChan <- struct{}{}
}

// Err returns the subscription error channel. Only one value will ever be sent.
// The error channel is closed by Unsubscribe.
func (s *ActionBatchSubscription) Err() <-chan error {
	return s.errChan
}

// ActionSender sends actions to a core contract.
type ActionSender struct {
	ethcli          EthCli
	actionSpecs     archtypes.ActionSpecs
	gasEstimator    ethereum.GasEstimator
	contractAddress common.Address
	from            common.Address
	nonce           uint64
	signerFn        bind.SignerFn
}

// TODO: Make these methods of action specs
// TODO: Rename action specs to something more self-explanatory
// TODO: check constructor params

// NewActionSender creates a new ActionSender.
func NewActionSender(
	ethcli EthCli,
	actionSpecs archtypes.ActionSpecs,
	gasEstimator ethereum.GasEstimator,
	coreAddress common.Address,
	from common.Address,
	nonce uint64,
	signerFn bind.SignerFn,
) *ActionSender {
	// if gasEstimator == nil {
	// 	gasEstimator = ethcli
	// }
	return &ActionSender{
		ethcli:          ethcli,
		actionSpecs:     actionSpecs,
		gasEstimator:    gasEstimator,
		contractAddress: coreAddress,
		from:            from,
		nonce:           nonce,
		signerFn:        signerFn,
	}
}

// packMultiActionCall packs multiple actions into a single call to the contract.
func (a *ActionSender) packMultiActionCall(actions []archtypes.Action) ([]byte, error) {
	var (
		actionIds   = make([]archtypes.RawIdType, 0)
		actionCount = make([]uint8, 0)
		actionData  = make([][]byte, 0, len(actions))
	)
	if len(actions) == 0 {
		return a.actionSpecs.ABI().Pack(params.MultiActionMethodName, actionIds, actionCount, actionData)
	}

	firstActionId, firstData, err := a.actionSpecs.EncodeAction(actions[0])
	if err != nil {
		return nil, err
	}
	actionIds = append(actionIds, firstActionId.Raw())
	actionCount = append(actionCount, 1)
	actionData = append(actionData, firstData)

	for _, action := range actions[1:] {
		_actionId, data, err := a.actionSpecs.EncodeAction(action)
		if err != nil {
			return nil, err
		}
		rawActionId := _actionId.Raw()
		actionData = append(actionData, data)
		if rawActionId == actionIds[len(actionIds)-1] {
			actionCount[len(actionCount)-1]++
		} else {
			actionIds = append(actionIds, rawActionId)
			actionCount = append(actionCount, 1)
		}
	}

	return a.actionSpecs.ABI().Pack(params.MultiActionMethodName, actionIds, actionCount, actionData)
}

// sendData sends a transaction to the contract with the provided data.
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
		To:        &a.contractAddress,
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

// signAndSend signs and sends a transaction.
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

// SendAction sends and action to the contract.
func (a *ActionSender) SendAction(action archtypes.Action) error {
	data, err := a.actionSpecs.ActionToCalldata(action)
	if err != nil {
		return err
	}
	return a.sendData(data)
}

// SendActions sends multiple actions to the contract in a single transaction.
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

// StartSendingActions starts sending actions from the provided channel.
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
