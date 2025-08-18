const hre = require("hardhat");

async function main() {
    // 部署 AppleStock
    const AppleStock = await hre.ethers.getContractFactory("AppleStock");
    const appleStock = await AppleStock.deploy();
    await appleStock.waitForDeployment();
    console.log("AppleStock 部署地址:", await appleStock.getAddress());

    // 部署 SCOS
    const SCOS = await hre.ethers.getContractFactory("SCOS");
    const scos = await SCOS.deploy();
    await scos.waitForDeployment();
    console.log("SCOS 部署地址:", await scos.getAddress());

    // 部署 StockVault
    const StockVault = await hre.ethers.getContractFactory("StockVault");
    const stockVault = await StockVault.deploy();
    await stockVault.waitForDeployment();
    console.log("StockVault 部署地址:", await stockVault.getAddress());
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
