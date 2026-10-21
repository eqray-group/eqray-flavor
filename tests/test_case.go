package tests

import (
	"github.com/goravel/framework/testing"

	"eqray-flavor/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
