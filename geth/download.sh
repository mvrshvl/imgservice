#!/bin/bash

ethereumetl export_blocks_and_transactions -w 20 --start-block "$1" --end-block "$2" --blocks-output  ./geth/data/blocks.csv --transactions-output  ./geth/data/transactions.csv \--provider-uri $3

ethereumetl export_token_transfers -w 20 --start-block "$1" --end-block "$2" --output  ./geth/data/token_transfers.csv --provider-uri $3

ethereumetl extract_csv_column --input  ./geth/data/transactions.csv --column hash --output ./geth/data/transaction_hashes.txt

ethereumetl export_receipts_and_logs --transaction-hashes  ./geth/data/transaction_hashes.txt --logs-output ./geth/data/logs.csv --provider-uri $3