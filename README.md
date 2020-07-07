# go-reflect

![Go](https://github.com/goccy/go-reflect/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-reflect?status.svg)](https://pkg.go.dev/github.com/goccy/go-reflect?tab=doc)
[![codecov](https://codecov.io/gh/goccy/go-reflect/branch/master/graph/badge.svg)](https://codecov.io/gh/goccy/go-reflect)
[![Go Report Card](https://goreportcard.com/badge/github.com/goccy/go-reflect)](https://goreportcard.com/report/github.com/goccy/go-reflect)

Zero-allocation (runtime path dependent) reflection library for Go

# Features

- 100% Compatibility APIs with `reflect` library
- No allocation occurs when using the reflect.Type features
- You can choose to escape ( `reflect.ValueOf` ) or noescape ( `reflect.ValueNoEscapeOf` ) when creating reflect.Value

# Status

All the tests in the reflect library have been passed
except the tests that use some private functions.

# Installation

```bash
go get github.com/goccy/go-reflect
```

# How to use

Replace import statement from `reflect` to `github.com/goccy/go-reflect`

```bash
-import "reflect"
+import "github.com/goccy/go-reflect"
```

# Benchmarks

Source https://github.com/goccy/go-reflect/blob/master/bechmark_test.go

## Benchmark about reflect.Type

```
$ go test -bench TypeOf
```

```
goos: darwin
goarch: amd64
pkg: github.com/goccy/go-reflect
Benchmark_TypeOf_Reflect-12             100000000               13.8 ns/op             8 B/op          1 allocs/op
Benchmark_TypeOf_GoReflect-12           2000000000               1.70 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/goccy/go-reflect     5.369s
```

## Benchmark about reflect.Value

```
$ go test -bench ValueOf
```

```
goos: darwin
goarch: amd64
pkg: github.com/goccy/go-reflect
Benchmark_ValueOf_Reflect-12            100000000               13.0 ns/op             8 B/op          1 allocs/op
Benchmark_ValueOf_GoReflect-12          300000000                4.64 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/goccy/go-reflect     3.578s
```

# Real World Example

## Implements Fast Marshaler

I would like to introduce the technique I use for [github.com/goccy/go-json](https://github.com/goccy/go-json).  
Using this technique, allocation can be suppressed to once for any marshaler.  

<img src="https://raw.githubusercontent.com/wiki/goccy/go-reflect/benchmark_marshaler_test.svg"></img>

The benchmark result is as follows.  

<img src="https://raw.githubusercontent.com/wiki/goccy/go-reflect/benchmark.svg"></img>

