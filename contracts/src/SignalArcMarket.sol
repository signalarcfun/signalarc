// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IERC20Like {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

contract SignalArcMarket {
    enum MarketStatus {
        Draft,
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
    address public resolver;
    IERC20Like public immutable collateralToken;
    MarketStatus public status;
    Outcome public winningOutcome;
    mapping(address user => uint256 amount) public yesPositions;
    mapping(address user => uint256 amount) public noPositions;
    mapping(address user => bool claimed) public hasClaimed;
    uint256 public totalYes;
    uint256 public totalNo;
    uint256 public totalCollateral;

    event MarketCreated(string question, uint256 closeTimestamp, address resolver);
    event MarketClosed();
    event MarketCancelled();
    event MarketResolved(Outcome winningOutcome);
    event PositionOpened(address indexed user, Outcome indexed side, uint256 amount);
    event PayoutClaimed(address indexed user, uint256 amount);
    event RefundClaimed(address indexed user, uint256 amount);

    error EmptyQuestion();
    error InvalidCloseTimestamp();
    error InvalidResolver();
    error InvalidCollateralToken();
    error MarketNotOpen();
    error MarketAlreadyFinalized();
    error UnauthorizedResolver();
    error InvalidOutcome();
    error InvalidAmount();
    error InvalidSide();
    error CollateralTransferFailed();
    error MarketNotResolved();
    error NothingToClaim();
    error AlreadyClaimed();
    error PayoutTransferFailed();
    error NoWinningStake();

    constructor(string memory question_, uint256 closeTimestamp_, address resolver_, address collateralToken_) {
        if (bytes(question_).length == 0) {
            revert EmptyQuestion();
        }

        if (closeTimestamp_ <= block.timestamp) {
            revert InvalidCloseTimestamp();
        }

        if (resolver_ == address(0)) {
            revert InvalidResolver();
        }

        if (collateralToken_ == address(0)) {
            revert InvalidCollateralToken();
        }

        question = question_;
        closeTimestamp = closeTimestamp_;
        resolver = resolver_;
        collateralToken = IERC20Like(collateralToken_);
        status = MarketStatus.Open;
        winningOutcome = Outcome.None;

        emit MarketCreated(question_, closeTimestamp_, resolver_);
    }

    function isOpen() external view returns (bool) {
        return _isOpen();
    }

    function openPosition(Outcome side, uint256 amount) external {
        if (!_isOpen()) {
            revert MarketNotOpen();
        }

        if (side != Outcome.Yes && side != Outcome.No) {
            revert InvalidSide();
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

        emit PositionOpened(msg.sender, side, amount);
    }

    function claimableAmount(address user) public view returns (uint256) {
        if (status == MarketStatus.Resolved) {
            if (winningOutcome == Outcome.Yes) {
                if (totalYes == 0) {
                    return 0;
                }

                return yesPositions[user] * totalCollateral / totalYes;
            }

            if (winningOutcome == Outcome.No) {
                if (totalNo == 0) {
                    return 0;
                }

                return noPositions[user] * totalCollateral / totalNo;
            }

            return 0;
        }

        if (status == MarketStatus.Cancelled) {
            return yesPositions[user] + noPositions[user];
        }

        return 0;
    }

    function claim() external {
        if (status != MarketStatus.Resolved && status != MarketStatus.Cancelled) {
            revert MarketNotResolved();
        }

        if (hasClaimed[msg.sender]) {
            revert AlreadyClaimed();
        }

        uint256 amount = claimableAmount(msg.sender);
        if (amount == 0) {
            revert NothingToClaim();
        }

        hasClaimed[msg.sender] = true;

        bool transferred = collateralToken.transfer(msg.sender, amount);
        if (!transferred) {
            revert PayoutTransferFailed();
        }

        if (status == MarketStatus.Cancelled) {
            emit RefundClaimed(msg.sender, amount);
        } else {
            emit PayoutClaimed(msg.sender, amount);
        }
    }

    function closeMarket() external {
        if (status != MarketStatus.Open || block.timestamp < closeTimestamp) {
            revert MarketNotOpen();
        }

        status = MarketStatus.Closed;

        emit MarketClosed();
    }

    function cancelMarket() external {
        if (msg.sender != resolver) {
            revert UnauthorizedResolver();
        }

        if (status == MarketStatus.Resolved || status == MarketStatus.Cancelled) {
            revert MarketAlreadyFinalized();
        }

        status = MarketStatus.Cancelled;

        emit MarketCancelled();
    }

    function resolve(Outcome winningOutcome_) external {
        if (msg.sender != resolver) {
            revert UnauthorizedResolver();
        }

        if (status != MarketStatus.Closed) {
            revert MarketNotOpen();
        }

        if (winningOutcome_ != Outcome.Yes && winningOutcome_ != Outcome.No) {
            revert InvalidOutcome();
        }

        if (winningOutcome_ == Outcome.Yes && totalYes == 0) {
            revert NoWinningStake();
        }

        if (winningOutcome_ == Outcome.No && totalNo == 0) {
            revert NoWinningStake();
        }

        status = MarketStatus.Resolved;
        winningOutcome = winningOutcome_;

        emit MarketResolved(winningOutcome_);
    }

    function _isOpen() private view returns (bool) {
        return status == MarketStatus.Open && block.timestamp < closeTimestamp;
    }
}
