
SRCS:=$(wildcard *.go)

lb: $(SRCS)
	go build

get-deps:
	go get github.com/codegangsta/cli
	go get github.com/satori/go.uuid
	go get github.com/hamano/golang-openldap

clean:
	rm -rf lb
