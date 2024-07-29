package e2e

import (
	"path/filepath"
	"testing"

	"github.com/concrete-eth/archetype/example/engine"
	"github.com/ethereum/go-ethereum/concrete/testtool"
)

func TestE2E(t *testing.T) {
	registry := engine.NewRegistry()
	config := testtool.TestConfig{
		Contract: filepath.Join("Test.sol:Test"),
		OutDir:   filepath.Join("..", "out"),
	}
	testtool.Test(t, registry, config)
}
