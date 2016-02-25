
SRCS:=$(wildcard *.go)

lb: $(SRCS)
	go build

get-deps:
	go get github.com/codegangsta/cli
	go get github.com/satori/go.uuid
	go get github.com/hamano/golang-openldap

clean:
	rm -rf lb

install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -m 755 lb $(DESTDIR)/usr/bin/
