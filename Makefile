GO := go

.SUFFIXES: .peg .go

ifdef DEBUG
	BUILD_FLAG += -gcflags="all=-N -l"
endif

CMDS := echo

ALL: ${CMDS}

grammar.go: grammar.peg
	peg -switch -inline -strict -output ./$@ $<

${CMDS}: grammar.go 
	${GO} build ${BUILD_FLAG} -o ./bin/$@  ./cmd/$@/*.go

all: grammar.go

clean:
	rm grammar.go 
	rm ./bin/*

.PHONY: ${CMDS}
