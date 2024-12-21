.PHONY: all .FORCE

RUN=Test*
RUN=TestLex*

all: .FORCE
	go test -run $(RUN) ./...

gen: .FORCE
	go generate ./...

cover.out:
	go test -run $(RUN) -coverprofile cover.out ./...

cover.html: cover.out
	go tool cover -html=cover.out -o cover.html

.FORCE:
