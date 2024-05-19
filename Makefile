GO := go

.SUFFIXES: .peg .go

ifdef DEBUG
	BUILD_FLAG += -gcflags="all=-N -l"
endif


grammar.go: grammar.peg
	peg -switch -inline -strict -output ./$@ $<

parser: grammar.go ./cmd/main.go
	${GO} build ${BUILD_FLAG} -o $@  ./cmd/*.go

all: grammar.go

clean:
	rm grammar.go 
	rm parser

.PHONY: cmd
