package types

import (
	"reflect"
	"testing"
)

func TestConvertStruct(t *testing.T) {
	str := "string"
	testData := []struct {
		name    string
		src     interface{}
		dstType reflect.Type
		ok      bool
	}{
		{
			name:    "0-field empty src and string dst",
			src:     struct{}{},
			dstType: reflect.TypeOf(&str),
			ok:      false,
		},
		{
			name:    "string src and 0-field dst",
			src:     str,
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      false,
		},
		{
			name:    "0-field empty src and 0-field dst",
			src:     struct{}{},
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      true,
		},
		{
			name: "0-field empty src and n-field dst",
			src:  struct{}{},
			dstType: reflect.TypeOf(&struct {
				A int
				B string
			}{}),
			ok: false,
		},
		{
			name: "n-field empty src and 0-field dst",
			src: struct {
				A int
				B string
			}{},
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      true,
		},
		{
			name: "n-field empty src and n-field dst",
			src: struct {
				A int
				B string
			}{},
			dstType: reflect.TypeOf(&struct {
				A int
				B string
			}{}),
			ok: true,
		},
		{
			name: "n-field empty src and m-field dst",
			src: struct {
				A int
				B string
			}{},
			dstType: reflect.TypeOf(&struct {
				C int
				D int
			}{}),
			ok: false,
		},
		{
			name: "n-field empty src and t-field dst",
			src: struct {
				A int
				B string
			}{},
			dstType: reflect.TypeOf(&struct {
				A string
				B int
			}{}),
			ok: false,
		},
		{
			name: "n-field non-empty src and 0-field dst",
			src: struct {
				A int
				B string
			}{
				A: 1,
				B: "2",
			},
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      true,
		},
		{
			name: "n-field non-empty src and n-field dst",
			src: struct {
				A int
				B string
			}{
				A: 1,
				B: "2",
			},
			dstType: reflect.TypeOf(&struct {
				A int
				B string
			}{}),
			ok: true,
		},
		{
			name: "n-field non-empty src and m-field dst",
			src: struct {
				A int
				B string
			}{
				A: 1,
				B: "2",
			},
			dstType: reflect.TypeOf(&struct {
				C int
				D string
			}{}),
			ok: false,
		},
		{
			name: "n-field non-empty src and t-field dst",
			src: struct {
				A int
				B string
			}{
				A: 1,
				B: "2",
			},
			dstType: reflect.TypeOf(&struct {
				A string
				B int
			}{}),
			ok: false,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal("PANIC", r)
				}
			}()
			src := data.src
			dst := reflect.New(data.dstType.Elem()).Interface()
			err := ConvertStruct(src, dst)
			if !data.ok {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			srcVal := reflect.ValueOf(src)
			dstVal := reflect.ValueOf(dst).Elem()
			dstType := dstVal.Type()
			for i := 0; i < dstVal.NumField(); i++ {
				dstFieldVal := dstVal.Field(i)
				dstFieldType := dstType.Field(i)
				srcFieldVal := srcVal.FieldByName(dstFieldType.Name)
				if !srcFieldVal.IsValid() {
					t.Fatalf("field %s is not valid", dstFieldType.Name)
				}
				if !reflect.DeepEqual(dstFieldVal.Interface(), srcFieldVal.Interface()) {
					t.Fatalf("field %s is not equal", dstFieldType.Name)
				}
			}
		})
	}
}

type testSrcStruct struct {
	_a int
	_b string
}

func (t *testSrcStruct) GetA() int {
	return t._a
}

func (t *testSrcStruct) GetB() string {
	return t._b
}

func TestCanPopulateStruct(t *testing.T) {
	testData := []struct {
		name    string
		srcType reflect.Type
		dstType reflect.Type
		ok      bool
	}{
		{
			name:    "string src and 0-field dst",
			srcType: reflect.TypeOf("string"),
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      false,
		},
		{
			name:    "string dst",
			srcType: reflect.TypeOf(&testSrcStruct{}),
			dstType: reflect.TypeOf("string"),
			ok:      false,
		},
		{
			name:    "0-field dst",
			srcType: reflect.TypeOf(&testSrcStruct{}),
			dstType: reflect.TypeOf(&struct{}{}),
			ok:      true,
		},
		{
			name:    "n-field dst",
			srcType: reflect.TypeOf(&testSrcStruct{}),
			dstType: reflect.TypeOf(&struct {
				A int
				B string
			}{}),
			ok: true,
		},
		{
			name:    "m-field dst",
			srcType: reflect.TypeOf(&testSrcStruct{}),
			dstType: reflect.TypeOf(&struct {
				C int
				D string
			}{}),
			ok: false,
		},
		{
			name:    "t-field dst",
			srcType: reflect.TypeOf(&testSrcStruct{}),
			dstType: reflect.TypeOf(&struct {
				A string
				B int
			}{}),
			ok: false,
		},
	}
	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal("PANIC", r)
				}
			}()
			err := CanPopulateStruct(data.srcType, data.dstType)
			if !data.ok {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
