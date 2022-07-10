package reflect_test

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"unsafe"

	"github.com/goccy/go-reflect"
)

var (
	typeToEncoderMap sync.Map
	bufpool          = sync.Pool{
		New: func() interface{} {
			return &buffer{
				b: make([]byte, 0, 1024),
			}
		},
	}
)

type buffer struct {
	b        []byte
	KeepRefs []unsafe.Pointer
}

type encoder func(*buffer, unsafe.Pointer) error

func Marshal(v interface{}) ([]byte, error) {

	// Technique 1.
	// Get type information and pointer from interface{} value without allocation.
	typ, ptr := reflect.TypeAndPtrOf(v)
	typeID := reflect.TypeID(v)

	// Technique 2.
	// Reuse the buffer once allocated using sync.Pool
	buf := bufpool.Get().(*buffer)
	buf.b = buf.b[:0]
	defer bufpool.Put(buf)

	buf.KeepRefs = buf.KeepRefs[:0]
	buf.KeepRefs = append(buf.KeepRefs, ptr)

	// Technique 3.
	// builds a optimized path by typeID and caches it
	if enc, ok := typeToEncoderMap.Load(typeID); ok {
		if err := enc.(encoder)(buf, ptr); err != nil {
			return nil, err
		}

		// allocate a new buffer required length only
		b := make([]byte, len(buf.b))
		copy(b, buf.b)
		return b, nil
	}

	// First time,
	// builds an optimized path by type and caches it with typeID.
	enc, err := compile(typ)
	if err != nil {
		return nil, err
	}
	typeToEncoderMap.Store(typeID, enc)
	if err := enc(buf, ptr); err != nil {
		return nil, err
	}

	// allocate a new buffer required length only
	b := make([]byte, len(buf.b))
	copy(b, buf.b)
	return b, nil
}

func compile(typ reflect.Type) (encoder, error) {
	switch typ.Kind() {
	case reflect.Slice:
		return compileSlice(typ)
	case reflect.Map:
		return compileMap(typ)
	case reflect.Struct:
		return compileStruct(typ)
	case reflect.Int:
		return compileInt(typ)
	case reflect.String:
		return compileString(typ)

	}
	return nil, errors.New("unsupported type")
}

func compileString(_ reflect.Type) (encoder, error) {
	return func(buf *buffer, ptr unsafe.Pointer) error {
		sVal := *(*string)(ptr)
		buf.b = strconv.AppendQuote(buf.b, sVal)

		return nil
	}, nil
}

func compileStruct(typ reflect.Type) (encoder, error) {
	encoders := []encoder{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		enc, err := compile(field.Type)
		if err != nil {
			return nil, err
		}
		offset := field.Offset

		indirect := IfaceIndir(field.Type)
		if indirect && field.Type.Kind() == reflect.Map {
			encoders = append(encoders, func(buf *buffer, p unsafe.Pointer) error {
				// use unsafe.Add if you are using go1.17+
				return enc(buf, ptrOfPtr(unsafe.Pointer(uintptr(p)+offset)))
			})
		} else {
			encoders = append(encoders, func(buf *buffer, p unsafe.Pointer) error {
				return enc(buf, unsafe.Pointer(uintptr(p)+offset))
			})
		}
	}

	return func(buf *buffer, p unsafe.Pointer) error {
		buf.b = append(buf.b, '{')
		for _, enc := range encoders {
			if err := enc(buf, p); err != nil {
				return err
			}
		}
		buf.b = append(buf.b, '}')
		return nil
	}, nil
}

func compileInt(typ reflect.Type) (encoder, error) {
	return func(buf *buffer, p unsafe.Pointer) error {
		value := *(*int)(p)
		buf.b = strconv.AppendInt(buf.b, int64(value), 10)
		return nil
	}, nil
}

func compileSlice(typ reflect.Type) (encoder, error) {
	var enc encoder
	var err error

	if typ.Elem().Kind() == reflect.Map {
		enc, err = compile(reflect.PtrTo(typ.Elem()))
	} else {
		enc, err = compile(typ.Elem())
	}

	if err != nil {
		return nil, err
	}

	offset := typ.Elem().Size()
	return func(buf *buffer, p unsafe.Pointer) error {
		// no data ptr, nil slice
		if uintptr(p) == 0 {
			buf.b = append(buf.b, "null"...)
			return nil
		}

		sh := *(*reflect.SliceHeader)(p)

		buf.b = append(buf.b, '[')
		for i := 0; i < sh.Len; i++ {
			// use unsafe.Add in go1.17+
			if err := enc(buf, unsafe.Pointer(sh.Data+offset*uintptr(i))); err != nil {
				return err
			}
			buf.b = append(buf.b, ',')
		}

		buf.b = append(buf.b, ']')
		return nil
	}, nil
}

