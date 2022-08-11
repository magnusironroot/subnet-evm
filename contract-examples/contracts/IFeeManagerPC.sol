//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
import "./IAllowList.sol";

interface IFeeManagerPC is IAllowList {
  // Set fee config fields to contract storage
  function setFeeConfigPC(
    uint256 gasLimit,
    uint256 targetBlockRate,
    uint256 minBaseFee,
    uint256 targetGas,
    uint256 baseFeeChangeDenominator,
    uint256 minBlockGasCost,
    uint256 maxBlockGasCost,
    uint256 blockGasCostStep
  ) external;

  // Get fee config from the contract storage
  function getFeeConfigPC()
    external
    view
    returns (
      uint256 gasLimit,
      uint256 targetBlockRate,
      uint256 minBaseFee,
      uint256 targetGas,
      uint256 baseFeeChangeDenominator,
      uint256 minBlockGasCost,
      uint256 maxBlockGasCost,
      uint256 blockGasCostStep
    );

  // Get the last block number changed the fee config from the contract storage
  function getFeeConfigPCLastChangedAt() external view returns (uint256 blockNumber);

  function getTestPC(uint256 testt) external view returns (uint256 blockNumber);

  fallback() external;
}