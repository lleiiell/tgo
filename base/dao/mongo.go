package base

import (
	"configs"
	"errors"
	"fmt"
	"github.com/tgo/base/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
	"util"
)

type BaseDaoMongo struct {
	CollectionName string

	AutoIncrementId bool

	PrimaryKey string
}

type BaseStructMongo struct {
	Id         int64     `bson:"_id,omitempty"`
	Created_at time.Time `bson:"created_at,omitempty"`
	Updated_at time.Time `bson:"updated_at,omitempty"`
}

type MongoCounter struct {
	Id  string `bson:"_id,omitempty"`
	Seq int64  `bson:"seq,omitempty"`
}

func NewDaoMongo() *BaseDaoMongo {

	return &BaseDaoMongo{}
}

func (m *BaseDaoMongo) getSession() (*mgo.Session, string, error) {

	config := configs.NewDb()

	configMongo := config.Mongo.GetMongoConfig()

	if configMongo == nil || configMongo.Servers == "" || configMongo.DbName == "" {
		return nil, "", errors.New("config error")
	}

	if strings.Trim(configMongo.Read_option, " ") == "" {
		configMongo.Read_option = "nearest"
	}

	connectionString := fmt.Sprintf("mongodb://%s", configMongo.Servers)

	session, err := mgo.Dial(connectionString)

	session.SetSocketTimeout(time.Duration(configMongo.Timeout) * time.Millisecond)

	if err != nil {

		util.LogErrorf("connect to mongo server error:%s,%s", err.Error(), connectionString)
		return nil, "", err
	}
	/*
		if configs.IsEnvDev() {
			defaultLogger := log.New(os.Stderr, "[Mongo] ", log.Ldate|log.Ltime|log.Lshortfile)

			mgo.SetLogger(defaultLogger)

			mgo.SetDebug(true)
		}*/

	return session, configMongo.DbName, err
}

func (m *BaseDaoMongo) GetNextSequence() (int64, error) {

	session, dbName, err := m.getSession()

	if err != nil {
		return 0, err
	}
	defer session.Close()

	c := session.DB(dbName).C("counters")

	condition := bson.M{"_id": m.CollectionName}

	//_, errUpsert := c.Upsert(condition, bson.M{"$inc": bson.M{"seq": 1}})

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	result := bson.M{}

	_, errApply := c.Find(condition).Apply(change, &result)

	if errApply != nil {
		util.LogErrorf("mongo findAndModify counter %s failed:%s", m.CollectionName, errApply.Error())
		return 0, errApply
	}

	setInt, resultNext := result["seq"].(int)

	if !resultNext {
		util.LogErrorf("mongo findAndModify get counter %sfailed", m.CollectionName)
	}
	seq := int64(setInt)

	return seq, nil
}
func (m *BaseDaoMongo) GetById(id int64, data interface{}) error {
	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	errFind := session.DB(dbName).C(m.CollectionName).Find(bson.M{"_id": id}).One(data)

	if errFind != nil {
		util.LogErrorf("mongo %s get id failed:%v", m.CollectionName, errFind.Error())
	}

	return errFind
}
func (m *BaseDaoMongo) Insert(data models.IModelMongo) error {

	if m.AutoIncrementId {

		id, err := m.GetNextSequence()

		if err != nil {
			return err
		}
		data.SetId(id)
	}

	data.InitTime(time.Now())

	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	coll := session.DB(dbName).C(m.CollectionName)

	errInsert := coll.Insert(data)

	if errInsert != nil {

		util.LogErrorf("mongo %s insert failed:%v", m.CollectionName, errInsert.Error())

		return errInsert
	}
	return nil
}

func (m *BaseDaoMongo) InsertM(data []models.IModelMongo) error {

	for _, item := range data {
		if m.AutoIncrementId {

			id, err := m.GetNextSequence()

			if err != nil {
				return err
			}
			item.SetId(id)
		}

		item.InitTime(time.Now())
	}

	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	coll := session.DB(dbName).C(m.CollectionName)

	var idata []interface{}

	for i := 0; i < len(data); i++ {
		idata = append(idata, data[i])
	}
	errInsert := coll.Insert(idata...)

	if errInsert != nil {

		util.LogErrorf("mongo %s insertM failed:%v", m.CollectionName, errInsert.Error())

		return errInsert
	}
	return nil
}

