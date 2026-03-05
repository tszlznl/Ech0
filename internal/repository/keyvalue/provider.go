package keyvalue

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewKeyValueRepository)
