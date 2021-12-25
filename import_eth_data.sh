#!/bin/bash

ethereumetl export_blocks_and_transactions -w 20 --start-block "$1" --end-block "$2" --blocks-output ./blockchain_data/mainnet/blocks.csv --transactions-output ./blockchain_data/mainnet/transactions.csv \--provider-uri file://$HOME/.ethereum/geth.ipc

ethereumetl export_token_transfers -w 20 --start-block "$1" --end-block "$2" --output ./blockchain_data/mainnet/token_transfers.csv --provider-uri file://$HOME/.ethereum/geth.ipc

ethereumetl extract_csv_column --input ./blockchain_data/mainnet/transactions.csv --column hash --output ./blockchain_data/mainnet/transaction_hashes.txt

ethereumetl export_receipts_and_logs --transaction-hashes ./blockchain_data/mainnet/transaction_hashes.txt --logs-output ./blockchain_data/mainnet/logs.csv --provider-uri file://$HOME/.ethereum/geth.ipc