up:
	rm -Rdf ./geth/docker-data
	docker-compose -f ./geth/docker/docker-compose.yml up --build --remove-orphans -V -d
	go run ./test
	docker-compose -f ./geth/docker/docker-compose.yml down
	chmod -R 777 blockchain_data

down:
	cd ./geth/docker && docker-compose down

stop:
	docker stop $(docker ps -a -q)