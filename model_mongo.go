package tgo

import (
	"time"
)

type ModelMongo struct {
	Id         int64     `bson:"_id,omitempty" json:"id"`
	Created_at time.Time `bson:"created_at,omitempty" json:"created_at"`
	Updated_at time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

func (m *ModelMongo) GetCreatedTime() time.Time {
	return m.Created_at
}

func (m *ModelMongo) InitTime(t time.Time) {
	m.Created_at = t
	m.Updated_at = t
}
func (m *ModelMongo) SetUpdatedTime(t time.Time) {
	m.Updated_at = t
}

func (m *ModelMongo) SetId(id int64) {
	m.Id = id
}

func (m *ModelMongo) GetId() int64 {
	return m.Id
}
