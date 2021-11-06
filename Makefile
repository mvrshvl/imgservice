up:
	rm -Rdf ./geth/docker-data
	cd ./geth/docker && docker-compose up --build --remove-orphans -V

down:
	cd ./clients/geth/docker && docker-compose down ;;

stop:
	docker stop $(docker ps -a -q)