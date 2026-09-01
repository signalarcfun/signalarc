// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../../src/agent/SignalArcAgentMarket.sol";
import "../MockUSDC.sol";

interface AgentMarketVm {
    function expectRevert(bytes4 revertData) external;
    function prank(address msgSender) external;
    function warp(uint256 newTimestamp) external;
}

contract SignalArcAgentMarketTest {
    AgentMarketVm private constant vm = AgentMarketVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant ADMIN = address(0xAD111);
    address private constant RESOLVER = address(0xA11CE);
    address private constant USER = address(0xB0B);
    address private constant OTHER_USER = address(0xCAFE);
    string private constant QUESTION = "Will SignalArc agent contracts stay isolated?";

    function testAgentMarketInitialState() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();

        assertEq(market.question(), QUESTION);
        assertEq(market.admin(), ADMIN);
        assertEq(market.resolver(), RESOLVER);
        assertEq(address(market.collateralToken()), address(token));
        assertEq(uint256(market.status()), uint256(SignalArcAgentMarket.MarketStatus.Open));
        assertEq(uint256(market.winningOutcome()), uint256(SignalArcAgentMarket.Outcome.None));
        assertTrue(market.isOpen());
    }

    function testBuyYes() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();

        buyYesAs(token, market, USER, 100);

        assertEq(market.yesPositions(USER), 100);
        assertEq(market.totalYes(), 100);
        assertEq(market.totalCollateral(), 100);
        assertEq(token.balanceOf(address(market)), 100);
    }

    function testBuyNo() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();

        buyNoAs(token, market, USER, 125);

        assertEq(market.noPositions(USER), 125);
        assertEq(market.totalNo(), 125);
        assertEq(market.totalCollateral(), 125);
        assertEq(token.balanceOf(address(market)), 125);
    }

    function testCancelAndClaimRefund() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();
        buyYesAs(token, market, USER, 100);
        buyNoAs(token, market, USER, 50);

        vm.prank(ADMIN);
        market.cancelMarket();

        assertEq(uint256(market.status()), uint256(SignalArcAgentMarket.MarketStatus.Cancelled));
        assertEq(market.claimableRefund(USER), 150);

        vm.prank(USER);
        market.claimRefund();

        assertEq(token.balanceOf(USER), 1_000);
        assertTrue(market.hasClaimed(USER));
    }

    function testCloseResolveYesAndClaimPayout() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();
        buyYesAs(token, market, USER, 200);
        buyNoAs(token, market, OTHER_USER, 75);

        closeAndResolve(market, SignalArcAgentMarket.Outcome.Yes);

        assertEq(uint256(market.status()), uint256(SignalArcAgentMarket.MarketStatus.Resolved));
        assertEq(uint256(market.winningOutcome()), uint256(SignalArcAgentMarket.Outcome.Yes));
        assertEq(market.claimablePayout(USER), 200);
        assertEq(market.claimablePayout(OTHER_USER), 0);

        vm.prank(USER);
        market.claimPayout();

        assertEq(token.balanceOf(USER), 1_000);
        assertTrue(market.hasClaimed(USER));
    }

    function testUnauthorizedAdminOrResolverActionFails() external {
        (SignalArcAgentMarket market,) = createFundedMarket();

        vm.prank(USER);
        vm.expectRevert(SignalArcAgentMarket.UnauthorizedAdminOrResolver.selector);
        market.cancelMarket();

        vm.warp(market.closeTimestamp());
        vm.prank(USER);
        vm.expectRevert(SignalArcAgentMarket.UnauthorizedAdminOrResolver.selector);
        market.closeMarket();
    }

    function testUnauthorizedResolverActionFails() external {
        (SignalArcAgentMarket market,) = createClosedMarket();

        vm.prank(USER);
        vm.expectRevert(SignalArcAgentMarket.UnauthorizedResolver.selector);
        market.resolve(SignalArcAgentMarket.Outcome.Yes);
    }

    function testInvalidLifecycleTransitionFails() external {
        (SignalArcAgentMarket market,) = createFundedMarket();

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcAgentMarket.MarketNotClosed.selector);
        market.resolve(SignalArcAgentMarket.Outcome.Yes);

        vm.prank(ADMIN);
        market.cancelMarket();

        vm.prank(RESOLVER);
        vm.expectRevert(SignalArcAgentMarket.MarketNotClosed.selector);
        market.resolve(SignalArcAgentMarket.Outcome.Yes);
    }

    function testCannotBuyAfterCancelled() external {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();

        vm.prank(ADMIN);
        market.cancelMarket();

        vm.prank(USER);
        token.approve(address(market), 100);

        vm.prank(USER);
        vm.expectRevert(SignalArcAgentMarket.MarketNotOpen.selector);
        market.buyYes(100);
    }

    function createFundedMarket() private returns (SignalArcAgentMarket, MockUSDC) {
        MockUSDC token = new MockUSDC();
        SignalArcAgentMarket market =
            new SignalArcAgentMarket(QUESTION, block.timestamp + 1 days, ADMIN, RESOLVER, address(token));
        token.mint(USER, 1_000);
        token.mint(OTHER_USER, 1_000);

        return (market, token);
    }

    function createClosedMarket() private returns (SignalArcAgentMarket, MockUSDC) {
        (SignalArcAgentMarket market, MockUSDC token) = createFundedMarket();

        vm.warp(market.closeTimestamp());
        vm.prank(ADMIN);
        market.closeMarket();

        return (market, token);
    }

    function buyYesAs(MockUSDC token, SignalArcAgentMarket market, address user, uint256 amount) private {
        vm.prank(user);
        token.approve(address(market), amount);

        vm.prank(user);
        market.buyYes(amount);
    }

    function buyNoAs(MockUSDC token, SignalArcAgentMarket market, address user, uint256 amount) private {
        vm.prank(user);
        token.approve(address(market), amount);

        vm.prank(user);
        market.buyNo(amount);
    }

    function closeAndResolve(SignalArcAgentMarket market, SignalArcAgentMarket.Outcome outcome) private {
        vm.warp(market.closeTimestamp());

        vm.prank(ADMIN);
        market.closeMarket();

        vm.prank(RESOLVER);
        market.resolve(outcome);
    }

    function assertTrue(bool actual) private pure {
        if (!actual) {
            revert("assertTrue failed");
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
