// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/SignalArcMarket.sol";
import "./MockUSDC.sol";

interface Vm {
    function expectRevert(bytes4 revertData) external;
    function prank(address msgSender) external;
    function warp(uint256 newTimestamp) external;
}

contract SignalArcMarketTest {
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant RESOLVER = address(0xA11CE);
    address private constant USER = address(0xB0B);
    address private constant SPENDER = address(0xCAFE);
    address private constant TRADER = address(0xD00D);
    string private constant QUESTION = "Will SignalArc complete Phase 6?";

    function testConstructorSetsQuestion() external {
        SignalArcMarket market = createMarket();

        assertEq(market.question(), QUESTION);
    }

    function testConstructorSetsCloseTimestamp() external {
        uint256 closeTimestamp = block.timestamp + 1 days;
        MockUSDC token = new MockUSDC();
        SignalArcMarket market = new SignalArcMarket(QUESTION, closeTimestamp, RESOLVER, address(token));

        assertEq(market.closeTimestamp(), closeTimestamp);
    }

    function testConstructorSetsResolver() external {
        SignalArcMarket market = createMarket();

        assertEq(market.resolver(), RESOLVER);
    }

    function testConstructorSetsStatusOpen() external {
        SignalArcMarket market = createMarket();

        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Open));
    }

    function testConstructorSetsWinningOutcomeNone() external {
        SignalArcMarket market = createMarket();

        assertEq(uint256(market.winningOutcome()), uint256(SignalArcMarket.Outcome.None));
    }

    function testConstructorStoresCollateralTokenAddress() external {
        MockUSDC token = new MockUSDC();
        SignalArcMarket market = new SignalArcMarket(QUESTION, block.timestamp + 1 days, RESOLVER, address(token));

        assertEq(address(market.collateralToken()), address(token));
    }

    function testConstructorRevertsOnEmptyQuestion() external {
        MockUSDC token = new MockUSDC();

        vm.expectRevert(SignalArcMarket.EmptyQuestion.selector);

        new SignalArcMarket("", block.timestamp + 1 days, RESOLVER, address(token));
    }

    function testConstructorRevertsOnCloseTimestampNotInFuture() external {
        MockUSDC token = new MockUSDC();

        vm.expectRevert(SignalArcMarket.InvalidCloseTimestamp.selector);

        new SignalArcMarket(QUESTION, block.timestamp, RESOLVER, address(token));
    }

    function testConstructorRevertsOnZeroResolver() external {
        MockUSDC token = new MockUSDC();

        vm.expectRevert(SignalArcMarket.InvalidResolver.selector);

        new SignalArcMarket(QUESTION, block.timestamp + 1 days, address(0), address(token));
    }

    function testConstructorRevertsOnZeroCollateralToken() external {
        vm.expectRevert(SignalArcMarket.InvalidCollateralToken.selector);

        new SignalArcMarket(QUESTION, block.timestamp + 1 days, RESOLVER, address(0));
    }

    function testIsOpenReturnsTrueBeforeClose() external {
        SignalArcMarket market = createMarket();

        assertTrue(market.isOpen());
    }

    function testCloseMarketRevertsBeforeCloseTime() external {
        SignalArcMarket market = createMarket();

        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.closeMarket();
    }

    function testCloseMarketSucceedsAfterCloseTime() external {
        uint256 closeTimestamp = block.timestamp + 1 days;
        MockUSDC token = new MockUSDC();
        SignalArcMarket market = new SignalArcMarket(QUESTION, closeTimestamp, RESOLVER, address(token));

        vm.warp(closeTimestamp);
        market.closeMarket();

        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Closed));
    }

    function testOnlyResolverCanCancel() external {
        SignalArcMarket market = createMarket();

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.UnauthorizedResolver.selector);
        market.cancelMarket();
    }

    function testResolverCanCancelBeforeFinalized() external {
        SignalArcMarket market = createMarket();

        vm.prank(RESOLVER);
        market.cancelMarket();

        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Cancelled));
    }

    function testCannotCancelAfterResolved() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(SignalArcMarket.Outcome.Yes);

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.MarketAlreadyFinalized.selector);
        market.cancelMarket();
    }

    function testOnlyResolverCanResolve() external {
        SignalArcMarket market = createClosedMarket();

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.UnauthorizedResolver.selector);
        market.resolve(SignalArcMarket.Outcome.Yes);
    }

    function testResolveRequiresClosedStatus() external {
        SignalArcMarket market = createMarket();

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.resolve(SignalArcMarket.Outcome.Yes);
    }

    function testCannotResolveAfterCancelled() external {
        SignalArcMarket market = createMarket();

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.resolve(SignalArcMarket.Outcome.Yes);
    }

    function testResolveRejectsOutcomeNone() external {
        SignalArcMarket market = createClosedMarket();

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.InvalidOutcome.selector);
        market.resolve(SignalArcMarket.Outcome.None);
    }

    function testResolveAcceptsOutcomeYes() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(SignalArcMarket.Outcome.Yes);

        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Resolved));
        assertEq(uint256(market.winningOutcome()), uint256(SignalArcMarket.Outcome.Yes));
    }

    function testResolveAcceptsOutcomeNo() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(SignalArcMarket.Outcome.No);

        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Resolved));
        assertEq(uint256(market.winningOutcome()), uint256(SignalArcMarket.Outcome.No));
    }

    function testResolverCannotResolveTwice() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(SignalArcMarket.Outcome.Yes);

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.resolve(SignalArcMarket.Outcome.No);
    }

    function testOpenPositionRevertsBeforeApprovalInsufficientAllowance() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();

        vm.prank(USER);
        vm.expectRevert(MockUSDC.InsufficientAllowance.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 100);

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testOpenPositionRevertsWithAmountZero() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.InvalidAmount.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 0);
    }

    function testOpenPositionRejectsOutcomeNone() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.InvalidSide.selector);
        market.openPosition(SignalArcMarket.Outcome.None, 100);
    }

    function testOpenPositionRejectsInvalidSideIfPossible() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(USER);
        (bool success,) = address(market).call(abi.encodeWithSignature("openPosition(uint8,uint256)", 3, 100));

        assertFalse(success);
    }

    function testOpenPositionAcceptsYesAfterApproval() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.Yes, 100);

        assertEq(market.yesPositions(USER), 100);
    }

    function testOpenPositionAcceptsNoAfterApproval() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.No, 100);

        assertEq(market.noPositions(USER), 100);
    }

    function testOpenPositionUpdatesYesPositions() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 150);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.Yes, 150);

        assertEq(market.yesPositions(USER), 150);
    }

    function testOpenPositionUpdatesNoPositions() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 200);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.No, 200);

        assertEq(market.noPositions(USER), 200);
    }

    function testOpenPositionUpdatesTotalYes() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 250);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.Yes, 250);

        assertEq(market.totalYes(), 250);
    }

    function testOpenPositionUpdatesTotalNo() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 300);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.No, 300);

        assertEq(market.totalNo(), 300);
    }

    function testOpenPositionUpdatesTotalCollateral() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 350);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.Yes, 350);

        assertEq(market.totalCollateral(), 350);
    }

    function testOpenPositionTransfersMockUSDCFromUserToMarket() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 400);

        vm.prank(USER);
        market.openPosition(SignalArcMarket.Outcome.Yes, 400);

        assertEq(token.balanceOf(USER), 600);
        assertEq(token.balanceOf(address(market)), 400);
    }

    function testUserCannotOpenMorePositionThanMockUSDCBalance() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 1_001);

        vm.prank(USER);
        vm.expectRevert(MockUSDC.InsufficientBalance.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 1_001);
    }

    function testUserCannotOpenMorePositionThanApprovedAllowance() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 99);

        vm.prank(USER);
        vm.expectRevert(MockUSDC.InsufficientAllowance.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 100);
    }

    function testMultipleUsersCanOpenPositionsIndependently() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.No, 250);

        assertEq(market.yesPositions(USER), 100);
        assertEq(market.noPositions(SPENDER), 250);
        assertEq(market.totalYes(), 100);
        assertEq(market.totalNo(), 250);
        assertEq(market.totalCollateral(), 350);
    }

    function testOpenPositionRevertsAfterMarketCloseTime() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.warp(market.closeTimestamp());
        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 100);
    }

    function testOpenPositionRevertsAfterCancelledMarket() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.openPosition(SignalArcMarket.Outcome.Yes, 100);
    }

    function testOpenPositionRevertsAfterResolvedMarket() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        approveMarket(token, market, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.openPosition(SignalArcMarket.Outcome.No, 100);
    }

    function testCannotCloseCancelledMarket() external {
        SignalArcMarket market = createMarket();

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.closeMarket();
    }

    function testCannotCloseResolvedMarket() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(SignalArcMarket.Outcome.Yes);

        vm.expectRevert(SignalArcMarket.MarketNotOpen.selector);
        market.closeMarket();
    }

    function testClaimableAmountReturnsZeroBeforeResolution() external {
        (SignalArcMarket market,) = createFundedMarket();

        assertEq(market.claimableAmount(USER), 0);
    }

    function testClaimableAmountReturnsProRataTotalCollateralAfterYesResolution() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 300);
        openPositionAs(token, market, TRADER, SignalArcMarket.Outcome.No, 600);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        assertEq(market.claimableAmount(USER), 250);
        assertEq(market.claimableAmount(SPENDER), 750);
    }

    function testClaimableAmountReturnsProRataTotalCollateralAfterNoResolution() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 150);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.No, 50);
        openPositionAs(token, market, TRADER, SignalArcMarket.Outcome.Yes, 200);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.No);

        assertEq(market.claimableAmount(USER), 300);
        assertEq(market.claimableAmount(SPENDER), 100);
    }

    function testClaimableAmountReturnsZeroForLosingSide() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 200);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        assertEq(market.claimableAmount(USER), 0);
    }

    function testClaimableAmountReturnsCombinedYesAndNoAmountAfterCancellation() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 125);
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 175);

        vm.prank(RESOLVER);
        market.cancelMarket();

        assertEq(market.claimableAmount(USER), 300);
    }

    function testResolveRevertsIfYesWinsWithNoYesStake() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.NoWinningStake.selector);
        market.resolve(SignalArcMarket.Outcome.Yes);
    }

    function testResolveRevertsIfNoWinsWithNoNoStake() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeMarket(market);

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcMarket.NoWinningStake.selector);
        market.resolve(SignalArcMarket.Outcome.No);
    }

    function testClaimRevertsBeforeResolutionOrCancellation() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.MarketNotResolved.selector);
        market.claim();
    }

    function testWinningYesUserCanClaimAfterYesResolution() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testMultipleUsersCanClaimIndependentlyAfterWinningOutcome() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 200);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        market.claim();

        vm.prank(SPENDER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
        assertEq(token.balanceOf(SPENDER), 1_000);
        assertTrue(market.hasClaimed(USER));
        assertTrue(market.hasClaimed(SPENDER));
    }

    function testLosingUserRemainsUnableToClaimWhileWinningUserCanClaim() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.No, 200);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        market.claim();

        vm.prank(SPENDER);
        vm.expectRevert(SignalArcMarket.NothingToClaim.selector);
        market.claim();

        assertEq(token.balanceOf(USER), 1_200);
        assertEq(token.balanceOf(SPENDER), 800);
    }

    function testWinningNoUserCanClaimAfterNoResolution() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.No);

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testLosingUserCannotClaim() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 100);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.NothingToClaim.selector);
        market.claim();
    }

    function testUserCannotClaimTwice() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        market.claim();

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.AlreadyClaimed.selector);
        market.claim();
    }

    function testClaimTransfersMockUSDCFromMarketToUser() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 250);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.No, 750);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        assertEq(token.balanceOf(address(market)), 1_000);

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_750);
        assertEq(token.balanceOf(address(market)), 0);
    }

    function testClaimMarksHasClaimedTrue() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(USER);
        market.claim();

        assertTrue(market.hasClaimed(USER));
    }

    function testCancelledMarketRefundsYesPosition() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testCancelledMarketRefundsNoPosition() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 100);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testCancelledMarketRefundsCombinedYesAndNoPosition() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.No, 200);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(USER);
        market.claim();

        assertEq(token.balanceOf(USER), 1_000);
    }

    function testResolvedMarketLeavesRoundingDustInContract() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 1);
        openPositionAs(token, market, SPENDER, SignalArcMarket.Outcome.Yes, 2);
        openPositionAs(token, market, TRADER, SignalArcMarket.Outcome.No, 1);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        assertEq(market.claimableAmount(USER), 1);
        assertEq(market.claimableAmount(SPENDER), 2);

        vm.prank(USER);
        market.claim();

        vm.prank(SPENDER);
        market.claim();

        assertEq(token.balanceOf(address(market)), 1);
    }

    function testRefundPathDoesNotAllowDoubleClaimAfterCancellation() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(USER);
        market.claim();

        vm.prank(USER);
        vm.expectRevert(SignalArcMarket.AlreadyClaimed.selector);
        market.claim();
    }

    function testUserWithoutPositionCannotClaimAfterResolution() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);
        closeAndResolveMarket(market, SignalArcMarket.Outcome.Yes);

        vm.prank(SPENDER);
        vm.expectRevert(SignalArcMarket.NothingToClaim.selector);
        market.claim();
    }

    function testUserWithoutPositionCannotClaimAfterCancellation() external {
        (SignalArcMarket market, MockUSDC token) = createFundedMarket();
        openPositionAs(token, market, USER, SignalArcMarket.Outcome.Yes, 100);

        vm.prank(RESOLVER);
        market.cancelMarket();

        vm.prank(SPENDER);
        vm.expectRevert(SignalArcMarket.NothingToClaim.selector);
        market.claim();
    }

    function testMockUSDCDecimalsIsSix() external {
        MockUSDC token = new MockUSDC();

        assertEq(uint256(token.decimals()), 6);
    }

    function testMockUSDCMintIncreasesBalance() external {
        MockUSDC token = new MockUSDC();

        token.mint(USER, 100);

        assertEq(token.balanceOf(USER), 100);
    }

    function testMockUSDCApproveSetsAllowance() external {
        MockUSDC token = new MockUSDC();

        vm.prank(USER);
        bool approved = token.approve(SPENDER, 100);

        assertTrue(approved);
        assertEq(token.allowance(USER, SPENDER), 100);
    }

    function testMockUSDCTransferMovesBalance() external {
        MockUSDC token = new MockUSDC();
        token.mint(USER, 100);

        vm.prank(USER);
        bool transferred = token.transfer(SPENDER, 40);

        assertTrue(transferred);
        assertEq(token.balanceOf(USER), 60);
        assertEq(token.balanceOf(SPENDER), 40);
    }

    function testMockUSDCTransferFromRespectsAllowance() external {
        MockUSDC token = new MockUSDC();
        token.mint(USER, 100);

        vm.prank(USER);
        token.approve(SPENDER, 80);

        vm.prank(SPENDER);
        bool transferred = token.transferFrom(USER, RESOLVER, 50);

        assertTrue(transferred);
        assertEq(token.balanceOf(USER), 50);
        assertEq(token.balanceOf(RESOLVER), 50);
        assertEq(token.allowance(USER, SPENDER), 30);
    }

    function testMockUSDCTransferFromRevertsOnInsufficientAllowance() external {
        MockUSDC token = new MockUSDC();
        token.mint(USER, 100);

        vm.prank(USER);
        token.approve(SPENDER, 40);

        vm.prank(SPENDER);
        vm.expectRevert(MockUSDC.InsufficientAllowance.selector);
        token.transferFrom(USER, RESOLVER, 50);
    }

    function testMockUSDCTransferRevertsOnInsufficientBalance() external {
        MockUSDC token = new MockUSDC();

        vm.prank(USER);
        vm.expectRevert(MockUSDC.InsufficientBalance.selector);
        token.transfer(SPENDER, 1);
    }

    function createMarket() private returns (SignalArcMarket) {
        MockUSDC token = new MockUSDC();

        return new SignalArcMarket(QUESTION, block.timestamp + 1 days, RESOLVER, address(token));
    }

    function createClosedMarket() private returns (SignalArcMarket) {
        uint256 closeTimestamp = block.timestamp + 1 days;
        MockUSDC token = new MockUSDC();
        SignalArcMarket market = new SignalArcMarket(QUESTION, closeTimestamp, RESOLVER, address(token));

        vm.warp(closeTimestamp);
        market.closeMarket();

        return market;
    }

    function createFundedMarket() private returns (SignalArcMarket, MockUSDC) {
        MockUSDC token = new MockUSDC();
        SignalArcMarket market = new SignalArcMarket(QUESTION, block.timestamp + 1 days, RESOLVER, address(token));
        token.mint(USER, 1_000);
        token.mint(SPENDER, 1_000);
        token.mint(TRADER, 1_000);

        return (market, token);
    }

    function approveMarket(MockUSDC token, SignalArcMarket market, uint256 amount) private {
        vm.prank(USER);
        token.approve(address(market), amount);
    }

    function openPositionAs(
        MockUSDC token,
        SignalArcMarket market,
        address user,
        SignalArcMarket.Outcome side,
        uint256 amount
    ) private {
        vm.prank(user);
        token.approve(address(market), amount);

        vm.prank(user);
        market.openPosition(side, amount);
    }

    function closeAndResolveMarket(SignalArcMarket market, SignalArcMarket.Outcome outcome) private {
        closeMarket(market);

        vm.prank(RESOLVER);
        market.resolve(outcome);
    }

    function closeMarket(SignalArcMarket market) private {
        vm.warp(market.closeTimestamp());
        market.closeMarket();
    }

    function assertTrue(bool actual) private pure {
        if (!actual) {
            revert("assertTrue failed");
        }
    }

    function assertFalse(bool actual) private pure {
        if (actual) {
            revert("assertFalse failed");
        }
    }

    function assertEq(string memory actual, string memory expected) private pure {
        if (keccak256(bytes(actual)) != keccak256(bytes(expected))) {
            revert("assertEq string failed");
        }
    }

    function assertEq(address actual, address expected) private pure {
        if (actual != expected) {
            revert("assertEq address failed");
        }
    }

    function assertEq(uint256 actual, uint256 expected) private pure {
        if (actual != expected) {
            revert("assertEq uint256 failed");
        }
    }
}
