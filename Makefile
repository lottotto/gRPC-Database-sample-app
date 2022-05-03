.PHONY: client server build

build: client server
	echo "build client & server docker image"

client: 
	docker build -f ./build/client/Dockerfile -t client:0.0.1  .

server: 
	docker build -f ./build/server/Dockerfile -t server:0.0.1  .
