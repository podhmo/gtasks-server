SHELL := bash
GO := go

lint:
	go vet ./...

run:
	$(GO) run .

doc:
	$(GO) run . --gendoc > openapi.json