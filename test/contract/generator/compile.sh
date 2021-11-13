#!/bin/bash

CONTRACTS=$(ls "$1"/*.sol)

for contract in ${CONTRACTS[@]}
do
  abigen --sol "$contract" --pkg contract --out "$contract".go
  chmod 666 "$contract".go
done