BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.PHONY: test

build:
	cd gateway && go build

run: build
	PORT=8080 TOKEN=dev DEBUG=1 NOTLS=1 gateway/gateway

test:
	go test -v ./...

savedeps:
	godep save -r ./...

deploy:
	git push -f heroku $(BRANCH):master
