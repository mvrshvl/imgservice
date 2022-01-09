generate:
	rm -Rdf ./test/geth/docker-data
	docker-compose -f ./test/geth/docker/docker-compose.yml up --build --remove-orphans -V -d
	go run ./test
	chmod -R 777 blockchain_data

stop-generate:
	docker-compose -f ./test/geth/docker/docker-compose.yml down

down:
	cd ./geth/docker && docker-compose down

stop:
	docker stop $(docker ps -a -q)

gen-binding:
	docker build -t bindings -f ./test/contract/generator/Dockerfile ./test/contract && docker run --privileged -v $(PWD)/test/contract:/contract bindings

network:
	docker network create --driver=bridge --subnet=172.25.0.0/24 chainnet || true