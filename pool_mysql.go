package tgo

import (
    "github.com/youtube/vitess/go/pools"
    "time"
    "github.com/jinzhu/gorm"
    "fmt"
    "golang.org/x/net/context"
)

const (
    MYSQL_CONNECTION_TYPE_READ = 1
    MYSQL_CONNECTION_TYPE_WRITE = 2
)

type MysqlConnectionPool struct {
    *pools.ResourcePool
}

type MysqlConnection struct {
    *gorm.DB
}

func (c MysqlConnection) Close() {
    c.DB.Close()
}

func (c MysqlConnection) Put() {
    MysqlPool.Put(c)
}

func NewMysqlConnectionPool(factory pools.Factory, capacity, maxCap int, idleTimeout time.Duration) (*MysqlConnectionPool) {
    return &MysqlConnectionPool{
        pools.NewResourcePool(factory, capacity, maxCap, idleTimeout),
    }
}

func CreateMysqlConnectionRead() (pools.Resource, error) {
    resultDb, err := initDb(MYSQL_CONNECTION_TYPE_READ)
    return MysqlConnection{resultDb}, err
}

func CreateMysqlConnectionWrite() (pools.Resource, error) {
    resultDb, err := initDb(MYSQL_CONNECTION_TYPE_WRITE)
    return MysqlConnection{resultDb}, err
}

func initDb(connectionType int) (*gorm.DB, error) {
    config := NewConfigDb()
    var dbConfig *ConfigDbBase
    if connectionType == MYSQL_CONNECTION_TYPE_READ {
        dbConfig = config.Mysql.GetRead()
    } else {
        dbConfig = config.Mysql.GetWrite()
    }
    address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Address, dbConfig.Port, dbConfig.DbName)
    resultDb, err := gorm.Open("mysql", address)
    if err != nil {
        UtilLogErrorf("connect mysql error: %s", err.Error())
        return nil, err
    }

    resultDb.SingularTable(true)

    if ConfigEnvGet() != "idc" {
        resultDb.LogMode(true)
    }

    return resultDb, err
}

func (p *MysqlConnectionPool) Get(isRead bool) (MysqlConnection, error) {
    ctx := context.TODO()
    r, err := p.ResourcePool.Get(ctx)
    if err != nil {
        UtilLogErrorf("connect mysql pool get error: %s", err.Error())
    }
    c := r.(MysqlConnection)
    //判断conn是否正常
    if c.Error != nil {
        c.Close()
        var db *gorm.DB
        if isRead {
            db, err = initDb(MYSQL_CONNECTION_TYPE_READ)
        } else {
            db, err = initDb(MYSQL_CONNECTION_TYPE_WRITE)
        }
        c = MysqlConnection{db}
        if err != nil {
            UtilLogErrorf("redo connect mysql error: %s", err.Error())
            p.Put(c)//放入失败的资源，保证下次重连
        }
    }
    return c, err
}