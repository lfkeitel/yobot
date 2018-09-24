package utils

import (
	"encoding/json"
	"reflect"
)

// FillStruct will attempt to fill a struct using values from a map.
// FillStruct will panic if s is not a pointer.
// FillStruct should only be used for initilization. There may be
// performance issues when using the function frequently.
func FillStruct(s interface{}, m map[string]interface{}) error {
	if reflect.ValueOf(s).Kind() != reflect.Ptr {
		panic("s must be a pointer to a struct")
	}

	// This could probably be done more effeciently using reflect, but
	// I tried and tried and could never get it to work right. At least
	// with this, every possible value* is supported. Is it a bit janky?
	// Probably. But this function should be ran often.
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}
