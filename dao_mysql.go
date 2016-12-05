package tgo

import (
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
	MysqlReadPool     *MysqlConnectionPool
	mysqlReadPoolMux  sync.Mutex
	MysqlWritePool    *MysqlConnectionPool
	mysqlWritePoolMux sync.Mutex
)

func init() {
	config := NewConfigDb()
	configPool := config.Mysql.GetPool()
	poolTicker := time.NewTicker(time.Second * 60)

	//todo 优化动态控制池子大小

	go monitorPool(configPool, poolTicker, true, MysqlReadPool)
	go monitorPool(configPool, poolTicker, false, MysqlWritePool)
}

func monitorPool(configPool *ConfigDbPool, poolTicker *time.Ticker, isRead bool, mysqlPool *MysqlConnectionPool) {
	var (
		caps         int
		poolCaps     int
		oldWaitCount int64
		waitCount    int64
	)
	for {
		if mysqlPool == nil || mysqlPool.IsClosed() {
			mysqlPool = initMysqlPool(isRead)
		}
		waitCount = mysqlPool.WaitCount() - oldWaitCount
		oldWaitCount = mysqlPool.WaitCount()
		poolCaps = int(mysqlPool.Capacity())
		if waitCount >= configPool.PoolWaitCount && poolCaps != configPool.PoolMaxCap { //定时循环内超出多少等待数目
			caps = poolCaps + configPool.PoolExCap
		} else if waitCount == 0 && poolCaps != configPool.PoolMinCap { //闲时减少池子容量
			caps = poolCaps - configPool.PoolExCap
		} else {
			<-poolTicker.C
			continue
		}
		if caps < configPool.PoolMinCap {
			caps = configPool.PoolMinCap
		}
		if caps > configPool.PoolMaxCap {
			caps = configPool.PoolMaxCap
		}
		mysqlPool.SetCapacity(caps)
		<-poolTicker.C
	}
}

func initMysqlPool(isRead bool) *MysqlConnectionPool {
	config := NewConfigDb()
	configPool := config.Mysql.GetPool()
	if isRead {
		if MysqlReadPool == nil || MysqlReadPool.IsClosed() {
			mysqlReadPoolMux.Lock()
			defer mysqlReadPoolMux.Unlock()
			MysqlReadPool = NewMysqlConnectionPool(CreateMysqlConnectionRead, configPool.PoolMinCap,
				configPool.PoolMaxCap, configPool.PoolIdleTimeout*time.Millisecond)
		}
		return MysqlReadPool
	} else {
		if MysqlWritePool == nil || MysqlWritePool.IsClosed() {
			mysqlWritePoolMux.Lock()
			defer mysqlWritePoolMux.Unlock()
			MysqlWritePool = NewMysqlConnectionPool(CreateMysqlConnectionWrite, configPool.PoolMinCap,
				configPool.PoolMaxCap, configPool.PoolIdleTimeout*time.Millisecond)
		}
		return MysqlWritePool
	}
}

func initMysqlPoolConnection(isRead bool) (MysqlConnection, error) {
	return initMysqlPool(isRead).Get(isRead)
}

func (d *DaoMysql) GetReadOrm() (MysqlConnection, error) {
	return d.getOrm(true)
}

func (d *DaoMysql) GetWriteOrm() (MysqlConnection, error) {
	return d.getOrm(false)
}

func (d *DaoMysql) getOrm(isRead bool) (MysqlConnection, error) {
	return initMysqlPoolConnection(isRead)
}

func (d *DaoMysql) Insert(model interface{}) error {
	orm, err := d.GetWriteOrm()
	if err != nil {
		return err
	}
	defer orm.Put()
	errInsert := orm.Table(d.TableName).Create(model).Error
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
		errFind = orm.Table(d.TableName).Where(condition).Find(data).Error
	} else {
		errFind = orm.Table(d.TableName).Where(condition).Select(field[0]).Find(data).Error
	}

	return errFind
}
