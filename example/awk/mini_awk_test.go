package awk

import (
	"strings"
	"testing"
)

func TestMiniAWK(t *testing.T) {
	scripts := `
(awk 
	(BEGIN (printf "Language\tFirstAppearAt\n" )) 
	(PATTERN (match ".*C.*" $1) (printf "%-8s\t%s\n" $1 $3)) 
	(END (printf "---------------------------------------------\n")) 
)
	`
	input := `
C Static 1972
C++ Static 1985
Python Dynamic 1991
Ruby Dynamic 1995
	`

	err := EvalAWK(scripts, strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
}
