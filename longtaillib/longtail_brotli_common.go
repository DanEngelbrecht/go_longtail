package longtaillib

// #cgo CFLAGS: -g -std=gnu99 -m64 -pthread -msse4.1 -maes -DLONGTAIL_ASSERTS
// #include "longtail/lib/brotli/ext/common/dictionary.c"
// #include "longtail/lib/brotli/ext/common/transform.c"
import "C"