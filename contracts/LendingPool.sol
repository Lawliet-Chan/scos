// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

// SCOS稳定币合约
contract SCOS is ERC20, Ownable {
    constructor() ERC20("SCOS Stablecoin", "SCOS") {}

    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }

    function burn(uint256 amount) external {
        _burn(msg.sender, amount);
    }

    function burnFrom(address account, uint256 amount) external onlyOwner {
        _burn(account, amount);
    }
}

// 借贷池合约
contract LendingPool is Ownable, ReentrancyGuard {
    SCOS public immutable scos;
    
    struct Position {
        uint256 collateralAmount;  // 质押的token数量
        uint256 borrowedAmount;    // 借贷的SCOS数量
        uint256 lastPrice;         // 质押时的token价格
        bool isActive;             // 头寸是否活跃
    }
    
    // 用户地址 => token地址 => 质押头寸
    mapping(address => mapping(address => Position)) public positions;
    
    // 支持的抵押品token
    mapping(address => bool) public supportedTokens;
    
    // 常数
    uint256 public constant COLLATERAL_RATIO = 140; // 140%抵押率
    uint256 public constant LIQUIDATION_THRESHOLD = 125; // 125%清算阈值
    
    event Deposit(address indexed user, address indexed token, uint256 amount, uint256 borrowed);
    event Withdraw(address indexed user, address indexed token, uint256 amount, uint256 repaid);
    event Liquidation(address indexed user, address indexed token, uint256 amount);
    
    constructor(address _scos) {
        scos = SCOS(_scos);
    }
    
    // 添加支持的token
    function addSupportedToken(address token) external onlyOwner {
        supportedTokens[token] = true;
    }
    
    // 质押token并借贷SCOS
    function deposit(
        address token,
        uint256 amount,
        uint256 tokenPrice
    ) external nonReentrant {
        require(supportedTokens[token], "Token not supported");
        require(amount > 0, "Amount must be greater than 0");
        require(!positions[msg.sender][token].isActive, "Position already exists");
        
        // 转入token
        IERC20(token).transferFrom(msg.sender, address(this), amount);
        
        // 计算可借贷的SCOS数量 (token数量 * token价格 / 1.4)
        uint256 borrowAmount = (amount * tokenPrice * 1e18) / (COLLATERAL_RATIO * 1e18 / 100);
        
        // 铸造SCOS给用户
        scos.mint(msg.sender, borrowAmount);
        
        // 记录头寸
        positions[msg.sender][token] = Position({
            collateralAmount: amount,
            borrowedAmount: borrowAmount,
            lastPrice: tokenPrice,
            isActive: true
        });
        
        emit Deposit(msg.sender, token, amount, borrowAmount);
    }
    
    // 赎回token
    function withdraw(address token) external nonReentrant {
        Position storage position = positions[msg.sender][token];
        require(position.isActive, "No active position");
        
        // 用户需要先归还SCOS
        require(scos.balanceOf(msg.sender) >= position.borrowedAmount, "Insufficient SCOS balance");
        
        // 销毁用户的SCOS
        scos.burnFrom(msg.sender, position.borrowedAmount);
        
        // 返还抵押的token
        IERC20(token).transfer(msg.sender, position.collateralAmount);
        
        emit Withdraw(msg.sender, token, position.collateralAmount, position.borrowedAmount);
        
        // 清除头寸
        delete positions[msg.sender][token];
    }
    
    // 清算（只有合约所有者可以调用）
    function liquidate(address user, address token, uint256 currentPrice) external onlyOwner {
        Position storage position = positions[user][token];
        require(position.isActive, "No active position");
        
        // 检查是否达到清算条件（token价格下跌超过25%）
        uint256 priceDropThreshold = (position.lastPrice * 75) / 100; // 下跌25%的阈值
        require(currentPrice <= priceDropThreshold, "Price drop not sufficient for liquidation");
        
        // 计算当前抵押率
        uint256 currentCollateralValue = position.collateralAmount * currentPrice;
        uint256 currentRatio = (currentCollateralValue * 100) / position.borrowedAmount;
        require(currentRatio <= LIQUIDATION_THRESHOLD, "Position is safe");
        
        // 清算：将抵押品转给合约所有者
        IERC20(token).transfer(owner(), position.collateralAmount);
        
        emit Liquidation(user, token, position.collateralAmount);
        
        // 清除头寸
        delete positions[user][token];
    }
    
    // 购买token（使用SCOS）
    function buyToken(address token, uint256 tokenAmount, uint256 tokenPrice) external nonReentrant {
        require(supportedTokens[token], "Token not supported");
        
        uint256 scosNeeded = tokenAmount * tokenPrice;
        require(scos.balanceOf(msg.sender) >= scosNeeded, "Insufficient SCOS balance");
        
        // 检查合约是否有足够的token
        uint256 contractTokenBalance = IERC20(token).balanceOf(address(this));
        if (contractTokenBalance < tokenAmount) {
            // 如果token不足，增发相应数量的SCOS给合约所有者用于购买token
            uint256 additionalScos = (tokenAmount - contractTokenBalance) * tokenPrice;
            scos.mint(owner(), additionalScos);
        }
        
        // 销毁用户的SCOS
        scos.burnFrom(msg.sender, scosNeeded);
        
        // 转token给用户
        IERC20(token).transfer(msg.sender, tokenAmount);
    }
    
    // 获取用户头寸信息
    function getUserPosition(address user, address token) external view returns (Position memory) {
        return positions[user][token];
    }
    
    // 紧急提取函数（仅所有者）
    function emergencyWithdraw(address token, uint256 amount) external onlyOwner {
        IERC20(token).transfer(owner(), amount);
    }
}
