TOKEN='YOUR-TOKEN-HERE'

start:
	@ACCESS_TOKEN=$(TOKEN) go run ./cmd -d ./default.yaml

tidy:
	@go mod tidy

build:
	@go build .

gen:
	@go run ./cmd -g -d .

.PHONY: start tidy build gen
