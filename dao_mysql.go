package tgo

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

type DaoMysql struct {
	TableName string
}

func NewDaoMysql() *DaoMysql {
	return &DaoMysql{}
}

type Condition struct {
	Field string
	Oper  string
	Value interface{}
}

type Sort struct {
	Field string
	Asc   bool
}

var (
	MysqlPool *MysqlConnectionPool
	mysqlReadPool *MysqlConnectionPool
	mysqlReadPoolMux sync.Mutex
	mysqlWritePool *MysqlConnectionPool
	mysqlWritePoolMux sync.Mutex
	poolTicker *time.Ticker
)

func initMysqlPool(isRead bool) (MysqlConnection, error) {
	config := NewConfigDb()
	configPool := config.Mysql.GetPool()
	if isRead {
		if mysqlReadPool == nil || mysqlReadPool.IsClosed() {
			mysqlReadPoolMux.Lock()
			defer mysqlReadPoolMux.Unlock()
			mysqlReadPool = NewMysqlConnectionPool(CreateMysqlConnectionRead, configPool.PoolCap,
				configPool.PoolMaxCap, configPool.PoolIdleTimeout*time.Millisecond)
		}
		MysqlPool = mysqlReadPool
	} else {
		if mysqlWritePool == nil || mysqlWritePool.IsClosed() {
			mysqlWritePoolMux.Lock()
			defer mysqlWritePoolMux.Unlock()
			mysqlWritePool = NewMysqlConnectionPool(CreateMysqlConnectionWrite, configPool.PoolCap,
				configPool.PoolMaxCap, configPool.PoolIdleTimeout*time.Millisecond)
		}
		MysqlPool = mysqlWritePool
	}
	if poolTicker == nil {
		poolTicker = time.NewTicker(time.Second * 60)
	}
	//todo 动态控制池子大小 - 优化
	if len(poolTicker.C) > 0 {
		<-poolTicker.C
		if MysqlPool.WaitCount() >= configPool.PoolWaitCount || MysqlPool.WaitTime() >= configPool.PoolWaitTimeout {
			caps := MysqlPool.Capacity() + configPool.PoolWaitCount
			if caps > int64(configPool.PoolMaxCap) {
				caps = int64(configPool.PoolMaxCap)
			}
			MysqlPool.SetCapacity(int(caps))
		} else {
			caps := MysqlPool.Capacity() - configPool.PoolWaitCount
			if caps < int64(configPool.PoolCap) {
				caps = int64(configPool.PoolCap)
			}
			MysqlPool.SetCapacity(int(caps))
		}
	}

	return MysqlPool.Get(isRead)
}

func (d *DaoMysql) GetReadOrm() (MysqlConnection, error) {
	return d.getOrm(true)
}

func (d *DaoMysql) GetWriteOrm() (MysqlConnection, error) {
	return d.getOrm(false)
}

func (d *DaoMysql) getOrm(isRead bool) (MysqlConnection, error) {
	return initMysqlPool(isRead)
}

func (d *DaoMysql) Insert(model interface{}) error {
	orm, err := d.GetWriteOrm()
	if err != nil {
		return err
	}
	defer orm.Put()
	errInsert := orm.Create(model).Error
	if errInsert != nil {
		//记录
		UtilLogError(fmt.Sprintf("insert data error:%s", errInsert.Error()))
	}

	return errInsert
}

func (d *DaoMysql) Select(condition string, data interface{}, field ...[]string) error {
	orm, err := d.GetReadOrm()
	if err != nil {
		return err
	}
	defer orm.Put()
	var errFind error
	if len(field) == 0 {
		errFind = orm.Where(condition).Find(data).Error
	} else {
		errFind = orm.Where(condition).Select(field[0]).Find(data).Error
	}
	if errFind != nil {
		UtilLogError(fmt.Sprintf("mysql select table %s error:%s", d.TableName, errFind.Error()))
		return errFind
	}

	return nil
}