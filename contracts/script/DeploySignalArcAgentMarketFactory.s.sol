// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/agent/SignalArcAgentMarketFactory.sol";

contract DeploySignalArcAgentMarketFactory {
    function run() external returns (SignalArcAgentMarketFactory factory) {
        factory = new SignalArcAgentMarketFactory();
    }
}
