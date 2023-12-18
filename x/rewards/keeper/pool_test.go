package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/lavanet/lava/testutil/keeper"
	"github.com/lavanet/lava/utils"
	"github.com/lavanet/lava/x/rewards/types"
	"github.com/stretchr/testify/require"
)

// GENERAL NOTES:
// 1. To avoid simulating Cosmos' distribution module, all tests check the "Fee Collector"
//    account's balance since the distribution module takes funds from there to reward
//    the validators every new block
//
// 2. The rewards pool mechanism relies on a monthly EndBlock timer callback that refills the pool
//	  and opens a new timer. In some tests you'll see: AdvanceMonth -> AdvanceBlock -> EndBlock
//	  This is because AdvanceMonth advances a month minus 5 seconds, AdvanceBlock advances the
//	  time by 300sec (block time) and EndBlock calls the EndBlock functions of all keepers (so
//    the timer callback will be called). We need to call EndBlock because AdvanceBlock calls
//	  EndBlock for the previous block, updates the context with the new block height and time
//	  and then calls BeginBlock. The timer callback will be called only through EndBlock that
//	  uses the new height and time, so EndBlock needs to be called explicitly.

// TestRewardsModuleSetup tests that the setup of the rewards module is as expected
// The setup does the following (as in the Rewards module genesis):
//  1. transfer funds to the allocation pools
//  2. inits the refill rewards pools timer store
//  3. calls RefillRewardsPools to transfer funds to the distribution pools
//
// The expected results (after an epoch has passed) is that:
//  1. The allocation pool has the expected allocated funds minus one block reward
//  2. The distribution pool has the expected monthly quota minus one block reward
//  3. The fee collector has one block reward
//
// the validator got rewards
func TestRewardsModuleSetup(t *testing.T) {
	ts := newTester(t)
	lifetime := types.RewardsAllocationPoolsLifetime

	// on init, the allocation pool lifetime should decrease by one
	res, err := ts.QueryRewardsPools()
	require.Nil(t, err)
	require.Equal(t, lifetime-1, res.AllocationPoolMonthsLeft)

	// in the end of the setup, there is an advancement of one block, so validator
	// rewards were distributed once. Since the block rewards depends on the distribution
	// pool balance (and it's not negligble), we'll calculate it manually
	expectedDistPoolBalanceBeforeReward := allocationPoolBalance / lifetime
	expectedBlocksToNextExpiry := ts.Keepers.Rewards.BlocksToNextTimerExpiry(ts.Ctx)
	require.NotEqual(t, int64(0), expectedBlocksToNextExpiry)
	expectedTargetFactor := int64(1)
	blockReward := expectedTargetFactor * expectedDistPoolBalanceBeforeReward / expectedBlocksToNextExpiry

	// after setup, the allocation pool got funded and sent the monthly quota to the distribution pool
	for _, pool := range res.Pools {
		switch pool.Name {
		case string(types.ValidatorsRewardsAllocationPoolName):
			require.Equal(t, allocationPoolBalance*(lifetime-1)/lifetime, pool.Balance.Amount.Int64())
		case string(types.ValidatorsRewardsDistributionPoolName):
			require.Equal(t, (allocationPoolBalance/lifetime)-blockReward, pool.Balance.Amount.Int64())
		}
	}

	// check the fee collector's balance is the block reward (see general note 1 above)
	balance := ts.GetBalance(ts.feeCollector())
	require.Equal(t, blockReward, balance)
}

// TestBurnRateParam tests that the BurnRate param influences tokens burning as expected
// BurnRate = 1 -> on monthly refill, burn all previous funds in the distribution pool
// BurnRate = 0 -> on monthly refill, burn none of the previous funds in the distribution pool
func TestBurnRateParam(t *testing.T) {
	ts := newTester(t)
	lifetime := types.RewardsAllocationPoolsLifetime
	allocPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsAllocationPoolName).Int64()

	// advance a month to trigger monthly pool refill callback
	// to see why these 3 are called, see general note 2
	ts.AdvanceMonths(1)
	ts.AdvanceBlock()
	testkeeper.EndBlock(ts.Ctx, ts.Keepers)

	// default burn rate = 1, distribution pool's old balance should be wiped
	// current balance should be exactly the expected monthly quota
	expectedMonthlyQuota := allocPoolBalance / (lifetime - 1)
	distPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsDistributionPoolName).Int64()
	require.Equal(t, expectedMonthlyQuota, distPoolBalance)

	// change the burn rate param to be zero
	paramKey := string(types.KeyLeftoverBurnRate)
	zeroBurnRate, err := sdk.ZeroDec().MarshalJSON()
	require.Nil(t, err)
	paramVal := string(zeroBurnRate)
	err = ts.TxProposalChangeParam(types.ModuleName, paramKey, paramVal)
	require.Nil(t, err)

	// advance a month to trigger monthly pool refill callback
	ts.AdvanceMonths(1)
	ts.AdvanceBlock()
	prevDistPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsDistributionPoolName).Int64()
	testkeeper.EndBlock(ts.Ctx, ts.Keepers)

	// burn rate = 0, distribution pool's old balance should not be wiped
	// current balance should be previous balance (minus block reward) plus new quota
	distPoolBalance = ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsDistributionPoolName).Int64()
	require.Equal(t, prevDistPoolBalance+expectedMonthlyQuota, distPoolBalance)
}

