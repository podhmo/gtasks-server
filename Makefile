SHELL := bash
GO := go

lint:
	go vet ./...

run:
	$(GO) run main.go

doc:
	$(GO) run main.go --gendoc > openapi.json