package tgo

import (
	"math/rand"
	"sync"
	"time"
)

var (
	dbConfigMux sync.Mutex
	dbConfig    *DbConfig
)

type DbConfig struct {
	Mysql MysqlConfig
	Mongo MongoConfig
}

func NewDb() *DbConfig {
	return &DbConfig{}
}

type DbBaseConfig struct {
	Address  string
	Port     int
	User     string
	Password string
	DbName   string `json:"-"`
}

type MysqlConfig struct {
	DbName string
	Write  DbBaseConfig
	Reads  []DbBaseConfig
}

type MongoConfig struct {
	DbName      string
	Servers     string
	Read_option string
	Timeout     int
}

func getDbConfigs() {

	if dbConfig == nil || dbConfig.Mysql.DbName == "" {

		dbConfigMux.Lock()

		defer dbConfigMux.Unlock()

		dbConfig = &DbConfig{}

		defaultDbConfig := getDefaultDbConfig()

		getConfigs("db", dbConfig, defaultDbConfig)

	}
}

func getDefaultDbConfig() *DbConfig {
	return &DbConfig{Mysql: MysqlConfig{Write: DbBaseConfig{"172.172.177.15", 33062, "root", "root@dev", ""},
		Reads: []DbBaseConfig{DbBaseConfig{"172.172.177.15", 33062, "root", "root@dev", ""},
			DbBaseConfig{"172.172.177.15", 33062, "root", "root@dev", ""}}},
		Mongo: MongoConfig{DbName: "Component", Servers: "172.172.177.20:36004", Read_option: "PRIMARY", Timeout: 1000}}
}

func NewMysqlConfig() *MysqlConfig {
	return &MysqlConfig{}
}

func (m *MysqlConfig) GetMysqlWriteConfig() *DbBaseConfig {

	getDbConfigs()

	writeConfig := dbConfig.Mysql.Write

	if &writeConfig == nil {

		writeConfig = getDefaultDbConfig().Mysql.Write
	}
	writeConfig.DbName = dbConfig.Mysql.DbName

	return &writeConfig
}

func (m *MysqlConfig) GetMysqlReadConfig() (config *DbBaseConfig) {
	getDbConfigs()

	readConfigs := dbConfig.Mysql.Reads

	if &readConfigs == nil || len(readConfigs) == 0 {
		return &getDefaultDbConfig().Mysql.Reads[0]
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

func (m *MongoConfig) GetMongoConfig() (config *MongoConfig) {
	getDbConfigs()

	mongoConfig := dbConfig.Mongo

	if mongoConfig.DbName == "" {
		mongoConfig = getDefaultDbConfig().Mongo
	}
	return &mongoConfig
}
