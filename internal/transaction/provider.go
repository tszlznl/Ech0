package transaction

import "github.com/google/wire"

func ProvideTransactionManager(factory *TransactionManagerFactory) TransactionManager {
	return factory.TransactionManager()
}

var FactorySet = wire.NewSet(NewTransactionManagerFactory)
var ManagerSet = wire.NewSet(ProvideTransactionManager)
var ProviderSet = wire.NewSet(FactorySet, ManagerSet)
