.PHONY: all .FORCE

RUN=Test*
#RUN=TestLex*
#RUN=TestParse*

all: .FORCE
	go test -run $(RUN) ./...

gen: .FORCE
	go generate ./...

cover.out: .FORCE
	go test -run $(RUN) -coverprofile cover.out ./...

cover.html: cover.out
	go tool cover -html=cover.out -o cover.html

koneko: .FORCE
	go build cmd/koneko/koneko.go

.FORCE:
