部署说明
1. 环境要求

Node.js 16+  
Go 1.21+  
Solidity编译器

2. 智能合约部署
```shell
# 安装Hardhat或Truffle
npm install -g hardhat

# 编译合约
npx hardhat compile

# 部署合约
npx hardhat run scripts/deploy.js --network <network_name>
```

3. 后端部署
```shell

# 安装依赖
go mod tidy

# 设置环境变量 (私钥已写死，只需设置合约地址)
export CONTRACT_ADDRESS="deployed_vault_contract_address"
export SCOS_ADDRESS="deployed_scos_contract_address"

# 运行服务器

go run main.go

```

4. 前端部署
```
npx http-server -p 3000
```

合约部署脚本 (deploy.js)
```
npx hardhat run scripts/deploy.js --network reddio

npx hardhat run scripts/deploy.js --network scrollSepolia

```


API文档
```shell
1. 获取Stock价格
   GET /api/stock/{symbol}/price
   Response: {
   "symbol": "STOCK",
   "price": 100.0,
   "updated_at": "2024-01-01T00:00:00Z"
   }
2. 更新Stock价格 (管理员)
   POST /api/stock/{symbol}/price
   Body: {
   "price": 120.0
   }
3. 获取用户SCOS余额
   GET /api/user/{address}/scos
   Response: {
   "address": "0x...",
   "scos_balance": 1000.0,
   "active_stakes": 2
   }
4. 质押Stock
   POST /api/stake
   Body: {
   "user_address": "0x...",
   "token_address": "0x...",
   "chain": "ethereum",
   "amount": "100",
   "contract_address": "0x..."
   }
5. 赎回Stock
   POST /api/redeem
   Body: {
   "user_address": "0x...",
   "token_address": "0x...",
   "chain": "ethereum",
   "contract_address": "0x..."
   }
6. 买入Stock
   POST /api/buy
   Body: {
   "user_address": "0x...",
   "token_address": "0x...",
   "chain": "ethereum",
   "amount": "100",
   "contract_address": "0x..."
   }
7. 卖出Stock
   POST /api/sell
   Body: {
   "user_address": "0x...",
   "token_address": "0x...",
   "chain": "ethereum",
   "amount": "100",
   "contract_address": "0x..."
   }
```