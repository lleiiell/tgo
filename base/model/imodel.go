package model

import (
	"time"
)

type IModel interface {
	InitTime(t time.Time)
	SetUpdatedTime(t time.Time)
	SetId(id int64)
	GetId() int64
}
