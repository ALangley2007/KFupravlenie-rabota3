GO ?= go
INPUT ?=
OUTPUT ?=
TEST ?= false
START ?=
END ?=

.PHONY: help run interpret build test check fmt clean

help:
	@./run.sh help

run:
	GO="$(GO)" ./run.sh run -input "$(INPUT)" -output "$(OUTPUT)" $(if $(filter true 1,$(TEST)),-test,)

interpret:
	GO="$(GO)" ./run.sh interpret -input "$(INPUT)" -output "$(OUTPUT)" -start "$(START)" -end "$(END)"

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
