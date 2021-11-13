FROM ethereum/solc:0.8.0-alpine

ADD . /contract
WORKDIR /contract

RUN apk add bash

ENV PATH /usr/local/bin:$PATH

COPY --from=ethereum/client-go:alltools-v1.10.4 /usr/local/bin/ /usr/local/bin/

ENTRYPOINT ["bash", "/contract/generator/compile.sh", "/contract"]