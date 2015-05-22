
SRCS:=$(wildcard *.go)

lb: $(SRCS)
	go build

clean:
	rm -rf lb
