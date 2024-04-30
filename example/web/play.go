//go:build js
// +build js

package main

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/client"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/rpc"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	game_contract "github.com/concrete-eth/archetype/example/gogen/abigen/game"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/hajimehoshi/ebiten/v2"
)

type URLParams struct {
	GameAddress    common.Address
	WsURL          string
	BlockTime      time.Duration
	DampeningDelay time.Duration
}

func getURLParams() (URLParams, error) {
	window := js.Global()
	href := window.Get("location").Get("href").String()

	parsedUrl, err := url.Parse(href)
	if err != nil {
		return URLParams{}, err
	}

	path := parsedUrl.Path
	segments := strings.Split(path, "/")
	if len(segments) != 3 {
		return URLParams{}, fmt.Errorf("invalid path")
	}

	// Game contract address
	gameAddressHex := segments[2]
	if gameAddressHex == "" {
		return URLParams{}, fmt.Errorf("address parameter is required")
	}
	gameAddress := common.HexToAddress(gameAddressHex)

	queryParams := parsedUrl.Query()
	var paramValue string

	// RPC URL
	paramValue = queryParams.Get("ws")
	if paramValue == "" {
		return URLParams{}, fmt.Errorf("ws parameter is required")
	}
	wsURL := paramValue

	// Block time
	paramValue = queryParams.Get("blockTime")
	var blockTimeDuration time.Duration
	if paramValue == "" {
		blockTimeDuration = 1 * time.Second
	} else {
		blockTime, err := strconv.Atoi(paramValue)
		if err != nil {
			return URLParams{}, fmt.Errorf("blockTime parameter is required")
		}
		blockTimeDuration = time.Duration(blockTime) * time.Millisecond
	}

	// Dampening delay
	paramValue = queryParams.Get("delay")
	var delayDuration time.Duration
	if paramValue == "" {
		delayDuration = 250 * time.Millisecond
	} else {
		delay, err := strconv.Atoi(paramValue)
		if err != nil {
			return URLParams{}, fmt.Errorf("delay parameter is required")
		}
		delayDuration = time.Duration(delay) * time.Millisecond
		if delayDuration < 0 {
			delayDuration = 0
		} else if delayDuration > blockTimeDuration {
			delayDuration = blockTimeDuration
		}
	}

	return URLParams{
		GameAddress:    gameAddress,
		WsURL:          wsURL,
		BlockTime:      blockTimeDuration,
		DampeningDelay: delayDuration,
	}, nil
}

func getPrivateKey() (string, error) {
	return "", nil
}

func setLoadStatus(status string) {
	window := js.Global()
	if element := window.Get("document").Call("getElementById", "loader-status-main"); !element.IsNull() {
		element.Set("innerText", status)
	}
}

func hideLoadStatus() {
	window := js.Global()
	if element := window.Get("document").Call("getElementById", "loader-container-main"); !element.IsNull() {
		element.Get("parentNode").Call("removeChild", element)
	}
}

func showErrorScreen(err error) {
	body := js.Global().Get("document").Call("getElementsByTagName", "body").Index(0)
	body.Set("innerHTML", `
        <div id="error-container-main" class="error-container">
            <div id="error-status-main" class="error-status">
				<h1>Error</h1>
				<p>`+err.Error()+`</p>
			</div>
        </div>
    `)
}

func logCrit(err error) {
	showErrorScreen(err)
	log.Error(err.Error())
	os.Exit(0)
}

func runGameClient(params URLParams, privateKeyHex string) {
	// Connect to rpc
	setLoadStatus("Connecting...")
	rpcClient, err := ethclient.Dial(params.WsURL)
	if err != nil {
		logCrit(fmt.Errorf("Failed to connect to RPC: %v", err))
	}
	log.Info("Connected to RPC", "url", params.WsURL)

	// Create signer
	chainId, err := rpcClient.ChainID(context.Background())
	if err != nil {
		logCrit(fmt.Errorf("Failed to get chain ID: %v", err))
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		logCrit(fmt.Errorf("Failed to parse private key: %v", err))
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		logCrit(fmt.Errorf("Failed to create transactor: %v", err))
	}
	log.Info("Loaded burner wallet", "address", opts.From)
	from := opts.From

	// Set nonce
	nonce, err := rpcClient.PendingNonceAt(context.Background(), from)
	if err != nil {
		panic(err)
	}
	opts.Nonce = new(big.Int).SetUint64(nonce)

	// Get core address
	gameContract, err := game_contract.NewContract(params.GameAddress, rpcClient)
	if err != nil {
		logCrit(err)
	}
	coreAddress, err := gameContract.Proxy(nil)
	if err != nil {
		logCrit(err)
	}

	// Create chain IO
	var (
		schemas             = arch.ArchSchemas{Actions: archmod.ActionSchemas, Tables: archmod.TableSchemas}
		blockTime           = params.BlockTime
		startingBlockNumber = uint64(0) // TODO
	)

	io := rpc.NewIO(rpcClient, blockTime, schemas, opts, params.GameAddress, coreAddress, startingBlockNumber, params.DampeningDelay)
	io.SetTxUpdateHook(func(txUpdate *rpc.ActionTxUpdate) {
		log.Info("Transaction "+txUpdate.Status.String(), "nonce", txUpdate.Nonce, "txHash", txUpdate.TxHash.Hex())
	})

	// Create and start client
	kv := kvstore.NewMemoryKeyValueStore()
	c := client.NewClient(kv, io)
	w, h := c.Layout(-1, -1)
	ebiten.SetWindowSize(w, h)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		logCrit(err)
	}
}

func main() {
	// Get URL params
	params, err := getURLParams()
	if err != nil {
		logCrit(fmt.Errorf("failed to get URL params: %v", err))
	}

	// Get private key
	privateKey, err := getPrivateKey()
	if err != nil {
		logCrit(fmt.Errorf("failed to get burner key: %v", err))
	}

	// Set log level
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	// Start
	log.Debug(
		"Starting game client",
		"gameAddress", params.GameAddress.Hex(),
		"wsURL", params.WsURL,
		"blockTime", params.BlockTime,
		"dampeningDelay", params.DampeningDelay,
	)
	runGameClient(params, privateKey)
}
