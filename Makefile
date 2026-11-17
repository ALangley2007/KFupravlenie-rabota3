GO ?= go
INPUT ?=
OUTPUT ?=
TEST ?= false

.PHONY: help run build test check fmt clean

help:
	@./run.sh help

run:
	GO="$(GO)" ./run.sh run -input "$(INPUT)" -output "$(OUTPUT)" $(if $(filter true 1,$(TEST)),-test,)

build:
	GO="$(GO)" ./run.sh build

test:
	GO="$(GO)" ./run.sh test

check:
	GO="$(GO)" ./run.sh check

fmt:
	GO="$(GO)" ./run.sh fmt

clean:
	./run.sh clean
