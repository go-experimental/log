GOCMD=GO111MODULE=on go

lint:
	go vet ./...

test:
	$(GOCMD) test -cover -race ./...

bench:
	$(GOCMD) test -bench=. -benchmem ./...

.PHONY: test lint