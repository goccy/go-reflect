package reflect_test

import (
	"reflect"
	"testing"

	goreflect "github.com/goccy/go-reflect"
)

func kindFromReflect(v interface{}) reflect.Kind {
	return reflect.TypeOf(v).Kind()
}

func kindFromGoReflect(v interface{}) goreflect.Kind {
	return goreflect.TypeOf(v).Kind()
}

func f(_ interface{}) {}

func valueFromReflect(v interface{}) {
	f(reflect.ValueOf(v))
}

func valueFromGoReflect(v interface{}) {
	f(goreflect.ValueOf(v))
}

func Benchmark_TypeOf_Reflect(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var v struct {
			i int
		}
		kindFromReflect(&v)
	}
}

func Benchmark_TypeOf_GoReflect(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var v struct {
			i int
		}
		kindFromGoReflect(&v)
	}
}

func Benchmark_ValueOf_Reflect(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		valueFromReflect(&struct {
			I int
		}{I: 10})
	}
}

func Benchmark_ValueOf_GoReflect(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		valueFromGoReflect(&struct {
			I int
		}{I: 10})
	}
}
