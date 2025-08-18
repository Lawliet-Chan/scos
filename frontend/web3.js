class Web3Helper {
    constructor() {
        this.web3 = null;
        this.account = null;
        this.chainId = null;
        this.contracts = {};
    }

    async connectWallet() {
        if (typeof window.ethereum !== 'undefined') {
            try {
                this.web3 = new Web3(window.ethereum);

                // 请求账户访问
                const accounts = await window.ethereum.request({
                    method: 'eth_requestAccounts'
                });

                this.account = accounts[0];
                this.chainId = await this.web3.eth.getChainId();

                // 监听账户切换
                window.ethereum.on('accountsChanged', (accounts) => {
                    if (accounts.length === 0) {
                        this.disconnect();
                    } else {
                        this.account = accounts[0];
                        this.updateWalletDisplay();
                    }
                });

                // 监听网络切换
                window.ethereum.on('chainChanged', (chainId) => {
                    this.chainId = parseInt(chainId, 16);
                    this.updateWalletDisplay();
                });

                this.updateWalletDisplay();
                return true;
            } catch (error) {
                console.error('连接钱包失败:', error);
                return false;
            }
        } else {
            alert('请安装MetaMask钱包!');
            return false;
        }
    }

    disconnect() {
        this.account = null;
        this.chainId = null;
        this.web3 = null;
        this.updateWalletDisplay();
    }

    updateWalletDisplay() {
        const connectBtn = document.getElementById('connectWallet');
        const walletInfo = document.getElementById('walletInfo');
        const walletAddress = document.getElementById('walletAddress');
        const networkInfo = document.getElementById('networkInfo');

        if (this.account) {
            connectBtn.style.display = 'none';
            walletInfo.style.display = 'block';
            walletAddress.textContent = `${this.account.slice(0, 6)}...${this.account.slice(-4)}`;
            networkInfo.textContent = this.getNetworkName(this.chainId);
        } else {
            connectBtn.style.display = 'block';
            walletInfo.style.display = 'none';
        }
    }

    getNetworkName(chainId) {
        const networks = {
            1: 'Ethereum',
            56: 'BSC',
            137: 'Polygon',
            5: 'Goerli',
            31337: 'Reddio' // Reddio链ID
        };
        return networks[chainId] || `Chain ${chainId}`;
    }

    getContract(address, abi) {
        if (!this.contracts[address]) {
            this.contracts[address] = new this.web3.eth.Contract(abi, address);
        }
        return this.contracts[address];
    }

    async sendTransaction(contract, method, params = [], value = 0) {
        try {
            const gasPrice = await this.web3.eth.getGasPrice();
            const gasEstimate = await contract.methods[method](...params).estimateGas({
                from: this.account,
                value: value
            });

            const tx = await contract.methods[method](...params).send({
                from: this.account,
                gas: Math.floor(gasEstimate * 1.2),
                gasPrice: gasPrice,
                value: value
            });

            return tx.transactionHash;
        } catch (error) {
            console.error('交易失败:', error);
            throw error;
        }
    }

    isConnected() {
        return this.account !== null && this.web3 !== null;
    }

    formatBalance(balance, decimals = 18) {
        return (balance / Math.pow(10, decimals)).toFixed(6);
    }

    toWei(amount, decimals = 18) {
        return this.web3.utils.toBN(amount).mul(this.web3.utils.toBN(10).pow(this.web3.utils.toBN(decimals)));
    }
}

// ERC20 ABI (简化版)
const ERC20_ABI = [
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "_spender", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "approve",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
];

// Vault合约ABI (更新合约名称)
const VAULT_ABI = [
    {
        "inputs": [
            {"name": "token", "type": "address"},
            {"name": "amount", "type": "uint256"},
            {"name": "scosAmount", "type": "uint256"}
        ],
        "name": "stakeStock",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [{"name": "token", "type": "address"}],
        "name": "unstakeStock",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "user", "type": "address"},
            {"name": "token", "type": "address"}
        ],
        "name": "getStakeInfo",
        "outputs": [
            {
                "components": [
                    {"name": "amount", "type": "uint256"},
                    {"name": "borrowedSCOS", "type": "uint256"},
                    {"name": "timestamp", "type": "uint256"},
                    {"name": "active", "type": "bool"}
                ],
                "name": "",
                "type": "tuple"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    }
];