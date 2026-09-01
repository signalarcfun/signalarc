// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../../src/agent/SignalArcAgentMarket.sol";
import "../../src/agent/SignalArcAgentMarketFactory.sol";
import "../MockUSDC.sol";

interface AgentFactoryVm {
    function expectEmit(bool checkTopic1, bool checkTopic2, bool checkTopic3, bool checkData) external;
    function expectRevert(bytes calldata revertData) external;
    function expectRevert(bytes4 revertData) external;
    function prank(address msgSender) external;
}

contract SignalArcAgentMarketFactoryTest {
    AgentFactoryVm private constant vm = AgentFactoryVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant ADMIN = address(0xAD111);
    address private constant RESOLVER = address(0xA11CE);
    string private constant MARKET_ID = "agent-market-1";
    string private constant QUESTION = "Will SignalArc deploy a separate agent market factory?";

    event AgentMarketDeployed(
        string indexed marketId,
        address indexed market,
        address indexed admin,
        address resolver,
        address collateralToken,
        uint256 closeTimestamp,
        string question
    );

    function testFactoryDeploysNewAgentMarket() external {
        (SignalArcAgentMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.prank(ADMIN);
        address marketAddress = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertTrue(marketAddress != address(0));
        assertTrue(factory.isMarket(marketAddress));
        assertEq(factory.marketById(MARKET_ID), marketAddress);
        assertEq(factory.marketCount(), 1);
        assertEq(factory.allMarkets(0), marketAddress);
    }

    function testFactoryDeploysAgentMarketWithInitialState() external {
        (SignalArcAgentMarketFactory factory, MockUSDC token) = createFactoryAndToken();
        uint256 marketCloseTimestamp = closeTimestamp();

        vm.prank(ADMIN);
        SignalArcAgentMarket market =
            SignalArcAgentMarket(factory.createMarket(MARKET_ID, QUESTION, marketCloseTimestamp, RESOLVER, address(token)));

        assertEq(market.question(), QUESTION);
        assertEq(market.closeTimestamp(), marketCloseTimestamp);
        assertEq(market.admin(), ADMIN);
        assertEq(market.resolver(), RESOLVER);
        assertEq(address(market.collateralToken()), address(token));
        assertEq(uint256(market.status()), uint256(SignalArcAgentMarket.MarketStatus.Open));
    }

    function testFactoryEmitsAgentMarketDeployedEvent() external {
        (SignalArcAgentMarketFactory factory, MockUSDC token) = createFactoryAndToken();
        uint256 marketCloseTimestamp = closeTimestamp();
        address expectedMarket = predictNextContractAddress(address(factory), 1);

        vm.expectEmit(true, true, true, true);
        emit AgentMarketDeployed(
            MARKET_ID, expectedMarket, ADMIN, RESOLVER, address(token), marketCloseTimestamp, QUESTION
        );

        vm.prank(ADMIN);
        factory.createMarket(MARKET_ID, QUESTION, marketCloseTimestamp, RESOLVER, address(token));
    }

    function testDuplicateAgentMarketIdReverts() external {
        (SignalArcAgentMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.prank(ADMIN);
        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        vm.expectRevert(abi.encodeWithSelector(SignalArcAgentMarketFactory.MarketAlreadyExists.selector, MARKET_ID));

        vm.prank(ADMIN);
        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));
    }

    function testFactoryRejectsInvalidInputs() external {
        (SignalArcAgentMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.expectRevert(SignalArcAgentMarketFactory.EmptyMarketId.selector);
        factory.createMarket("", QUESTION, closeTimestamp(), RESOLVER, address(token));

        vm.expectRevert(SignalArcAgentMarketFactory.EmptyQuestion.selector);
        factory.createMarket(MARKET_ID, "", closeTimestamp(), RESOLVER, address(token));

        vm.expectRevert(SignalArcAgentMarketFactory.InvalidResolver.selector);
        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), address(0), address(token));

        vm.expectRevert(SignalArcAgentMarketFactory.InvalidCollateralToken.selector);
        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(0));
    }

    function createFactoryAndToken() private returns (SignalArcAgentMarketFactory, MockUSDC) {
        return (new SignalArcAgentMarketFactory(), new MockUSDC());
    }

    function closeTimestamp() private view returns (uint256) {
        return block.timestamp + 1 days;
    }

    function predictNextContractAddress(address deployer, uint256 nonce) private pure returns (address) {
        if (nonce == 0) {
            return addressFromRlp(abi.encodePacked(bytes1(0xd6), bytes1(0x94), deployer, bytes1(0x80)));
        }

        if (nonce <= 0x7f) {
            return addressFromRlp(abi.encodePacked(bytes1(0xd6), bytes1(0x94), deployer, uint8(nonce)));
        }

        revert("nonce too large");
    }

    function addressFromRlp(bytes memory rlpEncoded) private pure returns (address) {
        return address(uint160(uint256(keccak256(rlpEncoded))));
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
