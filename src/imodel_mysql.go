package tgo

import (
	"time"
)

type IModelMongo interface {
	InitTime(t time.Time)
	SetUpdatedTime(t time.Time)
	SetId(id int)
	GetId() int
}
