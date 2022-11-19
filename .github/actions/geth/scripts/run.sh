#!/bin/sh

set -eu

geth \
    --verbosity 5 \
    --nodiscover \
    --syncmode 'full' \
    --nat none \
    --port 30310 \
    --http \
    --http.addr '0.0.0.0' \
    --http.port 8545 \
    --http.vhosts '*' \
    --http.api 'personal,eth,net,web3,txpool,miner,debug' \
    --networkid '1' \
    --mine \
    --miner.etherbase '0x4cdBd835fE18BD93ccA39A262Cff72dbAC99E24F' \
    --miner.gasprice 0 \
    --unlock "0,1,2" \
    --password /root/config/password.txt \
    --allow-insecure-unlock
