.PHONY: all .FORCE

all: .FORCE
	go test ./...

.FORCE:
