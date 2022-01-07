#!/usr/bin/env ash

#GAS_PRICE_PARAMS=--miner.gastarget=4294967040 --miner.gaslimit=4294967040
GAS_PRICE_PARAMS="--miner.gastarget=${MinerGasTarget} --miner.gaslimit=${MinerGasLimit} --miner.gasprice=${GasPrice}"
#TX_POOL="--txpool.pricelimit=0 --txpool.globalslots=140960 --txpool.globalqueue=160240"
TX_POOL="--txpool.nolocals --txpool.pricelimit=${TxpoolPriceLimit} --txpool.accountslots=${TxpoolAccountSlots} --txpool.globalslots=${TxpoolGlobalSlots} --txpool.accountqueue=${TxpoolAccountQueue} --txpool.globalqueue=${TxpoolGlobalQueue}"
CACHE_PARAMS="--cache=${Cache} --cache.database=${CacheDatabase} --cache.trie=${CacheTrie} --cache.gc=${CacheGC}"

if [ ! -d ${DATA_DIR}/geth ]
then
    geth --datadir ${DATA_DIR} ${GAS_PRICE_PARAMS} init /genesis.json
    mkdir -p ${DATA_DIR}/keystore && cp /keystore/* ${DATA_DIR}/keystore
fi
NETWORK_ID="6524631"
BOOTNODE_HASH="7d9bb2af83b9c60b548af6ce778faf1778c3bdc528e045f17f6e07cd5aadbfa6b614437264b7d9241191d0f581b8e54fb27c0efa3fff2fb5d8e42efa6dcd1c43"
BOOTNODE_IP="172.25.0.2"
HOST_IP="$(hostname -i)"
HOST_NAME="${NODE_TYPE}-${HOST_IP}"
VPC_CIDR_IP="172.25.0.0/24"

echo "GAS_PRICE ${GasPrice}"
echo "HOST_IP: ${HOST_IP}"
echo "HOST_NAME: ${HOST_NAME}"
echo "VPC_CIDR_IP: ${VPC_CIDR_IP}"
echo "NODE_TYPE: ${NODE_TYPE}"
echo "ETHER_BASE: ${ETHER_BASE}"

args="--datadir ${DATA_DIR} --password=keystore/password \
--networkid ${NETWORK_ID} ${CACHE_PARAMS} --port ${LISTEN_PORT} \
--maxpeers 50 --allow-insecure-unlock --nat extip:${HOST_IP} --nousb --nodiscover ${TX_POOL} ${GAS_PRICE_PARAMS}"
#        --ethstats ${HOST_NAME}:${ETH_STATS_SECRET}@${ETH_STATS_URL} --syncmode full
#        --maxpeers 50 --netrestrict ${VPC_CIDR_IP} --nousb --verbosity 4"

if [ ${NODE_TYPE} = "BOOTNODE" ]; then
    args="$args --nodekey /boot.key --lightpeers=10 --lightserv=50 --verbosity 1"
    exec /usr/local/bin/geth $args && chmod 0777 -R ${DATA_DIR}
elif [ ${NODE_TYPE} = "SEALER" ]; then
    echo $ENODE_KEY >> /node.key
    geth account import --password /keystore/password /node.key
    args="$args --nodekey /node.key --etherbase ${ETHER_BASE} --unlock ${ETHER_BASE}  --bootnodes enode://${BOOTNODE_HASH}@${BOOTNODE_IP}:${LISTEN_PORT}  \
--rpc --rpcaddr 0.0.0.0 --rpcport ${NODE_RPC} --rpcapi admin,web3,eth,net,personal,debug,txpool --rpccorsdomain=${CORS_DOMAIN} --rpcvhosts=${RPC_VHOSTS} \
--ws --wsaddr 0.0.0.0 --wsport ${NODE_WS} --wsapi admin,eth,net,personal,web3,debug,txpool --wsorigins=${ETH_WS_CORS_DOMAIN} \
--mine --miner.threads=1 --rpc.gascap=0 --verbosity 4 --syncmode full"
    exec /usr/local/bin/geth $args && chmod 0777 -R ${DATA_DIR}
elif [ ${NODE_TYPE} = "WORKER" ]; then
    args="$args --bootnodes enode://${BOOTNODE_HASH}@${BOOTNODE_IP}:${LISTEN_PORT} --unlock ${ETHER_BASE} \
--rpc --rpcaddr 0.0.0.0 --rpcport ${NODE_RPC} --rpcapi admin,web3,eth,net,personal,debug --rpccorsdomain=${CORS_DOMAIN} --rpcvhosts=${RPC_VHOSTS} \
--ws --wsaddr 0.0.0.0 --wsport ${NODE_WS} --rpc.gascap=0 --wsapi admin,eth,net,personal,web3,debug --wsorigins=${ETH_WS_CORS_DOMAIN} --verbosity 4 --syncmode full"
    exec /usr/local/bin/geth $args && chmod 0777 -R ${DATA_DIR}

else
    echo "[ERROR] Invalid NODE_TYPE!!!"
    exit 1
fi
