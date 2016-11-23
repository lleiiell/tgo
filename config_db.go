package tgo

import (
	"math/rand"
	"sync"
	"time"
)

var (
	dbConfigMux sync.Mutex
	dbConfig    *ConfigDb
)

type ConfigDb struct {
	Mysql ConfigMysql
	Mongo ConfigMongo
}

func NewConfigDb() *ConfigDb {
	return &ConfigDb{}
}

type ConfigDbBase struct {
	Address  string
	Port     int
	User     string
	Password string
	DbName   string `json:"-"`
}

type ConfigDbPool struct {
	PoolMinCap      int
	PoolExCap       int
	PoolMaxCap      int
	PoolIdleTimeout time.Duration
	PoolWaitCount   int64
	PoolWaitTimeout time.Duration
}

type ConfigMysql struct {
	DbName string
	Pool   ConfigDbPool
	Write  ConfigDbBase
	Reads  []ConfigDbBase
}

type ConfigMongo struct {
	DbName      string
	Servers     string
	Read_option string
	Timeout     int
	PoolLimit   int
}

func configDbInit() {

	if dbConfig == nil || dbConfig.Mysql.DbName == "" {

		dbConfigMux.Lock()

		defer dbConfigMux.Unlock()

		dbConfig = &ConfigDb{}

		defaultDbConfig := configDbGetDefault()

		configGet("db", dbConfig, defaultDbConfig)

	}
}

func configDbClear() {
	dbConfigMux.Lock()

	defer dbConfigMux.Unlock()

	dbConfig = nil
}
func configDbGetDefault() *ConfigDb {
	return &ConfigDb{Mysql: ConfigMysql{
		DbName: "",
		Pool:   ConfigDbPool{5, 5, 20, 3600, 100, 60},
		Write:  ConfigDbBase{"ip", 33062, "user", "password", ""},
		Reads: []ConfigDbBase{ConfigDbBase{"ip", 3306, "user", "password", ""},
			ConfigDbBase{"ip", 33062, "user", "password", ""}}},
		Mongo: ConfigMongo{DbName: "dbname", Servers: "servers", Read_option: "PRIMARY", Timeout: 1000, PoolLimit: 30}}
}

func NewConfigMysql() *ConfigMysql {
	return &ConfigMysql{}
}

func (m *ConfigMysql) GetPool() *ConfigDbPool {
	configDbInit()
	poolConfig := dbConfig.Mysql.Pool
	if &poolConfig == nil {
		poolConfig = configDbGetDefault().Mysql.Pool
	}
	return &poolConfig
}

func (m *ConfigMysql) GetWrite() *ConfigDbBase {

	configDbInit()

	writeConfig := dbConfig.Mysql.Write

	if &writeConfig == nil {

		writeConfig = configDbGetDefault().Mysql.Write
	}
	writeConfig.DbName = dbConfig.Mysql.DbName

	return &writeConfig
}

func (m *ConfigMysql) GetRead() (config *ConfigDbBase) {
	configDbInit()

	readConfigs := dbConfig.Mysql.Reads

	if &readConfigs == nil || len(readConfigs) == 0 {
		return &configDbGetDefault().Mysql.Reads[0]
	}
	count := len(readConfigs)

	if count > 1 {
		rand.Seed(time.Now().UnixNano())

		config = &readConfigs[rand.Intn(count-1)]
	}
	config = &readConfigs[0]
	config.DbName = dbConfig.Mysql.DbName

	return config
}

func (m *ConfigMongo) Get() (config *ConfigMongo) {
	configDbInit()

	mongoConfig := dbConfig.Mongo

	if mongoConfig.DbName == "" {
		mongoConfig = configDbGetDefault().Mongo
	}
	return &mongoConfig
}
