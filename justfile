fmt:
	gofumpt -w .

test:
	go test -v ./...

lint:
	revive -config revive.toml -formatter friendly ./...
