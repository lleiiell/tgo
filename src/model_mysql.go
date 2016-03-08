package tgo

import (
	"time"
)

type ModelMysql struct {
	Id         int `sql:"AUTO_INCREMENT"`
	Created_at time.Time
	Updated_at time.Time
}

func (m *ModelMysql) InitTime(t time.Time) {
	m.Created_at = t
	m.Updated_at = t
}
func (m *ModelMysql) SetUpdatedTime(t time.Time) {
	m.Updated_at = t
}

func (m *ModelMysql) SetId(id int) {
	m.Id = id
}

func (m *ModelMysql) GetId() int {
	return m.Id
}
