// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./SignalArcMarket.sol";

contract SignalArcMarketFactory {
    mapping(string => address) public marketById;
    mapping(address => bool) public isMarket;
    address[] public allMarkets;

    event MarketDeployed(
        string indexed marketId,
        address indexed market,
        address indexed creator,
        address resolver,
        address collateralToken,
        uint256 closeTimestamp,
        string question
    );

    error EmptyMarketId();
    error MarketAlreadyExists(string marketId);
    error EmptyQuestion();
    error InvalidResolver();
    error InvalidCollateralToken();

    function createMarket(
        string calldata marketId,
        string calldata question,
        uint256 closeTimestamp,
        address resolver,
        address collateralToken
    ) external returns (address market) {
        if (bytes(marketId).length == 0) {
            revert EmptyMarketId();
        }

        if (marketById[marketId] != address(0)) {
            revert MarketAlreadyExists(marketId);
        }

        if (bytes(question).length == 0) {
            revert EmptyQuestion();
        }

        if (resolver == address(0)) {
            revert InvalidResolver();
        }

        if (collateralToken == address(0)) {
            revert InvalidCollateralToken();
        }

        market = address(new SignalArcMarket(question, closeTimestamp, resolver, collateralToken));
        marketById[marketId] = market;
        isMarket[market] = true;
        allMarkets.push(market);

        emit MarketDeployed(marketId, market, msg.sender, resolver, collateralToken, closeTimestamp, question);
    }

    function marketCount() external view returns (uint256) {
        return allMarkets.length;
    }
}
