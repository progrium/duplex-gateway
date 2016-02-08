BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.PHONY: test

build: hub/hub
	cd hub && go build

run: build
	PORT=8080 TOKEN=dev DEBUG=1 NOTLS=1 hub/hub

test:
	go test -v ./...

savedeps:
	godep save -r ./...

deploy:
	git push -f heroku $(BRANCH):master
