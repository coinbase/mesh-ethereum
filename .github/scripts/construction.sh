#!/bin/bash

# downloading cli
curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-cli/master/scripts/install.sh | sh -s v0.10.0

echo "start check:construction"
./bin/rosetta-cli --configuration-file rosetta-cli-conf/devnet/config.json check:construction