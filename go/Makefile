CC=go
ACTION=build
ARCH:=$$(uname -s)
BIN=gotest
LC_ARCH:=$(shell echo $(ARCH) | tr A-Z a-z)

.PHONY: build

build:
	GOOS=$(LC_ARCH) $(CC) $(ACTION) -o $(BIN)
	docker $(ACTION) -t btoll/gotest:latest .

clean:
	@rm -f $(BIN)

