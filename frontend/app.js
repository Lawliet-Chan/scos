class DeFiApp {
    constructor() {
        this.web3Helper = new Web3Helper();
        this.apiBase = 'http://localhost:8080/api';
        this.vaultContractAddress = '0x0124e835BdE149aD885b765Bb8BF6f63735Fc4db'; // 替换为实际StockVault合约地址
        this.init();
    }

    async init() {
        this.setupEventListeners();
        await this.loadAllStockPrices();
        //await this.loadInitialData();
    }

    setupEventListeners() {
        // 连接钱包
        document.getElementById('connectWallet').addEventListener('click', async () => {
            await this.connectWallet();
        });

        // 刷新价格
        document.getElementById('refreshPrice').addEventListener('click', async () => {
            await this.loadStockPrice();
        });

        // 刷新余额
        document.getElementById('refreshBalance').addEventListener('click', async () => {
            await this.loadUserBalance();
        });

        // 股票选择变化
        document.getElementById('stockSelector').addEventListener('change', async () => {
            await this.loadStockPrice();
            this.calculateBorrowableAmount();
        });

        // 质押股票选择变化
        document.getElementById('stakeStockSelector').addEventListener('change', () => {
            this.calculateBorrowableAmount();
        });
        document.getElementById('stakeAmount').addEventListener('input', () => {
            this.calculateBorrowableAmount();
        });

        // 质押按钮
        document.getElementById('stakeBtn').addEventListener('click', async () => {
            await this.stakeStock();
        });

        // 买入按钮
        document.getElementById('buyBtn').addEventListener('click', async () => {
            await this.buyStock();
        });

        // 卖出按钮
        document.getElementById('sellBtn').addEventListener('click', async () => {
            await this.sellStock();
        });

        // 赎回按钮
        document.getElementById('redeemBtn').addEventListener('click', async () => {
            await this.redeemStock();
        });
    }

    async loadAllStockPrices() {
        try {
            const response = await fetch(`${this.apiBase}/stocks/prices`);
            const data = await response.json();

            if (response.ok) {
                this.stockPrices = {};
                data.stocks.forEach(stock => {
                    this.stockPrices[stock.symbol] = stock.price;
                });
                this.updateStockSelectors();
            }
        } catch (error) {
            console.error('加载股票价格失败:', error);
        }
    }

    updateStockSelectors() {
        // 更新质押股票选择器
        const stakeSelector = document.getElementById('stakeStockSelector');
        const tradeSelector = document.getElementById('tradeStockSelector');

        stakeSelector.innerHTML = '';
        tradeSelector.innerHTML = '';

        const stockIcons = {
            'APPLE': '🍎',
            'GOOGLE': '🔍',
            'MICROSOFT': '🪟'
        };

        Object.entries(this.stockPrices).forEach(([symbol, price]) => {
            const icon = stockIcons[symbol] || '📈';
            const option1 = new Option(`${icon} ${symbol} - ${price.toFixed(2)}`, symbol);
            const option2 = new Option(`${icon} ${symbol} - ${price.toFixed(2)}`, symbol);

            stakeSelector.appendChild(option1);
            tradeSelector.appendChild(option2);
        });
    }

    async connectWallet() {
        const connected = await this.web3Helper.connectWallet();
        if (connected) {
            await this.loadUserBalance();
            this.showMessage('钱包连接成功!', 'success');
        } else {
            this.showMessage('钱包连接失败!', 'error');
        }
    }

    async loadStockPrice() {
        try {
            const response = await fetch(`${this.apiBase}/stock/STOCK/price`);
            const data = await response.json();

            if (response.ok) {
                document.getElementById('stockPrice').textContent = `${data.price.toFixed(2)}`;
                this.calculateBorrowableAmount();
            } else {
                this.showMessage('获取价格失败', 'error');
            }
        } catch (error) {
            console.error('加载股票价格失败:', error);
            this.showMessage('网络错误', 'error');
        }
    }

    async loadUserBalance() {
        if (!this.web3Helper.isConnected()) return;

        try {
            const response = await fetch(`${this.apiBase}/user/${this.web3Helper.account}/scos`);
            const data = await response.json();

            if (response.ok) {
                document.getElementById('scosBalance').textContent = `${data.scos_balance.toFixed(6)} SCOS`;
            }
        } catch (error) {
            console.error('加载用户余额失败:', error);
        }
    }

    calculateBorrowableAmount() {
        const amountInput = document.getElementById('stakeAmount');
        const priceElement = document.getElementById('stockPrice');
        const borrowableElement = document.getElementById('borrowableAmount');

        const amount = parseFloat(amountInput.value) || 0;
        const priceText = priceElement.textContent.replace(',', '');
        const price = parseFloat(priceText) || 0;

        // 计算可借贷金额 (数量/140% × 价格)
        const borrowable = (amount / 1.4) * price;
        borrowableElement.textContent = `${borrowable.toFixed(6)} SCOS`;
    }

    async stakeStock() {
        if (!this.web3Helper.isConnected()) {
            this.showMessage('请先连接钱包', 'error');
            return;
        }

        const chain = document.getElementById('chainSelect').value;
        const stockAddress = document.getElementById('stockAddress').value;
        const amount = document.getElementById('stakeAmount').value;

        if (!stockAddress || !amount) {
            this.showMessage('请填写完整信息', 'error');
            return;
        }

        try {
            this.showLoading();

            // 首先批准代币转账
            const stockContract = this.web3Helper.getContract(stockAddress, ERC20_ABI);
            const amountWei = this.web3Helper.toWei(amount, 6); // 假设6位小数

            console.log("stockContract____", stockContract, this.vaultContractAddress, amountWei);
            await this.web3Helper.sendTransaction(stockContract, 'approve', [this.vaultContractAddress, amountWei]);

            // 调用后端API进行质押
            const response = await fetch(`${this.apiBase}/stake`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user_address: this.web3Helper.account,
                    token_address: stockAddress,
                    chain: chain,
                    amount: amount,
                    contract_address: this.vaultContractAddress
                })
            });

            const data = await response.json();
            this.hideLoading();

            if (response.ok) {
                this.showMessage(`质押成功! 借贷 ${data.scos_borrowed} SCOS`, 'success');
                await this.loadUserBalance();
                this.clearStakeForm();
            } else {
                this.showMessage(data.error || '质押失败', 'error');
            }
        } catch (error) {
            this.hideLoading();
            console.error('质押失败:', error);
            this.showMessage('质押失败: ' + error.message, 'error');
        }
    }

    async buyStock() {
        if (!this.web3Helper.isConnected()) {
            this.showMessage('请先连接钱包', 'error');
            return;
        }

        const stockAddress = document.getElementById('tradeStockAddress').value;
        const amount = document.getElementById('tradeAmount').value;
        const chain = document.getElementById('chainSelect').value;

        if (!stockAddress || !amount) {
            this.showMessage('请填写完整信息', 'error');
            return;
        }

        try {
            this.showLoading();

            const response = await fetch(`${this.apiBase}/buy`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user_address: this.web3Helper.account,
                    token_address: stockAddress,
                    chain: chain,
                    amount: amount,
                    scos_amount: "0", // 计算所需SCOS数量
                    contract_address: this.vaultContractAddress
                })
            });

            const data = await response.json();
            this.hideLoading();

            if (response.ok) {
                this.showMessage('买入订单已提交!', 'success');
                document.getElementById('tradeAmount').value = '';
            } else {
                this.showMessage(data.error || '买入失败', 'error');
            }
        } catch (error) {
            this.hideLoading();
            console.error('买入失败:', error);
            this.showMessage('买入失败: ' + error.message, 'error');
        }
    }

    async sellStock() {
        if (!this.web3Helper.isConnected()) {
            this.showMessage('请先连接钱包', 'error');
            return;
        }

        const stockAddress = document.getElementById('tradeStockAddress').value;
        const amount = document.getElementById('tradeAmount').value;
        const chain = document.getElementById('chainSelect').value;

        if (!stockAddress || !amount) {
            this.showMessage('请填写完整信息', 'error');
            return;
        }

        try {
            this.showLoading();

            const response = await fetch(`${this.apiBase}/sell`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user_address: this.web3Helper.account,
                    token_address: stockAddress,
                    chain: chain,
                    amount: amount,
                    contract_address: this.vaultContractAddress
                })
            });

            const data = await response.json();
            this.hideLoading();

            if (response.ok) {
                this.showMessage('卖出订单已提交!', 'success');
                document.getElementById('tradeAmount').value = '';
            } else {
                this.showMessage(data.error || '卖出失败', 'error');
            }
        } catch (error) {
            this.hideLoading();
            console.error('卖出失败:', error);
            this.showMessage('卖出失败: ' + error.message, 'error');
        }
    }

    async redeemStock() {
        if (!this.web3Helper.isConnected()) {
            this.showMessage('请先连接钱包', 'error');
            return;
        }

        const stockAddress = document.getElementById('redeemStockAddress').value;
        const chain = document.getElementById('chainSelect').value;

        if (!stockAddress) {
            this.showMessage('请填写Stock合约地址', 'error');
            return;
        }

        try {
            this.showLoading();

            const response = await fetch(`${this.apiBase}/redeem`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user_address: this.web3Helper.account,
                    token_address: stockAddress,
                    chain: chain,
                    contract_address: this.vaultContractAddress
                })
            });

            const data = await response.json();
            this.hideLoading();

            if (response.ok) {
                this.showMessage('赎回成功!', 'success');
                await this.loadUserBalance();
                document.getElementById('redeemStockAddress').value = '';
            } else {
                this.showMessage(data.error || '赎回失败', 'error');
            }
        } catch (error) {
            this.hideLoading();
            console.error('赎回失败:', error);
            this.showMessage('赎回失败: ' + error.message, 'error');
        }
    }

    clearStakeForm() {
        document.getElementById('stockAddress').value = '';
        document.getElementById('stakeAmount').value = '';
        document.getElementById('borrowableAmount').textContent = '0.00 SCOS';
    }

    showLoading() {
        document.getElementById('loadingModal').style.display = 'flex';
    }

    hideLoading() {
        document.getElementById('loadingModal').style.display = 'none';
    }

    showMessage(message, type = 'info') {
        // 创建消息提示
        const messageDiv = document.createElement('div');
        messageDiv.className = `message message-${type}`;
        messageDiv.textContent = message;

        // 样式
        Object.assign(messageDiv.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '15px 20px',
            borderRadius: '10px',
            color: 'white',
            fontWeight: 'bold',
            zIndex: '1001',
            minWidth: '250px',
            textAlign: 'center'
        });

        // 根据类型设置背景色
        const colors = {
            success: '#4CAF50',
            error: '#f44336',
            warning: '#ff9800',
            info: '#2196F3'
        };
        messageDiv.style.background = colors[type] || colors.info;

        document.body.appendChild(messageDiv);

        // 3秒后自动移除
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, 3000);
    }
}

// 初始化应用
window.addEventListener('DOMContentLoaded', () => {
    const app = new DeFiApp();
});