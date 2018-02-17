package utils

import (
	"errors"
	"fmt"
	"reflect"
)

// FillStruct will attempt to fill a struct using values from a map
// An error is returned if the struct doesn't have a map key
// or if the struct type doesn't match the map key's value's type.
// FillStruct will panic if s is not a pointer.
func FillStruct(s interface{}, m map[string]interface{}) error {
	if reflect.ValueOf(s).Kind() != reflect.Ptr {
		panic("s must be a pointer to a struct")
	}

	for k, v := range m {
		err := setField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
}
