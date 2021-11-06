FROM ethereum/client-go:v1.9.23

ENV DATA_DIR /root/.ethereum
ENV LISTEN_PORT 55503

COPY genesis.json /genesis.json
COPY geth.sh /geth.sh
COPY keystore /keystore
COPY boot.key /boot.key
COPY static-nodes.json /root/.ethereum/static-nodes.json

ENTRYPOINT ["/bin/sh", "/geth.sh"]