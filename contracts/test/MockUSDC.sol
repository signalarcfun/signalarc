// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

// Test-only USDC-like token for local collateral assumption tests.
contract MockUSDC {
    string public constant name = "Mock USDC";
    string public constant symbol = "mUSDC";
    uint8 public constant decimals = 6;

    mapping(address account => uint256 balance) public balanceOf;
    mapping(address owner => mapping(address spender => uint256 amount)) public allowance;

    event Transfer(address indexed from, address indexed to, uint256 amount);
    event Approval(address indexed owner, address indexed spender, uint256 amount);

    error InvalidRecipient();
    error InsufficientBalance();
    error InsufficientAllowance();

    function mint(address to, uint256 amount) external {
        if (to == address(0)) {
            revert InvalidRecipient();
        }

        balanceOf[to] += amount;

        emit Transfer(address(0), to, amount);
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;

        emit Approval(msg.sender, spender, amount);

        return true;
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        if (to == address(0)) {
            revert InvalidRecipient();
        }

        if (balanceOf[msg.sender] < amount) {
            revert InsufficientBalance();
        }

        unchecked {
            balanceOf[msg.sender] -= amount;
        }
        balanceOf[to] += amount;

        emit Transfer(msg.sender, to, amount);

        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        if (to == address(0)) {
            revert InvalidRecipient();
        }

        if (balanceOf[from] < amount) {
            revert InsufficientBalance();
        }

        uint256 currentAllowance = allowance[from][msg.sender];
        if (currentAllowance < amount) {
            revert InsufficientAllowance();
        }

        unchecked {
            allowance[from][msg.sender] = currentAllowance - amount;
            balanceOf[from] -= amount;
        }
        balanceOf[to] += amount;

        emit Transfer(from, to, amount);

        return true;
    }
}
