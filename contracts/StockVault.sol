// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract StockVault is Ownable, ReentrancyGuard {
    struct StakeInfo {
        uint256 amount;
        uint256 borrowedSCOS;
        uint256 timestamp;
        bool active;
    }

    mapping(address => mapping(address => StakeInfo)) public userStakes; // user => token => stake
    mapping(address => bool) public supportedTokens;

    IERC20 public scosToken;
    address public treasury;

    event StockStaked(address indexed user, address indexed token, uint256 amount, uint256 borrowedSCOS);
    event StockUnstaked(address indexed user, address indexed token, uint256 amount);
    event Liquidated(address indexed user, address indexed token, uint256 amount);

    constructor(address _scosToken, address _treasury) {
        scosToken = IERC20(_scosToken);
        treasury = _treasury;
    }

    function addSupportedToken(address token) external onlyOwner {
        supportedTokens[token] = true;
    }

    function stakeStock(address token, uint256 amount, uint256 scosAmount) external nonReentrant {
        require(supportedTokens[token], "Token not supported");
        require(amount > 0, "Amount must be greater than 0");

        IERC20(token).transferFrom(msg.sender, address(this), amount);

        userStakes[msg.sender][token] = StakeInfo({
        amount: amount,
        borrowedSCOS: scosAmount,
        timestamp: block.timestamp,
        active: true
        });

        emit StockStaked(msg.sender, token, amount, scosAmount);
    }

    function unstakeStock(address token) external nonReentrant {
        StakeInfo storage stake = userStakes[msg.sender][token];
        require(stake.active, "No active stake");

        uint256 amount = stake.amount;
        stake.active = false;
        stake.amount = 0;
        stake.borrowedSCOS = 0;

        IERC20(token).transfer(msg.sender, amount);

        emit StockUnstaked(msg.sender, token, amount);
    }

    function liquidate(address user, address token) external onlyOwner {
        StakeInfo storage stake = userStakes[user][token];
        require(stake.active, "No active stake");

        uint256 amount = stake.amount;
        stake.active = false;
        stake.amount = 0;
        stake.borrowedSCOS = 0;

        IERC20(token).transfer(treasury, amount);

        emit Liquidated(user, token, amount);
    }

    function getStakeInfo(address user, address token) external view returns (StakeInfo memory) {
        return userStakes[user][token];
    }
}