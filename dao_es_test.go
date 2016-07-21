package tgo

import (
	"reflect"
	"testing"

	"gopkg.in/olivere/elastic.v3"
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

	err := es.DaoES.Insert("1", model)

	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	}
}

func Benchmark_ESQuery(b *testing.B) {

	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			es := NewESTest()
			conn, err := es.DaoES.GetConnect()
			if err != nil {
				b.Errorf("connect error:%s", err.Error())
			}
			var searchResult *elastic.SearchResult
			searchResult, err = conn.Search().Index(es.IndexName).Do()

			if err != nil {
				b.Errorf("search error:%s", err.Error())
			} else {
				var ttyp ModelMongoHello
				var helloModels []ModelMongoHello
				for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
					t := item.(ModelMongoHello)

					helloModels = append(helloModels, t)
				}
				if len(helloModels) == 0 {
					b.Errorf("len is 0")
				}
			}

		}
	})
}