// TestAllocationPoolMonthlyQuota tests that the allocation pool transfers to the distribution pool
// its balance divided by months left (which should decrease with time). Also checks that if there are
// no months left, quota = 0 (and the chain doesn't panic)
func TestAllocationPoolMonthlyQuota(t *testing.T) {
	// after init, the allocation pool transfers funds to the distribution pool (no need to wait a month)
	ts := newTester(t)
	lifetime := types.RewardsAllocationPoolsLifetime

	// calc expectedMonthlyQuota. Check that it was subtracted from the allocation pool and added
	// to the distribution pool (its balance should be the monthly quota minus the fee collector's balance)
	expectedMonthlyQuota := allocationPoolBalance / lifetime
	currentAllocPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsAllocationPoolName)
	require.Equal(t, expectedMonthlyQuota, allocationPoolBalance-currentAllocPoolBalance.Int64())

	feeCollectorBalance := ts.GetBalance(ts.feeCollector())
	currentDistPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsDistributionPoolName)
	require.Equal(t, expectedMonthlyQuota, feeCollectorBalance+currentDistPoolBalance.Int64())

	// check the monthly quota is as expected with advancement of months
	// the last three iterations will be after the allocation pool's funds are depleted
	var feeCollectorFinalBalance int64
	for i := 0; i < int(lifetime+2); i++ {
		// to see why these 3 are called, see general note 2
		ts.AdvanceMonths(1)
		ts.AdvanceBlock()
		testkeeper.EndBlock(ts.Ctx, ts.Keepers)

		// check the allocation pool transfers the expected monthly quota each month
		if i >= 47 {
			expectedMonthlyQuota = 0
			if feeCollectorFinalBalance == 0 {
				feeCollectorFinalBalance = ts.GetBalance(ts.feeCollector())
			} else {
				// fee collector balance should not increase (rewards = 0)
				balance := ts.GetBalance(ts.feeCollector())
				require.Equal(t, feeCollectorFinalBalance, balance)
			}
		} else {
			// adding 1 because setup did the first month
			expectedMonthlyQuota = currentAllocPoolBalance.Int64() / (lifetime - (int64(i) + 1))
		}
		prevAllocPoolBalance := currentAllocPoolBalance
		currentAllocPoolBalance = ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsAllocationPoolName)
		require.Equal(t, expectedMonthlyQuota, prevAllocPoolBalance.Sub(currentAllocPoolBalance).Int64())
	}
}

