.DEFAULT_GOAL := run

run:
	go run .

run-build:
	make build
	./bin/crb

run-docker:
	docker run -v $$(pwd)/.env:/app/.env:ro -it --rm ghcr.io/nathan13888/coderunnerbot/crb:latest

build:
	go build -o bin/crb -ldflags "\
		-X 'main.BuildVersion=$$(git rev-parse --abbrev-ref HEAD)' \
		-X 'main.BuildTime=$$(date)' \
		-X 'main.GOOS=$$(go env GOOS)' \
		-X 'main.ARCH=$$(go env GOARCH)' \
		-s -w"

docker-build:
	docker build -t crb .

publish:
	make publish-ghcr

publish-ghcr:
	#make docker-build
	docker tag crb:latest ghcr.io/nathan13888/coderunnerbot/crb:latest
	docker push ghcr.io/nathan13888/coderunnerbot/crb:latest

pull-ghcr:
	docker pull ghcr.io/nathan13888/coderunnerbot/crb:latest

test:
	go test -v ./...

