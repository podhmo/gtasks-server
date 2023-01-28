SHELL := bash
GO := go

lint:
	go vet ./...

run:
	$(GO) run .

doc:
	go generate