// (c) 2022 Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package params

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile"
	"github.com/ava-labs/subnet-evm/utils"
)

// Upgrade is a helper struct embedded in UpgradesConfig, representing
// each of the possible stateful precompile types that can be activated
// through UpgradesConfig.
type Upgrade struct {
	ContractDeployerAllowListConfig *precompile.ContractDeployerAllowListConfig `json:"contractDeployerAllowListConfig,omitempty"` // Config for the contract deployer allow list precompile
	ContractNativeMinterConfig      *precompile.ContractNativeMinterConfig      `json:"contractNativeMinterConfig,omitempty"`      // Config for the native minter precompile
	TxAllowListConfig               *precompile.TxAllowListConfig               `json:"txAllowListConfig,omitempty"`               // Config for the tx allow list precompile
	FeeManagerConfig                *precompile.FeeConfigManagerConfig          `json:"feeManagerConfig,omitempty"`                // Config for the fee manager precompile
}

type getterFn func(*Upgrade) precompile.StatefulPrecompileConfig

// getters contains a list of functions that return each of the stateful precompile
// types that can be retrieved from an upgrade.
var getters = []getterFn{
	func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.ContractDeployerAllowListConfig },
	func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.ContractNativeMinterConfig },
	func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.TxAllowListConfig },
	func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.FeeManagerConfig },
}

// TODO: can we do better?
func isNil(s precompile.StatefulPrecompileConfig) bool {
	switch s := s.(type) {
	case *precompile.ContractDeployerAllowListConfig:
		return s == nil
	case *precompile.ContractNativeMinterConfig:
		return s == nil
	case *precompile.TxAllowListConfig:
		return s == nil
	case *precompile.FeeConfigManagerConfig:
		return s == nil
	}
	panic("unknown type of StatefulPrecompileConfig")
}

// ValidatePrecompileUpgrades checks the PrecompileUpgrades is well formed:
// - PrecompileUpgrades must specify only one key per Upgrade
// - the specified blockTimestamps must monotonically increase
// - the specified blockTimestamps must be compatible with those
//   specified in the chainConfig by genesis.
// - check a precompile is disabled before it is re-enabled
func (c *UpgradesConfig) ValidatePrecompileUpgrades() error {
	for i, upgrade := range c.PrecompileUpgrades {
		hasKey := false // used to verify if there is only one key per Upgrade

		for _, getter := range getters {
			if config := getter(&upgrade); isNil(config) {
				continue
			}
			if hasKey {
				return fmt.Errorf("PrecompileUpgrades[%d] has more than one key set", i)
			}
			hasKey = true
		}
	}

	for _, getter := range getters {
		var (
			lastUpgraded *big.Int
			disabled     bool
		)
		// check the genesis chain config for any enabled upgrade
		if config := getter(&c.Upgrade); !isNil(config) {
			disabled = false
			lastUpgraded = config.Timestamp()
		} else {
			disabled = true
		}
		// next range over UpgradesFromUpgradeBytes
		// to verify correct use of disabled and blockTimestamps.
		for i, upgrade := range c.PrecompileUpgrades {
			if config := getter(&upgrade); !isNil(config) {
				if disabled == config.IsDisabled() {
					return fmt.Errorf("PrecompileUpgrades[%d] disable should be [%v]", i, !disabled)
				}
				if lastUpgraded != nil && (config.Timestamp() == nil || lastUpgraded.Cmp(config.Timestamp()) > 0) {
					return fmt.Errorf("PrecompileUpgrades[%d] timestamp should not be less than [%v]", i, lastUpgraded)
				}

				disabled = config.IsDisabled()
				lastUpgraded = config.Timestamp()
			}
		}
	}

	return nil // successfully verified all conditions
}

// getActiveConfig returns the most recent config that has
// already forked, or nil if none have been configured or
// if none have forked yet.
func (c *UpgradesConfig) getActiveConfig(blockTimestamp *big.Int, getter getterFn) precompile.StatefulPrecompileConfig {
	configs := c.getActiveConfigs(nil, blockTimestamp, getter)
	if len(configs) == 0 {
		return nil
	}
	return configs[len(configs)-1] // return the most recent config
}

// getActiveConfigs returns all forks configured to activate in the
// specified [from, to] range, in the order of activation.
func (c *UpgradesConfig) getActiveConfigs(from *big.Int, to *big.Int, getter getterFn) []precompile.StatefulPrecompileConfig {
	configs := make([]precompile.StatefulPrecompileConfig, 0)
	// first check the embedded [upgrade] for precompiles configured
	// in the genesis chain config.
	if config := getter(&c.Upgrade); !isNil(config) {
		if utils.IsForkTransition(config.Timestamp(), from, to) {
			configs = append(configs, config)
		}
	}
	// loop on all upgrades
	for _, upgrade := range c.PrecompileUpgrades {
		if config := getter(&upgrade); !isNil(config) {
			// check if fork is activating in the specified range
			if utils.IsForkTransition(config.Timestamp(), from, to) {
				configs = append(configs, config)
			}
		}
	}
	return configs
}

// GetContractDeployerAllowListConfig returns the latest forked ContractDeployerAllowListConfig
// specified by [c] or nil if it was never enabled.
func (c *UpgradesConfig) GetContractDeployerAllowListConfig(blockTimestamp *big.Int) *precompile.ContractDeployerAllowListConfig {
	getter := func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.ContractDeployerAllowListConfig }
	if val := c.getActiveConfig(blockTimestamp, getter); val != nil {
		return val.(*precompile.ContractDeployerAllowListConfig)
	}
	return nil
}

