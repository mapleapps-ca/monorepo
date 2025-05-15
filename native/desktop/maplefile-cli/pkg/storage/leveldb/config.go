package leveldb

type LevelDBConfigurationProvider interface {
	GetDBPath() string
	GetDBName() string
}

type LevelDBConfigurationProviderImpl struct {
	dbPath string
	dbName string
}

func NewLevelDBConfigurationProvider(dbPath string, dbName string) LevelDBConfigurationProvider {
	return &LevelDBConfigurationProviderImpl{
		dbPath: dbPath,
		dbName: dbName,
	}
}

func (me *LevelDBConfigurationProviderImpl) GetDBPath() string {
	return me.dbPath
}

func (me *LevelDBConfigurationProviderImpl) GetDBName() string {
	return me.dbName
}
