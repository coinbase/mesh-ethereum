from web3 import Web3
from web3.middleware import geth_poa_middleware


web3 = Web3(Web3.HTTPProvider("http://127.0.0.1:8546"))
web3.middleware_onion.inject(geth_poa_middleware, layer=0)

print("latest block", web3.eth.block_number)

abi= [{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":"false","inputs":[{"indexed":"true","internalType":"address","name":"tokenOwner","type":"address"},{"indexed":"true","internalType":"address","name":"spender","type":"address"},{"indexed":"false","internalType":"uint256","name":"tokens","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":"false","inputs":[{"indexed":"true","internalType":"address","name":"from","type":"address"},{"indexed":"true","internalType":"address","name":"to","type":"address"},{"indexed":"false","internalType":"uint256","name":"tokens","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"delegate","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"delegate","type":"address"},{"internalType":"uint256","name":"numTokens","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenOwner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"receiver","type":"address"},{"internalType":"uint256","name":"numTokens","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"buyer","type":"address"},{"internalType":"uint256","name":"numTokens","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]
address = '0x62F3712A8A2bF3482F9Aa42F2C8296CF50774DDD'
contract= web3.eth.contract(address=address, abi=abi)

seedAddress="0x4cdBd835fE18BD93ccA39A262Cff72dbAC99E24F"
targetAddress="0x622Fbe99b3A378FAC736bf29d7e23B85E18816eB"

print("symbol is ", contract.functions.symbol().call())
print("decimal is ", contract.functions.decimals().call())
print("address is ", contract.address)
print("seedAddress balance is ", contract.functions.balanceOf(seedAddress).call())
print("targetAddress balance is ", contract.functions.balanceOf(targetAddress).call())