func (m *BaseDaoMongo) Count(condition interface{}) (int, error) {

	session, dbName, err := m.getSession()

	if err != nil {
		return 0, err
	}

	defer session.Close()

	count, errCount := session.DB(dbName).C(m.CollectionName).Find(condition).Count()

	if errCount != nil {

		util.LogErrorf("mongo %s count failed:%v", m.CollectionName, errCount.Error())

	}
	return count, errCount
}

func (m *BaseDaoMongo) Find(condition interface{}, limit int, skip int, data interface{}, sortFields ...string) error {

	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	s := session.DB(dbName).C(m.CollectionName).Find(condition)

	if len(sortFields) > 0 {
		s = s.Sort(sortFields...)
	}

	if skip > 0 {
		s = s.Skip(skip)
	}

	if limit > 0 {
		s = s.Limit(limit)
	}

	errSelect := s.All(data)

	if errSelect != nil {

		util.LogErrorf("mongo %s find failed:%v", m.CollectionName, errSelect.Error())

	}

	return errSelect
}

func (m *BaseDaoMongo) Distinct(condition interface{}, field string, data interface{}) error {

	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	errDistinct := session.DB(dbName).C(m.CollectionName).Find(condition).Distinct(field, data)

	if errDistinct != nil {

		util.LogErrorf("mongo %s distinct failed:%s", m.CollectionName, errDistinct.Error())

	}

	return errDistinct
}

func (m *BaseDaoMongo) Sum(condition interface{}, sumField string) (int, error) {
	session, dbName, err := m.getSession()

	if err != nil {
		return 0, err
	}

	defer session.Close()

	coll := session.DB(dbName).C(m.CollectionName)

	sumValue := bson.M{"$sum": sumField}

	pipe := coll.Pipe([]bson.M{{"$match": condition}, {"$group": bson.M{"_id": 1, "sum": sumValue}}})

	type SumStruct struct {
		_id int
		Sum int
	}

	var result SumStruct

	errPipe := pipe.One(&result)

	if errPipe != nil {
		util.LogErrorf("mongo %s sum failed: %s", m.CollectionName, errPipe.Error())

		return 0, errPipe
	}

	return result.Sum, nil
}

func (m *BaseDaoMongo) DistinctCount(condition interface{}, field string) (int, error) {
	session, dbName, err := m.getSession()

	if err != nil {
		return 0, err
	}

	defer session.Close()

	coll := session.DB(dbName).C(m.CollectionName)

	pipe := coll.Pipe([]bson.M{{"$match": condition}, {"$group": bson.M{"_id": fmt.Sprintf("$%s", field)}},
		{"$group": bson.M{"_id": "_id", "count": bson.M{"$sum": 1}}}})

	type CountStruct struct {
		_id   int
		Count int
	}

	var result CountStruct

	errPipe := pipe.One(&result)

	if errPipe != nil {
		util.LogErrorf("mongo %s distinct count failed: %s", m.CollectionName, errPipe.Error())

		return 0, errPipe
	}

	return result.Count, nil
}

func (m *BaseDaoMongo) Update(condition interface{}, data map[string]interface{}) error {
	session, dbName, err := m.getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	coll := session.DB(dbName).C(m.CollectionName)

	setBson := bson.M{}
	for key, value := range data {
		setBson[fmt.Sprintf("%s", key)] = value
	}

	updateData := bson.M{"$set": setBson, "$currentDate": bson.M{"updated_at": true}}

	errUpdate := coll.Update(condition, updateData)

	if errUpdate != nil {
		util.LogErrorf("mongo %s update failed: %s", m.CollectionName, errUpdate.Error())
	}

	return errUpdate
}
