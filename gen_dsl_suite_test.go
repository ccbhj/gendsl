package gendsl

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenDsl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenDsl Suite")
}