// these go private methods work in go1.18, you may need to test in new version

//go:linkname IfaceIndir reflect.ifaceIndir
//go:noescape
func IfaceIndir(typ reflect.Type) bool

func compileMap(typ reflect.Type) (encoder, error) {
	keyType := typ.Key()
	valueType := typ.Elem()

	keyEncoder, err := compile(keyType)
	if err != nil {
		return nil, err
	}

	var valueEncoder encoder

	// need special take care
	if valueType.Kind() == reflect.Map {
		valueEncoder, err = compile(reflect.PtrTo(valueType))
		if err != nil {
			return nil, err
		}
	} else {
		valueEncoder, err = compile(valueType)
		if err != nil {
			return nil, err
		}
	}

	return func(buf *buffer, p unsafe.Pointer) error {
		if uintptr(p) == 0 {
			buf.b = append(buf.b, "null"...)
		}

		mapLen := MapLen(p)
		if mapLen == 0 {
			buf.b = append(buf.b, "{}"...)
			return nil
		}

		iter := hashPool.Get().(*HashIter) // use a sync.Pool

		*iter = HashIter{} // reset all value

		// zero alloc map iter
		MapIterInit(typ, p, iter)
		buf.b = append(buf.b, '{')
		for i := 0; i < mapLen; i++ {
			if err := keyEncoder(buf, MapIterKey(iter)); err != nil {
				return err
			}

			buf.b = append(buf.b, ':')

			if err := valueEncoder(buf, MapIterValue(iter)); err != nil {
				return err
			}

			buf.b = append(buf.b, ',')

			MapIterNext(iter)
		}

		hashPool.Put(iter)

		buf.b = append(buf.b, '}')
		return nil
	}, nil
}

func ptrOfPtr(p unsafe.Pointer) unsafe.Pointer {
	return **(**unsafe.Pointer)(unsafe.Pointer(&p))
}

var hashPool = sync.Pool{
	New: func() interface{} {
		return &HashIter{}
	},
}

// copied from reflect.hiter
type HashIter struct {
	key         unsafe.Pointer
	elem        unsafe.Pointer
	t           unsafe.Pointer
	h           unsafe.Pointer
	buckets     unsafe.Pointer
	bptr        unsafe.Pointer
	overflow    unsafe.Pointer
	oldoverflow unsafe.Pointer
	startBucket uintptr
	offset      uint8
	wrapped     bool
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

//go:linkname MapLen reflect.maplen
//go:noescape
func MapLen(ptr unsafe.Pointer) int

//go:linkname MapIterInit runtime.mapiterinit
//go:noescape
func MapIterInit(mapType reflect.Type, m unsafe.Pointer, it *HashIter)

//go:linkname MapIterNext reflect.mapiternext
//go:noescape
func MapIterNext(it *HashIter)

//go:linkname MapIterKey reflect.mapiterkey
//go:noescape
func MapIterKey(it *HashIter) unsafe.Pointer

func Benchmark_Marshal(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		bytes, err := Marshal(struct{ I int }{10})
		if err != nil {
			b.Fatal(err)
		}
		if string(bytes) != "{10}" {
			b.Fatal("marshal result not expected, actual: " + string(bytes))
		}
	}
}

func Benchmark_Marshal_map(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		bytes, err := Marshal(map[string]int{
			"one": 1,
		})
		if err != nil {
			b.Fatal(err)
		}
		// map iter are unordered by default.
		if string(bytes) != `{"one":1,}` {
			b.Fatal("marshal result not expected, actual: " + string(bytes))
		}
	}
}

func Benchmark_Marshal_slice(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		bytes, err := Marshal([]int{5, 4, 3, 2, 1})
		if err != nil {
			b.Fatal(err)
		}
		if string(bytes) != "[5,4,3,2,1,]" {
			b.Fatal("marshal result not expected, actual: " + string(bytes))
		}
	}
}
