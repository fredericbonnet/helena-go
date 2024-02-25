package picol_dialect_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPicolDialect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Picol Dialect Suite")
}
