package longtaillib

// #cgo CFLAGS: -g -std=gnu99 -m64 -msse4.1 -maes -pthread -O3
// #include "longtail/lib/brotli/ext/enc/backward_references.c"
import "C"
