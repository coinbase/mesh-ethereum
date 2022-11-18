#!/bin/bash

nohup make run-rosetta > /dev/null 2>&1 &

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
