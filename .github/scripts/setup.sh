#!/bin/bash

cd examples/ethereum
nohup make run-rosetta > /dev/null 2>&1 &

nohup make run-rosetta-offline > /dev/null 2>&1 &

sleep 60

curl -s --location --request POST 'http://localhost:8080/network/list' \
--header 'Content-Type: application/json' \
--data-raw '{
    "metadata" : {}
}'

block_tip=($(curl -s --location --request POST 'http://localhost:8080/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Ethereum",
        "network": "Mainnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

echo "latest block index is", $block_tip

offline_network=($(curl -s --location --request POST 'http://localhost:8081/network/list' \
--header 'Content-Type: application/json' \
--data-raw '{
    "metadata" : {}
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["network_identifiers"][0]["network"])'))

echo "offline network is", $offline_network
