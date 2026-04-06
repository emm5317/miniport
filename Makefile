.PHONY: build test dev deploy docker docker-run clean

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o miniport ./cmd/miniport

test:
	go test ./...

dev:
	MINIPORT_HOST=127.0.0.1 MINIPORT_PORT=8092 go run ./cmd/miniport

deploy: build
	scp miniport voicetask:/opt/miniport/miniport
	ssh voicetask "sudo systemctl restart miniport"

docker:
	docker build -t miniport .

docker-run: docker
	docker run --rm -p 8092:8092 -v /var/run/docker.sock:/var/run/docker.sock miniport

clean:
	rm -f miniport
