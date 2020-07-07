package reflect_test

import (
	"fmt"
	corereflect "reflect"
	"testing"
	"unsafe"

	"github.com/goccy/go-reflect"
)

func TestTypeID(t *testing.T) {
	intID := reflect.TypeID(int(0))
	if intID != reflect.TypeID(int(0)) {
		t.Fatal("failed to get typeid")
	}
}

func TestTypeAndPtrOf(t *testing.T) {
	typ, ptr := reflect.TypeAndPtrOf(int(10))
	if typ.Kind() != reflect.Int {
		t.Fatal("failed to get type")
	}
	if *(*int)(ptr) != 10 {
		t.Fatal("failed to get ptr")
	}
}

func TestValueNoEscapeOf(t *testing.T) {
	v := reflect.ValueNoEscapeOf(&struct{ I int }{I: 10})
	if v.Elem().Field(0).Int() != 10 {
		t.Fatal("failed to create reflect.Value from ValueNoEscapeOf")
	}
}

func TestToReflectType(t *testing.T) {
	typ := reflect.ToReflectType(reflect.TypeOf(1))
	if fmt.Sprintf("%T", typ) != "*reflect.rtype" {
		t.Fatalf("failed to convert reflect.Type")
	}
}

func TestToReflectValue(t *testing.T) {
	v := reflect.ToReflectValue(reflect.ValueOf(1))
	if fmt.Sprintf("%T", v) != "reflect.Value" {
		t.Fatalf("failed to convert reflect.Value")
	}
}

func TestToType(t *testing.T) {
	typ := reflect.ToType(corereflect.TypeOf(1))
	if fmt.Sprintf("%T", typ) != "*reflect.rtype" {
		t.Fatalf("failed to convert reflect.Type")
	}
}

func TestToValue(t *testing.T) {
	v := reflect.ToValue(corereflect.ValueOf(1))
	if fmt.Sprintf("%T", v) != "reflect.Value" {
		t.Fatalf("failed to convert reflect.Value")
	}
}

func TestNewAt(t *testing.T) {
	var i int

	ivalFromField := reflect.ValueOf(&struct{ foo int }{100}).Elem().Field(0)
	ival := reflect.ValueOf(&i).Elem()

	v := reflect.NewAt(ivalFromField.Type(), unsafe.Pointer(ivalFromField.UnsafeAddr())).Elem()
	ival.Set(v)
	v.Set(ival)
	if v.Int() != 100 {
		t.Fatal("failed to NewAt")
	}
}

func TestTypeBits(t *testing.T) {
	bits := reflect.TypeOf(1).Bits()
	if bits != 1<<6 {
		t.Fatal("failed to get bits from type")
	}
}

func TestTypeIsVariadic(t *testing.T) {
	if !reflect.TypeOf(func(...int) {}).IsVariadic() {
		t.Fatal("doesn't work IsVariadic")
	}
}

func TestTypeFieldByNameFunc(t *testing.T) {
	_, ok := reflect.TypeOf(struct {
		Foo int
	}{}).FieldByNameFunc(func(name string) bool {
		return name == "Foo"
	})
	if !ok {
		t.Fatal("failed to FieldByNameFunc")
	}
}

func TestTypeLen(t *testing.T) {
	if reflect.TypeOf([3]int{}).Len() != 3 {
		t.Fatal("failed to Type.Len")
	}
}

func TestTypeOut(t *testing.T) {
	if reflect.TypeOf(func() int { return 0 }).Out(0).Kind() != reflect.Int {
		t.Fatal("failed to get output parameter")
	}
}

func TestTypeCanInterface(t *testing.T) {
	v := "hello"
	if !reflect.ValueOf(v).CanInterface() {
		t.Fatal("failed to Type.CanInterface")
	}
}

func TestValueFieldByNameFunc(t *testing.T) {
	field := reflect.ValueOf(struct {
		Foo int
	}{}).FieldByNameFunc(func(name string) bool {
		return name == "Foo"
	})
	if field.Type().Kind() != reflect.Int {
		t.Fatal("failed to FieldByNameFunc")
	}
}
