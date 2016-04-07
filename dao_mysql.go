package tgo

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
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

func initMysql(isRead bool) (*gorm.DB, error) {

	config := NewConfigDb()

	var dbConfig *ConfigDbBase
	if isRead {
		dbConfig = config.Mysql.GetRead()
	} else {
		dbConfig = config.Mysql.GetWrite()
	}
	//user:password@tcp(172.172.177.15:3306)dbname?charset=utf8
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Address, dbConfig.Port, dbConfig.DbName)
	//connectionString := dbConfig.User + ":" + dbConfig.Password + "@tcp(" + dbConfig.Address + ":" + dbConfig.Port + ")" + dbConfig.DbName + "?charset=utf8"

	db, err := gorm.Open("mysql", connectionString)

	if err != nil {
		//记录
		//errors.New("connect mysql error:" + err.Error())
		UtilLogError(fmt.Sprintf("connect mysql error:%s", err.Error()))
	}
	db.SingularTable(true)

	if ConfigEnvIsDev() {
		db.LogMode(true)
	}

	return db, err
}

func (d *DaoMysql) GetReadOrm() (*gorm.DB, error) {
	return d.getOrm(true)
}

func (d *DaoMysql) GetWriteOrm() (*gorm.DB, error) {
	return d.getOrm(false)
}

func (d *DaoMysql) getOrm(isRead bool) (*gorm.DB, error) {
	db, err := initMysql(isRead)

	if err != nil {
		return db, err
	}
	if d.TableName != "" {
		return db.Table(d.TableName), nil
	}
	return db, err
}

func (d *DaoMysql) Insert(model interface{}) error {
	orm, err := d.GetWriteOrm()

	if err != nil {
		return err
	}

	defer orm.Close()
	//mapData := structs.Map(model)

	errInsert := orm.Create(model).Error

	if errInsert != nil {
		//记录
		//errors.New("connect mysql error:" + err.Error())
		UtilLogError(fmt.Sprintf("insert data error:%s", errInsert.Error()))
	}

	return errInsert
}

func (d *DaoMysql) Select(condition string, data interface{}, field ...[]string) error {

	orm, err := d.GetReadOrm()

	if err != nil {
		return err
	}

	defer orm.Close()

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
