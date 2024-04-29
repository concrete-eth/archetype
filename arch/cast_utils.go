package arch

import (
	"errors"
	"fmt"
	"reflect"
)

// ConvertStruct copies the fields from src to dest if they have the same name and type.
// All fields in dest must be set.
func ConvertStruct(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	if !isStruct(srcVal.Type()) {
		return fmt.Errorf("expected src to be a struct, got %v", srcVal.Type())
	}

	destVal := reflect.ValueOf(dest)
	if !isStructPtr(destVal.Type()) {
		return fmt.Errorf("expected dest to be a pointer to a struct, got %v", destVal.Type())
	}

	destElem := destVal.Elem()
	destType := destElem.Type()

	for i := 0; i < destElem.NumField(); i++ {
		destField := destElem.Field(i)
		destFieldType := destType.Field(i)
		if !destField.CanSet() {
			return fmt.Errorf("field %s is not settable", destFieldType.Name)
		}
		srcField := srcVal.FieldByName(destFieldType.Name)
		if !srcField.IsValid() {
			return fmt.Errorf("field %s not found", destFieldType.Name)
		}
		if srcField.Type() != destField.Type() {
			return fmt.Errorf("field %s has different type", destFieldType.Name)
		}
		destField.Set(srcField)
	}

	return nil
}

// CanPopulateStruct returns an error if it is not possible to populate a struct with the values returned by the Get<field name> methods in src.
func CanPopulateStruct(srcType reflect.Type, destType reflect.Type) error {
	if !isStruct(srcType) && !isStructPtr(srcType) {
		return errors.New("src is not a struct or a pointer to a struct")
	}
	if !isStructPtr(destType) {
		return errors.New("dest is not a pointer to a struct")
	}

	destElemType := destType.Elem()

	// TODO: checks to avoid panics
	// TODO: dest vs dst

	for i := 0; i < destElemType.NumField(); i++ {
		destField := destElemType.Field(i)
		destFieldType := destElemType.Field(i)
		getMethodName := "Get" + destFieldType.Name
		srcGetMethod, ok := srcType.MethodByName(getMethodName)
		if !ok {
			return fmt.Errorf("method %s not found", getMethodName)
		}
		if srcGetMethod.Type.NumOut() != 1 {
			return errors.New("method has more than one return value")
		}
		if srcGetMethod.Type.Out(0) != destField.Type {
			return fmt.Errorf("field %s has different type", destFieldType.Name)
		}
	}

	return nil
}

// PopulateStruct sets all the fields in dest to the values returned by the Get<field name> methods in src.
func PopulateStruct(src interface{}, dest interface{}) error {
	if err := CanPopulateStruct(reflect.TypeOf(src), reflect.TypeOf(dest)); err != nil {
		return err
	}
	var (
		srcVal       = reflect.ValueOf(src)
		destVal      = reflect.ValueOf(dest)
		destElem     = destVal.Elem()
		destElemType = destElem.Type()
	)
	for i := 0; i < destElem.NumField(); i++ {
		var (
			destField     = destElem.Field(i)
			destTypeField = destElemType.Field(i)
			getMethodName = "Get" + destTypeField.Name
			srcGetMethod  = srcVal.MethodByName(getMethodName)
			values        = srcGetMethod.Call(nil)
			value         = values[0]
		)
		destField.Set(value)
	}
	return nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}