// GetContractNativeMinterConfig returns the latest forked ContractNativeMinterConfig
// specified by [c] or nil if it was never enabled.
func (c *UpgradesConfig) GetContractNativeMinterConfig(blockTimestamp *big.Int) *precompile.ContractNativeMinterConfig {
	getter := func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.ContractNativeMinterConfig }
	if val := c.getActiveConfig(blockTimestamp, getter); val != nil {
		return val.(*precompile.ContractNativeMinterConfig)
	}
	return nil
}

// GetTxAllowListConfig returns the latest forked TxAllowListConfig
// specified by [c] or nil if it was never enabled.
func (c *UpgradesConfig) GetTxAllowListConfig(blockTimestamp *big.Int) *precompile.TxAllowListConfig {
	getter := func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.TxAllowListConfig }
	if val := c.getActiveConfig(blockTimestamp, getter); val != nil {
		return val.(*precompile.TxAllowListConfig)
	}
	return nil
}

// GetFeeConfigManagerConfig returns the latest forked FeeManagerConfig
// specified by [c] or nil if it was never enabled.
func (c *UpgradesConfig) GetFeeConfigManagerConfig(blockTimestamp *big.Int) *precompile.FeeConfigManagerConfig {
	getter := func(u *Upgrade) precompile.StatefulPrecompileConfig { return u.FeeManagerConfig }
	if val := c.getActiveConfig(blockTimestamp, getter); val != nil {
		return val.(*precompile.FeeConfigManagerConfig)
	}
	return nil
}

// CheckCompatible checks if [newcfg] is compatible with [c] at [headTimestamp].
// Returns a ConfigCompatError if upgrades that have forked at [headTimestamp]
// are missing from [newcfg]. Upgrades that have not forked yet may be modified
// or absent from [newcfg]. Returns nil if [newcfg] is compatible with [c].
func (c *UpgradesConfig) CheckCompatible(newcfg *UpgradesConfig, headTimestamp *big.Int) *ConfigCompatError {
	for _, getter := range getters {
		// all active upgrades must match
		newUpgrades := newcfg.getActiveConfigs(nil, headTimestamp, getter)
		activeUpgrades := c.getActiveConfigs(nil, headTimestamp, getter)
		// first, check existing upgrades are there
		for i, upgrade := range activeUpgrades {
			if len(newUpgrades) <= i {
				// missing upgrade
				return newCompatError(
					fmt.Sprintf("missing PrecompileUpgrade[%d]", i),
					upgrade.Timestamp(),
					nil,
				)
			}
			// All upgrades that have forked must be identical.
			// TODO: verify config?
			if upgrade.Timestamp().Cmp(newUpgrades[i].Timestamp()) != 0 ||
				upgrade.IsDisabled() != newUpgrades[i].IsDisabled() {
				return newCompatError(
					fmt.Sprintf("PrecompileUpgrade[%d]", i),
					upgrade.Timestamp(),
					newUpgrades[i].Timestamp(),
				)
			}
		}
		// then, make sure newUpgrades does not have additional upgrades
		// that are already activated. (cannot perform retroactive upgrade)
		if len(newUpgrades) > len(activeUpgrades) {
			return newCompatError(
				fmt.Sprintf("cannot retroactively enable PrecompileUpgrade[%d]", len(activeUpgrades)),
				nil,
				newUpgrades[len(activeUpgrades)].Timestamp(), // this indexes to the first element in newUpgrades after the end of activeUpgrades
			)
		}
	}
	return nil // newcfg is compatible
}

// EnabledStatefulPrecompiles returns a slice of stateful precompile configs that
// have been activated through an upgrade.
func (c *UpgradesConfig) EnabledStatefulPrecompiles(blockTimestamp *big.Int) []precompile.StatefulPrecompileConfig {
	statefulPrecompileConfigs := make([]precompile.StatefulPrecompileConfig, 0)
	for _, getter := range getters {
		if config := c.getActiveConfig(blockTimestamp, getter); config != nil {
			statefulPrecompileConfigs = append(statefulPrecompileConfigs, config)
		}
	}

	return statefulPrecompileConfigs
}

// CheckConfigure checks if any of the precompiles specified by [c] is enabled or disabled by the block transition
// from [parentTimestamp] to the timestamp set in [blockContext]. If this is the case, it calls [Configure] or
// [Deconfigure] to apply the necessary state transitions for the upgrade.
// This function is called:
// - within genesis setup to configure the starting state for precompiles enabled at genesis,
// - during block processing to update the state before processing the given block.
func (c *UpgradesConfig) CheckConfigure(chainConfig precompile.ChainConfig, parentTimestamp *big.Int, blockContext precompile.BlockContext, statedb precompile.StateDB) {
	blockTimestamp := blockContext.Timestamp()
	for _, getter := range getters {
		for _, config := range c.getActiveConfigs(parentTimestamp, blockTimestamp, getter) {
			// If this transition activates the upgrade, configure the stateful precompile.
			// (or deconfigure it if it is being disabled.)
			if config.IsDisabled() {
				precompile.Deconfigure(config.Address(), statedb)
			} else {
				precompile.Configure(chainConfig, blockContext, config, statedb)
			}
		}
	}
}