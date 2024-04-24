package rpc

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/params"
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

	var header *types.Header

	// Wait for both goroutines to send a value or an error
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			return nil, nil, err
		case header = <-headerChan:
		case gasTipCap = <-gasTipCapChan:
		}
	}

	gasFeeCap = new(big.Int).Add(header.BaseFee, gasTipCap)

	return gasFeeCap, gasTipCap, nil
}

// ActionBatchSubscription is a subscription to action batches emitted by a core contract.
type ActionBatchSubscription struct {
	ethcli            EthCli
	actionSpecs       arch.ActionSpecs
	coreAddress       common.Address
	actionBatchesChan chan<- arch.ActionBatch
	txHashesChan      chan<- common.Hash
	unsubChan         chan struct{}
	errChan           chan error
	closeUnsubOnce    sync.Once
	closeErrOnce      sync.Once
	unsubscribed      bool
}

var _ ethereum.Subscription = (*ActionBatchSubscription)(nil)

// SubscribeActionBatches subscribes to action batches emitted by the core contract at coreAddress.
func SubscribeActionBatches(
	ethcli EthCli,
	actionSpecs arch.ActionSpecs,
	coreAddress common.Address,
	startingBlockNumber uint64,
	actionBatchesChan chan<- arch.ActionBatch,
	txHashesChan chan<- common.Hash,
) *ActionBatchSubscription {
	sub := &ActionBatchSubscription{
		ethcli:            ethcli,
		actionSpecs:       actionSpecs,
		coreAddress:       coreAddress,
		actionBatchesChan: actionBatchesChan,
		txHashesChan:      txHashesChan,
		unsubChan:         make(chan struct{}),
		errChan:           make(chan error, 1),
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
		s.unsubscribe()
		return true
	default:
		return false
	}
}

func (s *ActionBatchSubscription) unsubscribe() {
	s.unsubscribed = true
	s.tryCloseUnsub()
	s.tryCloseErr()
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
		Topics:    [][]common.Hash{{params.ActionExecutedEventID}},
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
		case <-s.unsubChan:
			s.unsubscribe()
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
	actions := make([]arch.Action, 0, len(logBatch))
	for _, log := range logBatch {
		action, err := s.actionSpecs.LogToAction(log)
		if err != nil {
			return err
		}
		actions = append(actions, action)
	}
	actionBatch := arch.NewActionBatch(blockNumber, actions)
	select {
	case <-s.unsubChan:
		s.unsubscribe()
		return nil
	case s.actionBatchesChan <- actionBatch:
		if s.txHashesChan != nil {
			for _, log := range logBatch {
				select {
				case <-s.unsubChan:
					s.unsubscribe()
					return nil
				case s.txHashesChan <- log.TxHash:
				}
			}
		}
	}
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
	actionSpecs     arch.ActionSpecs
	gasEstimator    ethereum.GasEstimator
	contractAddress common.Address
	from            common.Address
	nonce           uint64
	signerFn        bind.SignerFn
}

// TODO: Make these methods of action specs
// TODO: Rename action specs to something more self-explanatory

// NewActionSender creates a new ActionSender.
func NewActionSender(
	ethcli EthCli,
	actionSpecs arch.ActionSpecs,
	gasEstimator ethereum.GasEstimator,
	contractAddress common.Address,
	from common.Address,
	nonce uint64,
	signerFn bind.SignerFn,
) *ActionSender {
	if gasEstimator == nil {
		gasEstimator = ethcli
	}
	return &ActionSender{
		ethcli:          ethcli,
		actionSpecs:     actionSpecs,
		gasEstimator:    gasEstimator,
		contractAddress: contractAddress,
		from:            from,
		nonce:           nonce,
		signerFn:        signerFn,
	}
}

