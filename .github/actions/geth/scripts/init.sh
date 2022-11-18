#!/bin/sh

set -eu

geth init /root/config/genesis.json
geth --password /root/config/password.txt account import /root/config/private-key1.txt
geth --password /root/config/password.txt account import /root/config/private-key2.txt
geth --password /root/config/password.txt account import /root/config/private-key3.txt 

