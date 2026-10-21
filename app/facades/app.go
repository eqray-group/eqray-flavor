package facades

import (
	contractsFoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func App() contractsFoundation.Application {
	return foundation.App
}
