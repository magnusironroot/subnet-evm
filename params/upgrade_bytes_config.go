// (c) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package params

import (
	"encoding/json"
	"math/big"
)

// UpgradeBytesConfig represents part of the ChainConfig
// that can be upgraded via upgradeBytes
type UpgradeBytesConfig struct {
	// Config for blocks/timestamps that enable network upgrades.
	// Note: if NetworkUpgrades is specified in the JSON all previously activated
	// forks must be present or upgradeBytes will be rejected.
	NetworkUpgrades *NetworkUpgrades `json:"networkUpgrades,omitempty"`

	// Config for enabling and disabling precompiles as network upgrades.
	PrecompileUpgrades []Upgrade `json:"precompileUpgrades,omitempty"`
}

// ApplyUpgradeBytes applies modifications from upgradeBytes to chainConfig
// if upgradeBytes is compatible with activated forks.
func (c *ChainConfig) ApplyUpgradeBytes(upgradeBytes []byte, headTimestamp *big.Int) error {
	var upgradeBytesConfig UpgradeBytesConfig

	// Note: passing an empty slice is considered equivalent to an empty upgradeBytesConfig
	// we will still verify the empty config against the existing chainConfig, to ensure
	// activated upgrades are not removed.
	if len(upgradeBytes) > 0 {
		if err := json.Unmarshal(upgradeBytes, &upgradeBytesConfig); err != nil {
			return err
		}
	}

	// Create an new UpgradesConfig, including the newly parsed upgradeBytes
	newUpgradesConfig := &UpgradesConfig{
		Upgrade:         c.Upgrade,
		NetworkUpgrades: c.NetworkUpgrades,

		UpgradeBytesConfig: upgradeBytesConfig,
	}
	// verify the newly constructed UpgradesConfig is consistent.
	if err := newUpgradesConfig.ValidatePrecompileUpgrades(); err != nil {
		return err
	}

	// Check compatibility of network upgrades
	if c.UpgradeBytesConfig.NetworkUpgrades != nil && upgradeBytesConfig.NetworkUpgrades == nil {
		// if we have previously applied persisted upgrade bytes,
		// missing "networkUpgrades" will be treated as intention to
		// abort forks. Initialize NetworkUpgrades here.
		upgradeBytesConfig.NetworkUpgrades = &NetworkUpgrades{}
	}
	if networkUpgrades := upgradeBytesConfig.NetworkUpgrades; networkUpgrades != nil {
		if err := c.getNetworkUpgrades().CheckCompatible(networkUpgrades, headTimestamp); err != nil {
			return err
		}
	}

	// verify the newly constructed UpgradesConfig is compatible with the existing chainConfig.
	if err := c.UpgradesConfig.CheckCompatible(newUpgradesConfig, headTimestamp); err != nil {
		return err
	}

	// Apply upgrades to chainConfig.
	c.UpgradeBytesConfig = upgradeBytesConfig
	return nil
}