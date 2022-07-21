// (c) 2022 Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package params

import (
	"math/big"
	"testing"

	"github.com/ava-labs/subnet-evm/precompile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestValidateWithChainConfig(t *testing.T) {
	admins := []common.Address{{1}}
	config := &UpgradesConfig{
		Upgrade: Upgrade{
			TxAllowListConfig: &precompile.TxAllowListConfig{
				UpgradeableConfig: precompile.UpgradeableConfig{
					BlockTimestamp: big.NewInt(2),
				},
			},
		},
	}
	config.PrecompileUpgrades = []Upgrade{
		{
			// disable TxAllowList at timestamp 4
			TxAllowListConfig: precompile.NewDisableTxAllowListConfig(big.NewInt(4)),
		},
		{
			// re-enable TxAllowList at timestamp 5
			TxAllowListConfig: precompile.NewTxAllowListConfig(big.NewInt(5), admins),
		},
	}

	// check this config is valid
	err := config.ValidatePrecompileUpgrades()
	assert.NoError(t, err)

	// entries must be monotonically increasing
	badConfig := *config
	badConfig.PrecompileUpgrades = append(
		badConfig.PrecompileUpgrades,
		Upgrade{
			TxAllowListConfig: precompile.NewDisableTxAllowListConfig(big.NewInt(1)),
		},
	)
	err = badConfig.ValidatePrecompileUpgrades()
	assert.ErrorContains(t, err, "timestamp should not be less than [5]")

	// cannot enable a precompile without disabling it first.
	badConfig = *config
	badConfig.PrecompileUpgrades = append(
		badConfig.PrecompileUpgrades,
		Upgrade{
			TxAllowListConfig: precompile.NewTxAllowListConfig(big.NewInt(5), admins),
		},
	)
	err = badConfig.ValidatePrecompileUpgrades()
	assert.ErrorContains(t, err, "disable should be [true]")
}

func TestValidate(t *testing.T) {
	admins := []common.Address{{1}}
	config := &UpgradesConfig{}
	config.PrecompileUpgrades = []Upgrade{
		{
			TxAllowListConfig: precompile.NewTxAllowListConfig(big.NewInt(1), admins),
		},
		{
			TxAllowListConfig: precompile.NewDisableTxAllowListConfig(big.NewInt(2)),
		},
	}

	// check this config is valid
	err := config.ValidatePrecompileUpgrades()
	assert.NoError(t, err)
}