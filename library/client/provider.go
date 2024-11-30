package fclient

import "github.com/google/wire"

var Provider = wire.NewSet(NewLicenseManagerClient)
