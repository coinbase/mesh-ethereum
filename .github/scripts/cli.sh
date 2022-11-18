#!/bin/bash

# downloading cli
curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-cli/master/scripts/install.sh | sh -s

echo "start check:data"
./bin/rosetta-cli --configuration-file rosetta-cli-conf/mainnet/config.json check:data --start-block 0 --end-block 20
