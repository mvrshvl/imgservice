up:
	rm -rf ./test/geth.ipc
	touch ./test/geth.ipc
	rm -Rdf ./geth/docker-data
	cd ./geth/docker && docker-compose up --build --remove-orphans -V

down:
	cd ./geth/docker && docker-compose down

stop:
	docker stop $(docker ps -a -q)