// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/SignalArcMarket.sol";
import "../src/SignalArcMarketFactory.sol";
import "./MockUSDC.sol";

interface FactoryVm {
    function expectEmit(bool checkTopic1, bool checkTopic2, bool checkTopic3, bool checkData) external;
    function expectRevert(bytes calldata revertData) external;
    function expectRevert(bytes4 revertData) external;
    function prank(address msgSender) external;
}

contract SignalArcMarketFactoryTest {
    FactoryVm private constant vm = FactoryVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant CREATOR = address(0xC0FFEE);
    address private constant RESOLVER = address(0xA11CE);
    string private constant MARKET_ID = "10000000-0000-4000-8000-000000000003";
    string private constant SECOND_MARKET_ID = "10000000-0000-4000-8000-000000000004";
    string private constant QUESTION = "Will SignalArc deploy per-market contracts?";
    string private constant SECOND_QUESTION = "Will SignalArc support a second market?";

    event MarketDeployed(
        string indexed marketId,
        address indexed market,
        address indexed creator,
        address resolver,
        address collateralToken,
        uint256 closeTimestamp,
        string question
    );

    function testCreateMarketDeploysSignalArcMarket() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        address marketAddress = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertTrue(marketAddress != address(0));
        assertTrue(factory.isMarket(marketAddress));
    }

    function testMarketByIdReturnsDeployedAddress() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        address marketAddress = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertEq(factory.marketById(MARKET_ID), marketAddress);
    }

    function testIsMarketReturnsTrueForDeployedMarket() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        address marketAddress = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertTrue(factory.isMarket(marketAddress));
    }

    function testMarketCountIncrements() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        assertEq(factory.marketCount(), 0);

        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertEq(factory.marketCount(), 1);
    }

    function testAllMarketsStoresDeployedAddress() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        address marketAddress = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        assertEq(factory.allMarkets(0), marketAddress);
    }

    function testDeployedMarketConstructorValuesAreCorrect() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();
        uint256 marketCloseTimestamp = closeTimestamp();

        SignalArcMarket market =
            SignalArcMarket(factory.createMarket(MARKET_ID, QUESTION, marketCloseTimestamp, RESOLVER, address(token)));

        assertEq(market.question(), QUESTION);
        assertEq(market.closeTimestamp(), marketCloseTimestamp);
        assertEq(market.resolver(), RESOLVER);
        assertEq(address(market.collateralToken()), address(token));
        assertEq(uint256(market.status()), uint256(SignalArcMarket.MarketStatus.Open));
    }

    function testMarketDeployedEventIsEmitted() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();
        uint256 marketCloseTimestamp = closeTimestamp();
        address expectedMarket = predictNextContractAddress(address(factory), 1);

        vm.expectEmit(true, true, true, true);
        emit MarketDeployed(
            MARKET_ID, expectedMarket, CREATOR, RESOLVER, address(token), marketCloseTimestamp, QUESTION
        );

        vm.prank(CREATOR);
        factory.createMarket(MARKET_ID, QUESTION, marketCloseTimestamp, RESOLVER, address(token));
    }

    function testDuplicateMarketIdReverts() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();
        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));

        vm.expectRevert(abi.encodeWithSelector(SignalArcMarketFactory.MarketAlreadyExists.selector, MARKET_ID));

        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));
    }

    function testEmptyMarketIdReverts() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.expectRevert(SignalArcMarketFactory.EmptyMarketId.selector);

        factory.createMarket("", QUESTION, closeTimestamp(), RESOLVER, address(token));
    }

    function testEmptyQuestionReverts() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.expectRevert(SignalArcMarketFactory.EmptyQuestion.selector);

        factory.createMarket(MARKET_ID, "", closeTimestamp(), RESOLVER, address(token));
    }

    function testZeroResolverReverts() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        vm.expectRevert(SignalArcMarketFactory.InvalidResolver.selector);

        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), address(0), address(token));
    }

    function testZeroCollateralTokenReverts() external {
        (SignalArcMarketFactory factory,) = createFactoryAndToken();

        vm.expectRevert(SignalArcMarketFactory.InvalidCollateralToken.selector);

        factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(0));
    }

    function testFactoryCanDeployTwoMarketsWithDifferentIdsAndAddresses() external {
        (SignalArcMarketFactory factory, MockUSDC token) = createFactoryAndToken();

        address firstMarket = factory.createMarket(MARKET_ID, QUESTION, closeTimestamp(), RESOLVER, address(token));
        address secondMarket = factory.createMarket(
            SECOND_MARKET_ID, SECOND_QUESTION, closeTimestamp() + 1 days, RESOLVER, address(token)
        );

        assertTrue(firstMarket != secondMarket);
        assertEq(factory.marketById(MARKET_ID), firstMarket);
        assertEq(factory.marketById(SECOND_MARKET_ID), secondMarket);
        assertEq(factory.marketCount(), 2);
        assertEq(factory.allMarkets(0), firstMarket);
        assertEq(factory.allMarkets(1), secondMarket);
        assertTrue(factory.isMarket(firstMarket));
        assertTrue(factory.isMarket(secondMarket));
    }

    function createFactoryAndToken() private returns (SignalArcMarketFactory, MockUSDC) {
        return (new SignalArcMarketFactory(), new MockUSDC());
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
