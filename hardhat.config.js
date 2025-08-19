import "@nomicfoundation/hardhat-toolbox";

const PRIVATE_KEY = "0x01c7939dc6827ee10bb7d26f420618c4af88c0029aa70be202f1ca7f29fe5bb4";

const config = {
    solidity: "0.8.20",
    networks: {
        reddio: {
            url: "https://reddio-dev.reddio.com/",
            accounts: [PRIVATE_KEY],
        },
        scrollSepolia: {
            url: "https://sepolia-rpc.scroll.io",
            accounts: [PRIVATE_KEY],
        },
    },
};

export default config;
