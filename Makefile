
SRCS:=$(wildcard *.go)

get-deps:
	go get github.com/codegangsta/cli
	go get github.com/satori/go.uuid
	go get github.com/hamano/golang-openldap

lb: $(SRCS)
	go build

clean:
	rm -rf lb
