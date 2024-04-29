package arch

import (
	"errors"
	"fmt"
	"reflect"
)

// ConvertStruct copies the fields from src to dst if they have the same name and type.
// All fields in dst must be set.
func ConvertStruct(dst interface{}, src interface{}) error {
	srcVal := reflect.ValueOf(src)
	if !isStruct(srcVal.Type()) {
		return fmt.Errorf("expected src to be a struct, got %v", srcVal.Type())
	}

	dstVal := reflect.ValueOf(dst)
	if !isStructPtr(dstVal.Type()) {
		return fmt.Errorf("expected dst to be a pointer to a struct, got %v", dstVal.Type())
	}

	dstElem := dstVal.Elem()
	dstType := dstElem.Type()

	for i := 0; i < dstElem.NumField(); i++ {
		dstField := dstElem.Field(i)
		dstFieldType := dstType.Field(i)
		if !dstField.CanSet() {
			return fmt.Errorf("field %s is not settable", dstFieldType.Name)
		}
		srcField := srcVal.FieldByName(dstFieldType.Name)
		if !srcField.IsValid() {
			return fmt.Errorf("field %s not found", dstFieldType.Name)
		}
		if srcField.Type() != dstField.Type() {
			return fmt.Errorf("field %s has different type", dstFieldType.Name)
		}
		dstField.Set(srcField)
	}

	return nil
}

// CanPopulateStruct returns an error if it is not possible to populate a struct with the values returned by the Get<field name> methods in src.
func CanPopulateStruct(dstType reflect.Type, srcType reflect.Type) error {
	if !isStruct(srcType) && !isStructPtr(srcType) {
		return errors.New("src is not a struct or a pointer to a struct")
	}
	if !isStructPtr(dstType) {
		return errors.New("dst is not a pointer to a struct")
	}

	dstElemType := dstType.Elem()

	// TODO: checks to avoid panics
	// TODO: dst vs dst

	for i := 0; i < dstElemType.NumField(); i++ {
		dstField := dstElemType.Field(i)
		dstFieldType := dstElemType.Field(i)
		getMethodName := "Get" + dstFieldType.Name
		srcGetMethod, ok := srcType.MethodByName(getMethodName)
		if !ok {
			return fmt.Errorf("method %s not found", getMethodName)
		}
		if srcGetMethod.Type.NumOut() != 1 {
			return errors.New("method has more than one return value")
		}
		if srcGetMethod.Type.Out(0) != dstField.Type {
			return fmt.Errorf("field %s has different type", dstFieldType.Name)
		}
	}

	return nil
}

// PopulateStruct sets all the fields in dst to the values returned by the Get<field name> methods in src.
func PopulateStruct(dst interface{}, src interface{}) error {
	if err := CanPopulateStruct(reflect.TypeOf(dst), reflect.TypeOf(src)); err != nil {
		return err
	}
	var (
		srcVal      = reflect.ValueOf(src)
		dstVal      = reflect.ValueOf(dst)
		dstElem     = dstVal.Elem()
		dstElemType = dstElem.Type()
	)
	for i := 0; i < dstElem.NumField(); i++ {
		var (
			dstField      = dstElem.Field(i)
			dstTypeField  = dstElemType.Field(i)
			getMethodName = "Get" + dstTypeField.Name
			srcGetMethod  = srcVal.MethodByName(getMethodName)
			values        = srcGetMethod.Call(nil)
			value         = values[0]
		)
		dstField.Set(value)
	}
	return nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}
