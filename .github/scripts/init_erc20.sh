#!/bin/bash

echo "update apt-get"
apt-get update

echo "install nodejs"
apt-get install -y nodejs

echo "check npm version"
npm -v

echo "install truffle"
npm install -g truffle

echo "install solc"
npm install -g solc

cd .github/actions/geth/truffle
echo "deploy the contract"
truffle migrate 
