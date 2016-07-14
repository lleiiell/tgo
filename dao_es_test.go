package tgo

import (
	"testing"
)



type TestDaoES struct {
	DaoES
}

func NewESTest() *TestDaoES {

	dao := &TestDaoES{DaoES{IndexName: "test", TypeName: "hello"}}

	return dao
}

func Test_ESInsert(t *testing.T) {
	es := NewESTest()
	model := &ModelMongoHello{}
	model.HelloWord = "1"
	model.Name = "name2"

	err := es.DaoES.Insert(model)

	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	}
}
