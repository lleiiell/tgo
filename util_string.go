package tgo

import (
	"strconv"
	"strings"
)

func UtilIsEmpty(data string) bool {
	return strings.Trim(data, " ") == ""
}

func UtilGetStringFromIntArray(data []int, sep string) string {

	dataStr := UtilGetStringArrayFromIntArray(data)

	return strings.Join(dataStr, sep)

}

func UtilGetStringFromInt64Array(data []int64, sep string) string {

	dataStr := UtilGetStringArrayFromInt64Array(data)

	return strings.Join(dataStr, sep)

}

func UtilGetStringArrayFromIntArray(data []int) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.Itoa(item)

		model = append(model, m)

	}
	return model
}
func UtilGetStringArrayFromInt64Array(data []int64) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.FormatInt(item, 10)

		model = append(model, m)

	}
	return model
}

func UtilSplitToIntArray(data string, sep string) []int {
	var model []int

	dataArray := strings.Split(data, sep)

	for _, item := range dataArray {
		m, err := strconv.Atoi(item)

		if err != nil {
			continue
		}

		model = append(model, m)
	}
	return model
}

func UtilSplitToInt64Array(data string, sep string) []int64 {
	var model []int64

	dataArray := strings.Split(data, sep)

	for _, item := range dataArray {
		m, err := strconv.ParseInt(item, 10, 64)

		if err != nil {
			continue
		}

		model = append(model, m)
	}
	return model
}
