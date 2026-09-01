// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IAgentCollateralToken {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

contract SignalArcAgentMarket {
    enum MarketStatus {
        Open,
        Closed,
        Resolved,
        Cancelled
    }

    enum Outcome {
        None,
        Yes,
        No
    }

    string public question;
    uint256 public closeTimestamp;
    address public admin;
    address public resolver;
    IAgentCollateralToken public immutable collateralToken;
    MarketStatus public status;
    Outcome public winningOutcome;
    mapping(address user => uint256 amount) public yesPositions;
    mapping(address user => uint256 amount) public noPositions;
    mapping(address user => bool claimed) public hasClaimed;
    uint256 public totalYes;
    uint256 public totalNo;
    uint256 public totalCollateral;

    event AgentMarketCreated(string question, uint256 closeTimestamp, address indexed admin, address indexed resolver);
    event AgentMarketClosed(address indexed caller);
    event AgentMarketCancelled(address indexed caller);
    event AgentMarketResolved(Outcome winningOutcome);
    event AgentPositionOpened(address indexed user, Outcome indexed side, uint256 amount);
    event AgentPayoutClaimed(address indexed user, uint256 amount);
    event AgentRefundClaimed(address indexed user, uint256 amount);

    error EmptyQuestion();
    error InvalidCloseTimestamp();
    error InvalidAdmin();
    error InvalidResolver();
    error InvalidCollateralToken();
    error MarketNotOpen();
    error MarketNotClosed();
    error MarketNotResolved();
    error MarketNotCancelled();
    error MarketAlreadyFinalized();
    error UnauthorizedAdminOrResolver();
    error UnauthorizedResolver();
    error InvalidOutcome();
    error InvalidAmount();
    error CollateralTransferFailed();
    error PayoutTransferFailed();
    error RefundTransferFailed();
    error NothingToClaim();
    error AlreadyClaimed();

    constructor(
        string memory question_,
        uint256 closeTimestamp_,
        address admin_,
        address resolver_,
        address collateralToken_
    ) {
        if (bytes(question_).length == 0) {
            revert EmptyQuestion();
        }

        if (closeTimestamp_ <= block.timestamp) {
            revert InvalidCloseTimestamp();
        }

        if (admin_ == address(0)) {
            revert InvalidAdmin();
        }

        if (resolver_ == address(0)) {
            revert InvalidResolver();
        }

        if (collateralToken_ == address(0)) {
            revert InvalidCollateralToken();
        }

        question = question_;
        closeTimestamp = closeTimestamp_;
        admin = admin_;
        resolver = resolver_;
        collateralToken = IAgentCollateralToken(collateralToken_);
        status = MarketStatus.Open;
        winningOutcome = Outcome.None;

        emit AgentMarketCreated(question_, closeTimestamp_, admin_, resolver_);
    }

    function isOpen() external view returns (bool) {
        return _isOpen();
    }

    function buyYes(uint256 amount) external {
        _buy(Outcome.Yes, amount);
    }

    function buyNo(uint256 amount) external {
        _buy(Outcome.No, amount);
    }

    function closeMarket() external {
        if (!_isAdminOrResolver()) {
            revert UnauthorizedAdminOrResolver();
        }

        if (status != MarketStatus.Open || block.timestamp < closeTimestamp) {
            revert MarketNotOpen();
        }

        status = MarketStatus.Closed;

        emit AgentMarketClosed(msg.sender);
    }

    function cancelMarket() external {
        if (!_isAdminOrResolver()) {
            revert UnauthorizedAdminOrResolver();
        }

        if (status == MarketStatus.Resolved || status == MarketStatus.Cancelled) {
            revert MarketAlreadyFinalized();
        }

        status = MarketStatus.Cancelled;

        emit AgentMarketCancelled(msg.sender);
    }

    function resolve(Outcome winningOutcome_) external {
        if (msg.sender != resolver) {
            revert UnauthorizedResolver();
        }

        if (status != MarketStatus.Closed) {
            revert MarketNotClosed();
        }

        if (winningOutcome_ != Outcome.Yes && winningOutcome_ != Outcome.No) {
            revert InvalidOutcome();
        }

        status = MarketStatus.Resolved;
        winningOutcome = winningOutcome_;

        emit AgentMarketResolved(winningOutcome_);
    }

    function claimRefund() external {
        if (status != MarketStatus.Cancelled) {
            revert MarketNotCancelled();
        }

        uint256 amount = yesPositions[msg.sender] + noPositions[msg.sender];
        _markClaimedAndValidate(amount);

        bool transferred = collateralToken.transfer(msg.sender, amount);
        if (!transferred) {
            revert RefundTransferFailed();
        }

        emit AgentRefundClaimed(msg.sender, amount);
    }

    function claimPayout() external {
        if (status != MarketStatus.Resolved) {
            revert MarketNotResolved();
        }

        uint256 amount;
        if (winningOutcome == Outcome.Yes) {
            amount = yesPositions[msg.sender];
        } else if (winningOutcome == Outcome.No) {
            amount = noPositions[msg.sender];
        }

        _markClaimedAndValidate(amount);

        bool transferred = collateralToken.transfer(msg.sender, amount);
        if (!transferred) {
            revert PayoutTransferFailed();
        }

        emit AgentPayoutClaimed(msg.sender, amount);
    }

    function claimableRefund(address user) external view returns (uint256) {
        if (status != MarketStatus.Cancelled) {
            return 0;
        }

        return yesPositions[user] + noPositions[user];
    }

    function claimablePayout(address user) external view returns (uint256) {
        if (status != MarketStatus.Resolved) {
            return 0;
        }

        if (winningOutcome == Outcome.Yes) {
            return yesPositions[user];
        }

        if (winningOutcome == Outcome.No) {
            return noPositions[user];
        }

        return 0;
    }

    function _buy(Outcome side, uint256 amount) private {
        if (!_isOpen()) {
            revert MarketNotOpen();
        }

        if (amount == 0) {
            revert InvalidAmount();
        }

        bool transferred = collateralToken.transferFrom(msg.sender, address(this), amount);
        if (!transferred) {
            revert CollateralTransferFailed();
        }

        if (side == Outcome.Yes) {
            yesPositions[msg.sender] += amount;
            totalYes += amount;
        } else {
            noPositions[msg.sender] += amount;
            totalNo += amount;
        }

        totalCollateral += amount;

        emit AgentPositionOpened(msg.sender, side, amount);
    }

    function _markClaimedAndValidate(uint256 amount) private {
        if (hasClaimed[msg.sender]) {
            revert AlreadyClaimed();
        }

        if (amount == 0) {
            revert NothingToClaim();
        }

        hasClaimed[msg.sender] = true;
    }

    function _isOpen() private view returns (bool) {
        return status == MarketStatus.Open && block.timestamp < closeTimestamp;
    }

    function _isAdminOrResolver() private view returns (bool) {
        return msg.sender == admin || msg.sender == resolver;
    }
}
