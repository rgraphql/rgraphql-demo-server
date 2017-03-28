IMAGE=paralin/rgraphql-demo-server:latest

build:
	go build -v .

build-static:
	CGO_ENABLED=0 go build -v .
	strip -s ./rgraphql-demo-server

docker:
	docker build -t "$(IMAGE)" .

push:
	docker push $(IMAGE)