// packMultiActionCall packs multiple actions into a single call to the contract.
func (a *ActionSender) packMultiActionCall(actions []arch.Action) ([]byte, error) {
	var (
		actionIds   = make([]arch.RawIdType, 0)
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

// sendData sends a transaction to the contract with the given data.
func (a *ActionSender) sendData(data []byte) (*types.Transaction, error) {
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
	gasEstGasFeeCap := common.Big0
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
			return nil, err
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
	tx, err := a.signAndSend(txData)

	// Retry if nonce too low or too high
	if err != nil && (err.Error() == "nonce too low" || err.Error() == "nonce too high") {
		a.nonce, err = getPendingNonce(a.ethcli, a.from)
		if err != nil {
			return nil, err
		}
		txData.Nonce = a.nonce
		tx, err = a.signAndSend(txData)
	}

	// Increment nonce
	a.nonce++

	return tx, err
}

// signAndSend signs and sends a transaction.
func (a *ActionSender) signAndSend(txData types.TxData) (*types.Transaction, error) {
	tx := types.NewTx(txData)

	// Sign transaction
	signedTx, err := a.signerFn(a.from, tx)
	if err != nil {
		return nil, err
	}

	// Send transaction
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	if err := a.ethcli.SendTransaction(ctx, signedTx); err != nil {
		return nil, err
	}

	return signedTx, nil
}

// SendAction sends and action to the contract.
func (a *ActionSender) SendAction(action arch.Action) (*types.Transaction, error) {
	data, err := a.actionSpecs.ActionToCalldata(action)
	if err != nil {
		return nil, err
	}
	return a.sendData(data)
}

// SendActions sends multiple actions to the contract in a single transaction.
func (a *ActionSender) SendActions(actionBatch []arch.Action) (*types.Transaction, error) {
	if len(actionBatch) == 0 {
		return nil, nil
	} else if len(actionBatch) == 1 {
		return a.SendAction(actionBatch[0])
	} else {
		data, err := a.packMultiActionCall(actionBatch)
		if err != nil {
			return nil, err
		}
		return a.sendData(data)
	}
}

// StartSendingActions starts sending actions from the given channel.
func (a *ActionSender) StartSendingActions(actionsChan <-chan []arch.Action, txUpdateOutChan chan<- *ActionTxUpdate) (<-chan error, func()) {
	stopChan := make(chan struct{})
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case actions, ok := <-actionsChan:
				if !ok {
					return
				}
				// Copy nonce as it will be updated during SendActions
				nonce := a.nonce
				if txUpdateOutChan != nil {
					// Announce the actions before sending them
					txUpdateOutChan <- &ActionTxUpdate{Actions: actions, Nonce: nonce, Status: ActionTxStatus_Unsent}
				}
				tx, err := a.SendActions(actions)
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
				}
				if txUpdateOutChan != nil {
					if err != nil {
						// Announce failure
						txUpdateOutChan <- &ActionTxUpdate{Nonce: nonce, Status: ActionTxStatus_Failed, Err: err}
					} else {
						// Announce success
						txUpdateOutChan <- &ActionTxUpdate{TxHash: tx.Hash(), Nonce: nonce, Status: ActionTxStatus_Pending}
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

// TableGetter reads a table from the core contract.
type TableGetter struct {
	ethcli          EthCli
	tableSpecs      arch.TableSpecs
	contractAddress common.Address
}

// NewTableReader creates a new TableGetter.
func NewTableReader(
	ethcli EthCli,
	tableSpecs arch.TableSpecs,
	coreAddress common.Address,
) *TableGetter {
	return &TableGetter{
		ethcli:          ethcli,
		tableSpecs:      tableSpecs,
		contractAddress: coreAddress,
	}
}

// TODO: error message
// TODO: check assumptions about method signatures

// ReadTable reads a table from the contract.
func (t *TableGetter) Read(tableName string, keys ...interface{}) (interface{}, error) {
	// Get table ID from table name
	tableId, ok := t.tableSpecs.TableIdFromName(tableName)
	if !ok {
		return nil, errors.New("table name does not match any table")
	}

	// Get table schema from table ID
	schema := t.tableSpecs.GetTableSchema(tableId)
	data, err := t.tableSpecs.ABI().Pack(schema.Method.Name, keys...)
	if err != nil {
		return nil, err
	}

	// Call contract
	msg := ethereum.CallMsg{
		To:   &t.contractAddress,
		Data: data,
	}
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	result, err := t.ethcli.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	// Unpack result
	_ret, err := schema.Method.Outputs.Unpack(result)
	if err != nil {
		return nil, err
	}
	ret := _ret[0]

	// Convert result to canonical type
	row := reflect.New(schema.Type).Interface()
	if err := arch.ConvertStruct(ret, row); err != nil {
		return nil, err
	}

	return row, nil
}

type TxMonitor struct {
	client     EthCli
	txHashes   map[common.Hash]struct{}
	timestamps map[common.Hash]int64
}

func NewTxMonitor(ethcli EthCli) *TxMonitor {
	return &TxMonitor{
		client:     ethcli,
		txHashes:   make(map[common.Hash]struct{}),
		timestamps: make(map[common.Hash]int64),
	}
}

func (txm *TxMonitor) AddTxHash(txHash common.Hash) {
	txm.txHashes[txHash] = struct{}{}
	txm.timestamps[txHash] = time.Now().Unix()
}

func (txm *TxMonitor) RemoveTx(txHash common.Hash) {
	delete(txm.txHashes, txHash)
	delete(txm.timestamps, txHash)
}

func (txm *TxMonitor) HasTx(txHash common.Hash) bool {
	_, ok := txm.txHashes[txHash]
	return ok
}

func (txm *TxMonitor) IsPending(tx *types.Transaction) bool {
	isPending, _ := txm.isPending(tx.Hash())
	return isPending
}

func (txm *TxMonitor) isPending(txHash common.Hash) (bool, error) {
	_, isPending, err := txm.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return false, err
	}
	return isPending, nil
}

func (txm *TxMonitor) Update() bool {
	modified := false
	for txHash := range txm.txHashes {
		isPending, err := txm.isPending(txHash)
		if err != nil {
			// Remove a transaction that cannot be retrieved if it was added more than 6 seconds ago
			isStale := time.Now().Unix()-txm.timestamps[txHash] > 6
			if isStale {
				txm.RemoveTx(txHash)
				modified = true
			} else {
				isPending = true
			}
		}
		if !isPending {
			txm.RemoveTx(txHash)
			modified = true
		}
	}
	return modified
}

func (txm *TxMonitor) PendingTxs() []common.Hash {
	pendingTxs := make([]common.Hash, 0, len(txm.txHashes))
	for txHash := range txm.txHashes {
		pendingTxs = append(pendingTxs, txHash)
	}
	return pendingTxs
}

func (txm *TxMonitor) PendingTxsCount() int {
	return len(txm.txHashes)
}

type ActionTxStatus uint8

const (
	ActionTxStatus_Unsent ActionTxStatus = iota
	ActionTxStatus_Pending
	ActionTxStatus_Included
	ActionTxStatus_Failed
)

func (c *ActionTxStatus) String() string {
	switch *c {
	case ActionTxStatus_Unsent:
		return "unsent"
	case ActionTxStatus_Pending:
		return "pending"
	case ActionTxStatus_Included:
		return "included"
	case ActionTxStatus_Failed:
		return "failed"
	default:
		return "unknown"
	}
}

type ActionTxUpdate struct {
	Actions []arch.Action
	TxHash  common.Hash
	Nonce   uint64
	Status  ActionTxStatus
	Err     error
}

type TxHinter struct {
	txm            *TxMonitor
	unsentActions  map[uint64][]arch.Action
	actions        map[common.Hash][]arch.Action
	hintNonce      uint64
	txUpdateInChan <-chan *ActionTxUpdate
	stopChan       chan struct{}
	mutex          sync.Mutex
}

func NewTxHinter(ethcli EthCli, txUpdateInChan <-chan *ActionTxUpdate) *TxHinter {
	return &TxHinter{
		txm:            NewTxMonitor(ethcli),
		unsentActions:  make(map[uint64][]arch.Action),
		actions:        make(map[common.Hash][]arch.Action),
		txUpdateInChan: txUpdateInChan,
		hintNonce:      1,
		stopChan:       make(chan struct{}),
	}
}

func (txh *TxHinter) GetHints() (uint64, [][]arch.Action) {
	txh.mutex.Lock()
	defer txh.mutex.Unlock()

	pendingTxs := txh.txm.PendingTxs()
	hints := make([][]arch.Action, 0, len(pendingTxs))
	for _, txHash := range pendingTxs {
		hints = append(hints, txh.actions[txHash])
	}
	for _, action := range txh.unsentActions {
		hints = append(hints, action)
	}
	return txh.hintNonce, hints
}

func (txh *TxHinter) HintNonce() uint64 {
	return txh.hintNonce
}

func (txh *TxHinter) addAction(action *ActionTxUpdate) {
	txh.mutex.Lock()
	defer txh.mutex.Unlock()

	if action.Status == ActionTxStatus_Unsent {
		txh.unsentActions[action.Nonce] = action.Actions
	} else if action.Status == ActionTxStatus_Pending {
		actions := txh.unsentActions[action.Nonce]
		txh.actions[action.TxHash] = actions
		txh.txm.AddTxHash(action.TxHash)
		delete(txh.unsentActions, action.Nonce)
	} else if action.Status == ActionTxStatus_Failed {
		txh.txm.RemoveTx(action.TxHash)
		delete(txh.unsentActions, action.Nonce)
	} else if action.Status == ActionTxStatus_Included {
		txh.txm.RemoveTx(action.TxHash)
	}

	txh.hintNonce++
}

func (txh *TxHinter) Update() bool {
	txh.mutex.Lock()
	defer txh.mutex.Unlock()

	modified := txh.txm.Update()
	if modified {
		txh.hintNonce++
	}
	return modified
}

func (txh *TxHinter) Start(updateInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(updateInterval)
		for {
			select {
			case <-txh.stopChan:
				ticker.Stop()
				return
			case action := <-txh.txUpdateInChan:
				txh.addAction(action)
			case <-ticker.C:
				txh.Update()
			}
		}
	}()
}

func DampenLatency[T any](in chan T, out chan T, interval time.Duration, delay time.Duration) {
	go func() {
		// Send all until interval/2 has passed without any new items
		for sendUntilEmpty(in, out) {
			time.Sleep(interval / 2)
		}

		// Wait for and send next item
		item, ok := <-in
		if !ok {
			close(out)
			return
		}
		earliestExpectedTime := time.Now().Add(interval)
		time.Sleep(delay)
		out <- item

		for {
			// Send item at the right time
			item, ok := <-in
			if !ok {
				close(out)
				return
			}
			receivedTime := time.Now()
			time.Sleep(time.Until(earliestExpectedTime.Add(delay)))
			out <- item
			// Send all buffered items
			sendUntilEmpty(in, out)

			// Adjust the next expected time

			// If the item was received before or after the reception window, move the window closer to the received time
			// Window = [earliestExpectedTime, earliestExpectedTime + delay]
			// Latency deviation has a floor but no ceiling, so early blocks result in a bigger shift than late blocks.

			if receivedTime.Before(earliestExpectedTime) {
				// Received early
				earliestExpectedTime = midpoint(earliestExpectedTime, receivedTime, 0.5)
			} else if receivedTime.After(earliestExpectedTime.Add(delay)) {
				// Received late
				earliestExpectedTime = midpoint(earliestExpectedTime, receivedTime.Add(-delay), 0.25)
			}
			earliestExpectedTime = earliestExpectedTime.Add(interval)
		}
	}()
}

func midpoint(t1, t2 time.Time, f float64) time.Time {
	diff := t2.Sub(t1)
	midpoint := t1.Add(time.Duration(float64(diff) * f))
	return midpoint
}

func sendUntilEmpty[T any](in chan T, out chan T) bool {
	sent := false
	for {
		select {
		case item, ok := <-in:
			if !ok {
				close(out)
				return sent
			}
			out <- item
			sent = true
		default:
			return sent
		}
	}
}
