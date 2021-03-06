#pragma once

#include "../../src/longtail.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct Longtail_CompressionAPI_CompressionContext* Longtail_CompressionAPI_HCompressionContext;
typedef struct Longtail_CompressionAPI_DecompressionContext* Longtail_CompressionAPI_HDecompressionContext;

LONGTAIL_EXPORT extern struct Longtail_BlockStoreAPI* Longtail_CreateCompressBlockStoreAPI(
    struct Longtail_BlockStoreAPI* backing_block_store,
    struct Longtail_CompressionRegistryAPI* compression_registry);

#ifdef __cplusplus
}
#endif
