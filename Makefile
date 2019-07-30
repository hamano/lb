
SRCS:=$(wildcard *.go)

lb: $(SRCS)
	go build

deps:
	go get -u github.com/urfave/cli
	go get -u github.com/satori/go.uuid
	go get -u gopkg.in/ldap.v3

clean:
	rm -rf lb

install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -m 755 lb $(DESTDIR)/usr/bin/
