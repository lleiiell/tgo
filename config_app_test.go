package tgo

import (
	"testing"
)

func Test_ConfigAppGetSliceString(t *testing.T) {

	key := "sliceStr"

	var slice []string

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceInt(t *testing.T) {

	key := "sliceInt"

	var slice []int

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}
func Test_ConfigAppGetSliceInt64(t *testing.T) {

	key := "sliceInt64"

	var slice []int64

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceBool(t *testing.T) {

	key := "sliceBool"

	var slice []bool

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}
func Test_ConfigAppGetSliceFloat64(t *testing.T) {

	key := "sliceFloat64"

	var slice []float64

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}
func Test_ConfigAppGetSliceFloat32(t *testing.T) {

	key := "sliceFloat32"

	var slice []float32

	err := ConfigAppGetSlice(key, &slice)

	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}
