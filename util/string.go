package util

import (
	"strconv"
	"strings"
)

func IsEmpty(data string) bool {
	return strings.Trim(data, " ") == ""
}

func GetStringFromIntArray(data []int, sep string) string {

	dataStr := GetStringArrayFromIntArray(data)

	return strings.Join(dataStr, sep)

}

func GetStringFromInt64Array(data []int64, sep string) string {

	dataStr := GetStringArrayFromInt64Array(data)

	return strings.Join(dataStr, sep)

}

func GetStringArrayFromIntArray(data []int) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.Itoa(item)

		model = append(model, m)

	}
	return model
}
func GetStringArrayFromInt64Array(data []int64) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.FormatInt(item, 10)

		model = append(model, m)

	}
	return model
}

func SplitToIntArray(data string, sep string) []int {
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

func SplitToInt64Array(data string, sep string) []int64 {
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
