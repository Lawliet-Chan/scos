async function deployTokenContract() {
    const gasPrice = (await ethers.provider.getFeeData()).gasPrice;
  
    const deployGasParams = {
      gasLimit: 4200000,
      maxFeePerGas: gasPrice,
    };
  
    let tokenContract;
    const TokenContractAddress = await ethers.getContractFactory('SCOS');
    
    tokenContract = await TokenContractAddress.deploy(
      ethers.parseEther("800000000"),
      deployGasParams);
    console.log('tokenContract:', tokenContract.deploymentTransaction().hash);
    await tokenContract.waitForDeployment();
    console.log(`tokenContract deployed: ${tokenContract.target}`);
    return tokenContract;
  }
  
  // We recommend this pattern to be able to use async/await everywhere
  // and properly handle errors.
  if (require.main === module) {
    deployTokenContract()
      .then(() => process.exit(0))
      .catch((error) => {
        console.error(error);
        process.exit(1);
      });
  }
  
  exports.deployTokenContract = deployTokenContract;
  