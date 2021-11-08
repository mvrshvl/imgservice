FROM ethereum/client-go:v1.9.23

ENV DATA_DIR /root/.ethereum
ENV LISTEN_PORT 55503

COPY genesis.json /genesis.json
COPY geth.sh /geth.sh
COPY keystore /keystore
COPY boot.key /boot.key
COPY static-nodes.json /root/.ethereum/static-nodes.json

ENV PYTHONUNBUFFERED=1

RUN apk add build-base
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN apk add python3-dev
RUN python3 -m ensurepip
RUN python -m pip install Cython
RUN python -m pip install --no-cache --upgrade pip setuptools
RUN python -m pip install ethereum-etl

ENTRYPOINT ["/bin/sh", "/geth.sh"]