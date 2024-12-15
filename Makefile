.PHONY: all .FORCE

all: .FORCE
	go test ./...

gen: .FORCE
	go generate ./...

.FORCE:
