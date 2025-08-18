require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config();

const PRIVATE_KEY = "0x36a15f8d63742eaabf9ebb32a8551db13d6a3167";

module.exports = {
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
