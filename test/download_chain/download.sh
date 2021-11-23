#!/bin/bash

docker exec docker_blockchain-sealer1_1 ethereumetl export_blocks_and_transactions --start-block 0 --end-block "$1" --blocks-output /root/output/blocks.csv --transactions-output /root/output/transactions.csv \--provider-uri file:///root/.ethereum/geth.ipc

docker exec docker_blockchain-sealer1_1 ethereumetl export_token_transfers --start-block 0 --end-block "$1" --output /root/output/token_transfers.csv --provider-uri file:///root/.ethereum/geth.ipc

docker exec docker_blockchain-sealer1_1 ethereumetl extract_csv_column --input /root/output/transactions.csv --column hash --output transaction_hashes.txt

docker exec docker_blockchain-sealer1_1 ethereumetl export_receipts_and_logs --transaction-hashes transaction_hashes.txt --logs-output /root/output/logs.csv --provider-uri file:///root/.ethereum/geth.ipc