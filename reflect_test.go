package reflect_test

import (
	"fmt"
	"testing"

	"github.com/goccy/go-reflect"
)

func TestReflect(t *testing.T) {
	var v int
	fmt.Println("v = ", reflect.TypeOf(v).Kind())
	fmt.Println(reflect.ValueOf(v))
}
