// github.com/comiccoin-network/monorepo/cloud/comiccoin/common/storage/database/mongodbcache
package mongodbcache

type CacheConfigurationProvider interface {
	GetDatabaseName() string
}

type cacheConfigurationProviderImpl struct {
	databaseName string
}

func NewCacheConfigurationProvider(databaseName string) CacheConfigurationProvider {
	return &cacheConfigurationProviderImpl{
		databaseName: databaseName,
	}
}

func (impl *cacheConfigurationProviderImpl) GetDatabaseName() string {
	return impl.databaseName
}
