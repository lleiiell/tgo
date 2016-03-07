package model

import (
	"time"
)

type Mysql struct {
	Id         int `sql:"AUTO_INCREMENT"`
	Created_at time.Time
	Updated_at time.Time
}

func (m *Mysql) InitTime(t time.Time) {
	m.Created_at = t
	m.Updated_at = t
}
func (m *Mysql) SetUpdatedTime(t time.Time) {
	m.Updated_at = t
}

func (m *Mysql) SetId(id int64) {
	m.Id = id
}

func (m *Mysql) GetId() int64 {
	return m.Id
}
