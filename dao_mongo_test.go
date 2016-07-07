package tgo

import (
	"testing"
	"gopkg.in/mgo.v2/bson"
)

type ModelMongoHello struct {
	ModelMongo
	HelloWord string
	Name      string
}

type TestDaoMongo struct {
	DaoMongo
}

func NewMongoTest() *TestDaoMongo {

	dao := &TestDaoMongo{DaoMongo{AutoIncrementId: true, CollectionName: "test"}}

	return dao
}

func Test_MongoInsert(t *testing.T) {
	mongo := NewMongoTest()
	model := &ModelMongoHello{}
	model.HelloWord = "1"
	model.Name = "name2"

	err := mongo.DaoMongo.Insert(model)

	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	}
}

func Test_MongoDistinct(t *testing.T) {
	mongo := NewMongoTest()

	var modelList []string

	condition :=bson.M{}
	field:="name"
	err:=mongo.DaoMongo.Distinct(condition, field, &modelList)
	if err!=nil{
		t.Errorf("distinct error:%s",err.Error())
	}else{
		t.Errorf("data is %v",modelList)
	}
}

func Test_MongoDistinctPage(t *testing.T) {
	mongo := NewMongoTest()

	type dd struct{
		Name string
	}

	var modelList []dd

	condition :=bson.M{}
	field:="name"
	skip := 0
	limit:=1
	sortFields := make(map[string]bool)
	sortFields["name"]= true
	err:=mongo.DaoMongo.DistinctWithPage(condition, field,limit,skip,&modelList,sortFields)
	if err!=nil{
		t.Errorf("distinct error:%s",err.Error())
	}else{
		t.Errorf("data is %v",modelList)
	}
}

func Test_MongoFind(t *testing.T){
	mongo := NewMongoTest()
	condition :=bson.M{}
	limit:=0
	skip:=0
	var sortFileds []string
	var data []ModelMongoHello

	err:=mongo.DaoMongo.Find(condition,limit,skip,&data,sortFileds...)

	if err!=nil{
		t.Errorf("mogno find error,%s",err.Error())
	}else{
		t.Errorf("data is %v",data)
	}
}
