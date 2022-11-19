var USDC = artifacts.require("./USDC.sol");
module.exports = function(deployer) {
   deployer.deploy(USDC);
};