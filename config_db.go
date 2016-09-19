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

type ConfigMysql struct {
	DbName string
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
	return &ConfigDb{Mysql: ConfigMysql{Write: ConfigDbBase{"172.172.177.15", 33062, "root", "root@dev", ""},
		Reads: []ConfigDbBase{ConfigDbBase{"172.172.177.15", 33062, "root", "root@dev", ""},
			ConfigDbBase{"172.172.177.15", 33062, "root", "root@dev", ""}}},
		Mongo: ConfigMongo{DbName: "Component", Servers: "172.172.177.20:36004", Read_option: "PRIMARY", Timeout: 1000,PoolLimit:30}}
}

func NewConfigMysql() *ConfigMysql {
	return &ConfigMysql{}
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
