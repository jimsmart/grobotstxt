package grobotstxt_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRobotstxt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "grobotstxt Suite")
}
