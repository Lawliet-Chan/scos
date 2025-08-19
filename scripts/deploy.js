// scripts/deploy.js
import { ethers } from "hardhat";

async function main() {
    const AppleStock = await ethers.getContractFactory("AppleStock");
    const appleStock = await AppleStock.deploy();
    await appleStock.deployed();
    console.log(`AppleStock deployed to: ${appleStock.address}`);

    const SCOS = await ethers.getContractFactory("SCOS");
    const scos = await SCOS.deploy();
    await scos.deployed();
    console.log(`SCOS deployed to: ${scos.address}`);

    const StockVault = await ethers.getContractFactory("StockVault");
    const stockVault = await StockVault.deploy();
    await stockVault.deployed();
    console.log(`StockVault deployed to: ${stockVault.address}`);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
