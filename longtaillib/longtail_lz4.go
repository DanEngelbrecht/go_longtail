package longtaillib

// #cgo CFLAGS: -g -std=gnu99 -m64 -pthread -msse4.1 -maes -DLONGTAIL_ASSERTS -DLZ4_DISABLE_DEPRECATE_WARNINGS
// #include "longtail/lib/lz4/longtail_lz4.c"
// #include "longtail/lib/lz4/ext/lz4.c"
// #include "longtail/lib/lz4/ext/lz4frame.c"
// #include "longtail/lib/lz4/ext/lz4hc.c"
import "C"
