package longtaillib

// #cgo CFLAGS: -g -std=gnu99 -m64 -pthread -msse4.1 -maes -O3 -D__linux__
// #include "longtail/lib/longtail_platform.c"
import "C"