// TestValidatorBlockRewards tests that the expected block reward is transferred to the fee collector
// the reward should be: (distributionPoolBalance * bondedTargetFactor) / blocksToNextTimerExpiry
func TestValidatorBlockRewards(t *testing.T) {
	ts := newTester(t)
	res, err := ts.QueryRewardsBlockReward()
	require.Nil(t, err)
	refBlockReward := res.Reward.Amount

	// by default, BondedRatio staking module param is smaller than MinBonded rewards module param
	// so bondedTargetFactor = 1. We change MinBonded to zero to change bondedTargetFactor
	paramKey := string(types.KeyMinBondedTarget)
	zeroMinBonded, err := sdk.ZeroDec().MarshalJSON()
	require.Nil(t, err)
	paramVal := string(zeroMinBonded)
	err = ts.TxProposalChangeParam(types.ModuleName, paramKey, paramVal)
	require.Nil(t, err)

	// compare the new block reward to refBlockReward
	params := ts.Keepers.Rewards.GetParams(ts.Ctx)
	maxBonded := params.MaxBondedTarget
	bonded := ts.Keepers.StakingKeeper.BondedRatio(ts.Ctx)
	lowFactor := params.LowFactor
	e1 := maxBonded.Sub(bonded).Quo(maxBonded)
	e2 := bonded.Quo(maxBonded)
	expectedBondedTargetFactor := e1.Add(e2.Mul(lowFactor))
	expectedBlockReward := expectedBondedTargetFactor.MulInt(refBlockReward).TruncateInt()

	res, err = ts.QueryRewardsBlockReward()
	require.Nil(t, err)
	blockReward := res.Reward.Amount
	require.True(t, blockReward.Equal(expectedBlockReward))

	// return the params to default values
	ts.Keepers.Rewards.SetParams(ts.Ctx, types.DefaultParams())
	minBonded := ts.Keepers.Rewards.GetParams(ts.Ctx).MinBondedTarget
	require.True(t, minBonded.Equal(types.DefaultMinBondedTarget))

	// transfer funds from the distribution pool to the allocation pool and check reward
	distPoolBalance := ts.Keepers.Rewards.TotalPoolTokens(ts.Ctx, types.ValidatorsRewardsDistributionPoolName)
	err = ts.Keepers.BankKeeper.SendCoinsFromModuleToModule(
		ts.Ctx,
		string(types.ValidatorsRewardsDistributionPoolName),
		string(types.ValidatorsRewardsAllocationPoolName),
		sdk.NewCoins(sdk.NewCoin(ts.TokenDenom(), distPoolBalance.QuoRaw(2))),
	)
	require.Nil(t, err)

	expectedBlockReward = refBlockReward.QuoRaw(2)
	res, err = ts.QueryRewardsBlockReward()
	require.Nil(t, err)
	blockReward = res.Reward.Amount
	require.True(t, blockReward.Equal(expectedBlockReward))

	// transfer funds back
	err = ts.Keepers.BankKeeper.SendCoinsFromModuleToModule(
		ts.Ctx,
		string(types.ValidatorsRewardsAllocationPoolName),
		string(types.ValidatorsRewardsDistributionPoolName),
		sdk.NewCoins(sdk.NewCoin(ts.TokenDenom(), distPoolBalance.QuoRaw(2))),
	)
	require.Nil(t, err)

	// advance block -> both distPoolBalance and blocksToNextTimerExpiry are changed
	expectedBlocksToExpiry := ts.Keepers.Rewards.BlocksToNextTimerExpiry(ts.Ctx) - 1
	expectedDistPoolBalance := distPoolBalance.Sub(refBlockReward)
	expectedBlockReward = expectedDistPoolBalance.QuoRaw(expectedBlocksToExpiry)
	ts.AdvanceBlock()

	res, err = ts.QueryRewardsBlockReward()
	require.Nil(t, err)
	blockReward = res.Reward.Amount
	require.True(t, blockReward.Equal(expectedBlockReward))
}

// TestBlocksAndTimeToNextExpiry tests that the time/blocks to the next timer expiry are as expected
func TestBlocksAndTimeToNextExpiry(t *testing.T) {
	ts := newTester(t)

	// TimeToNextTimerExpiry should be equal to the number of seconds in a month
	blockTime := ts.BlockTime()
	nextMonth := utils.NextMonth(blockTime)
	secondsInAMonth := nextMonth.UTC().Unix() - blockTime.UTC().Unix()
	timeToExpiry := ts.Keepers.Rewards.TimeToNextTimerExpiry(ts.Ctx)
	require.Equal(t, secondsInAMonth, timeToExpiry)

	// BlocksToNextTimerExpiry should be equal to the number of blocks that pass in a month +5%
	blockCreationTime := int64(ts.Keepers.Downtime.GetParams(ts.Ctx).DowntimeDuration.Seconds())
	blocksInAMonth := (secondsInAMonth / blockCreationTime) * 105 / 100
	blocksToExpiry := ts.Keepers.Rewards.BlocksToNextTimerExpiry(ts.Ctx)
	require.Equal(t, blocksInAMonth, blocksToExpiry)

	// Advance 3 blocks and check again
	ts.AdvanceBlocks(3)
	expectedTimeToExpiry := secondsInAMonth - 3*blockCreationTime
	timeToExpiry = ts.Keepers.Rewards.TimeToNextTimerExpiry(ts.Ctx)
	require.Equal(t, expectedTimeToExpiry, timeToExpiry)

	expectedBlocksToExpiry := blocksInAMonth - 3
	blocksToExpiry = ts.Keepers.Rewards.BlocksToNextTimerExpiry(ts.Ctx)
	require.Equal(t, expectedBlocksToExpiry, blocksToExpiry)
}
