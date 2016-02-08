BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.PHONY: test

run:
	PORT=8080 TOKEN=dev DEBUG=1 NOTLS=1 go run hub/hub.go

test:
	go test -v ./...

savedeps:
	godep save -r ./...

deploy:
	git push -f heroku $(BRANCH):master
