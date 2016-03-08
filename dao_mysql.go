package tgo

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Mysql struct {
	TableName string
}

func NewDaoMysql() *Mysql {

	return &Mysql{}
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

func initMysql(isRead bool) (gorm.DB, error) {

	config := configs.NewDb()

	var dbConfig *configs.DbBaseConfig
	if isRead {
		dbConfig = config.Mysql.GetMysqlReadConfig()
	} else {
		dbConfig = config.Mysql.GetMysqlWriteConfig()
	}
	//user:password@tcp(172.172.177.15:3306)dbname?charset=utf8
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Address, dbConfig.Port, dbConfig.DbName)
	//connectionString := dbConfig.User + ":" + dbConfig.Password + "@tcp(" + dbConfig.Address + ":" + dbConfig.Port + ")" + dbConfig.DbName + "?charset=utf8"

	db, err := gorm.Open("mysql", connectionString)

	if err != nil {
		//记录
		//errors.New("connect mysql error:" + err.Error())
		util.LogError(fmt.Sprintf("connect mysql error:%s", err.Error()))
	}
	db.SingularTable(true)

	if configs.IsEnvDev() {
		db.LogMode(true)
	}

	return db, err
}

func (d *Mysql) GetReadOrm() (gorm.DB, error) {
	return d.getOrm(true)
}

func (d *Mysql) GetWriteOrm() (gorm.DB, error) {
	return d.getOrm(false)
}

func (d *Mysql) getOrm(isRead bool) (gorm.DB, error) {
	db, err := initMysql(isRead)

	if err != nil {
		return db, err
	}
	if d.TableName != "" {
		return *db.Table(d.TableName), nil
	}
	return db, err
}

func (d *Mysql) Insert(model interface{}) error {
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
		util.LogError(fmt.Sprintf("insert data error:%s", errInsert.Error()))
	}

	return errInsert
}

func (d *Mysql) Select(condition string, data interface{}) error {

	orm, err := d.GetReadOrm()

	if err != nil {
		return err
	}

	defer orm.Close()

	errFind := orm.Where(condition).Find(&data).Error

	if errFind != nil {
		util.LogError(fmt.Sprintf("mysql select table %s error:%s", d.TableName, errFind.Error()))
		return errFind
	}

	return nil
}
