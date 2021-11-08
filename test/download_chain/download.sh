#!/bin/bash

docker exec docker_blockchain-sealer1_1 ethereumetl export_blocks_and_transactions --start-block 0 --end-block "$1" --blocks-output /root/output/blocks.csv --transactions-output /root/output/transactions.csv \--provider-uri file:///root/.ethereum/geth.ipc