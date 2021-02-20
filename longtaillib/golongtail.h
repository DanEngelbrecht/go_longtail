#ifndef _GNU_SOURCE
    #define _GNU_SOURCE
#endif
#include "longtail/include/src/longtail.h"
#include "longtail/include/lib/bikeshed/longtail_bikeshed.h"
#include "longtail/include/lib/blake2/longtail_blake2.h"
#include "longtail/include/lib/blake3/longtail_blake3.h"
#include "longtail/include/lib/blockstorestorage/longtail_blockstorestorage.h"
#include "longtail/include/lib/brotli/longtail_brotli.h"
#include "longtail/include/lib/cacheblockstore/longtail_cacheblockstore.h"
#include "longtail/include/lib/compressblockstore/longtail_compressblockstore.h"
#include "longtail/include/lib/compressionregistry/longtail_compression_registry.h"
#include "longtail/include/lib/compressionregistry/longtail_full_compression_registry.h"
#include "longtail/include/lib/compressionregistry/longtail_zstd_compression_registry.h"
#include "longtail/include/lib/hashregistry/longtail_hash_registry.h"
#include "longtail/include/lib/hashregistry/longtail_full_hash_registry.h"
#include "longtail/include/lib/hashregistry/longtail_blake3_hash_registry.h"
#include "longtail/include/lib/hpcdcchunker/longtail_hpcdcchunker.h"
#include "longtail/include/lib/lrublockstore/longtail_lrublockstore.h"
#include "longtail/include/lib/ratelimitedprogress/longtail_ratelimitedprogress.h"
#include "longtail/include/lib/shareblockstore/longtail_shareblockstore.h"
#include "longtail/include/lib/filestorage/longtail_filestorage.h"
#include "longtail/include/lib/fsblockstore/longtail_fsblockstore.h"
#include "longtail/include/lib/lz4/longtail_lz4.h"
#include "longtail/include/lib/memstorage/longtail_memstorage.h"
#include "longtail/include/lib/memtracer/longtail_memtracer.h"
#include "longtail/include/lib/meowhash/longtail_meowhash.h"
#include "longtail/include/lib/zstd/longtail_zstd.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#ifdef __cplusplus
extern "C" {
#endif

static struct Longtail_LogField* GetLogField(struct Longtail_LogContext* log_context, int index) { return &log_context->fields[index]; }

void LogProxy_Log(struct Longtail_LogContext* log_context, char* str);

void AssertProxy_Assert(char* expression, char* file, int line);

////////////// Longtail_BlockStoreAPI

struct BlockStoreAPIProxy
{
    struct Longtail_BlockStoreAPI m_API;
    void* m_Context;
};

static void* BlockStoreAPIProxy_GetContext(void* api) { return ((struct BlockStoreAPIProxy*)api)->m_Context; }
void BlockStoreAPIProxy_Dispose(struct Longtail_API* api);
int BlockStoreAPIProxy_PutStoredBlock(struct Longtail_BlockStoreAPI* api, struct Longtail_StoredBlock* stored_block, struct Longtail_AsyncPutStoredBlockAPI* async_complete_api);
int BlockStoreAPIProxy_PreflightGet(struct Longtail_BlockStoreAPI* block_store_api, uint64_t chunk_count, TLongtail_Hash* chunk_hashes);
int BlockStoreAPIProxy_GetStoredBlock(struct Longtail_BlockStoreAPI* api, uint64_t block_hash, struct Longtail_AsyncGetStoredBlockAPI* async_complete_api);
int BlockStoreAPIProxy_GetExistingContent(struct Longtail_BlockStoreAPI* api, uint64_t chunk_count, TLongtail_Hash* chunk_hashes, uint32_t min_block_usage_percent, struct Longtail_AsyncGetExistingContentAPI* async_complete_api);
int BlockStoreAPIProxy_GetStats(struct Longtail_BlockStoreAPI* api, struct Longtail_BlockStore_Stats* out_stats);
int BlockStoreAPIProxy_Flush(struct Longtail_BlockStoreAPI* api, struct Longtail_AsyncFlushAPI* async_complete_api);

static struct Longtail_BlockStoreAPI* CreateBlockStoreProxyAPI(void* context)
{
    struct BlockStoreAPIProxy* api = (struct BlockStoreAPIProxy*)Longtail_Alloc("BlockStoreProxyAPI", sizeof(struct BlockStoreAPIProxy));
    api->m_Context = context;
    return Longtail_MakeBlockStoreAPI(
        api,
        BlockStoreAPIProxy_Dispose,
        BlockStoreAPIProxy_PutStoredBlock,
        (Longtail_BlockStore_PreflightGetFunc)BlockStoreAPIProxy_PreflightGet,
        BlockStoreAPIProxy_GetStoredBlock,
        (Longtail_BlockStore_GetExistingContentFunc)BlockStoreAPIProxy_GetExistingContent,
        BlockStoreAPIProxy_GetStats,
        BlockStoreAPIProxy_Flush);
}

////////////// Longtail_PathFilterAPI

struct PathFilterAPIProxy
{
    struct Longtail_PathFilterAPI m_API;
    void* m_Context;
};

static void* PathFilterAPIProxy_GetContext(void* api) { return ((struct PathFilterAPIProxy*)api)->m_Context; }
void PathFilterAPIProxy_Dispose(struct Longtail_API* api);
int PathFilterAPIProxy_Include(struct Longtail_PathFilterAPI* path_filter_api, char* root_path, char* asset_path, char* asset_name, int is_dir, uint64_t size, uint16_t permissions);

static struct Longtail_PathFilterAPI* CreatePathFilterProxyAPI(void* context)
{
    struct PathFilterAPIProxy* api = (struct PathFilterAPIProxy*)Longtail_Alloc("PathFilterProxyAPI", sizeof(struct PathFilterAPIProxy));
    api->m_Context = context;
    return Longtail_MakePathFilterAPI(
        api,
        PathFilterAPIProxy_Dispose,
        (Longtail_PathFilter_IncludeFunc)PathFilterAPIProxy_Include);   // Constness cast
}

////////////// Longtail_ProgressAPI

struct ProgressAPIProxy
{
    struct Longtail_ProgressAPI m_API;
    void* m_Context;
};

static void* ProgressAPIProxy_GetContext(void* api) { return ((struct ProgressAPIProxy*)api)->m_Context; }
void ProgressAPIProxy_Dispose(struct Longtail_API* api);
void ProgressAPIProxy_OnProgress(struct Longtail_ProgressAPI* progress_api, uint32_t total_count, uint32_t done_count);

static struct Longtail_ProgressAPI* CreateProgressProxyAPI(void* context)
{
    struct ProgressAPIProxy* api = (struct ProgressAPIProxy*)Longtail_Alloc("ProgressProxyAPI", sizeof(struct ProgressAPIProxy));
    api->m_Context = context;
    return Longtail_MakeProgressAPI(
        api,
        ProgressAPIProxy_Dispose,
        ProgressAPIProxy_OnProgress);
}

////////////// Longtail_AsyncPutStoredBlockAPI

struct AsyncPutStoredBlockAPIProxy
{
    struct Longtail_AsyncPutStoredBlockAPI m_API;
    void* m_Context;
};

static void* AsyncPutStoredBlockAPIProxy_GetContext(void* api) { return ((struct AsyncPutStoredBlockAPIProxy*)api)->m_Context; }
void AsyncPutStoredBlockAPIProxy_OnComplete(struct Longtail_AsyncPutStoredBlockAPI* async_complete_api, int err);
void AsyncPutStoredBlockAPIProxy_Dispose(struct Longtail_API* api);

static struct Longtail_AsyncPutStoredBlockAPI* CreateAsyncPutStoredBlockAPI(void* context)
{
    struct AsyncPutStoredBlockAPIProxy* api = (struct AsyncPutStoredBlockAPIProxy*)Longtail_Alloc("AsyncPutStoredBlockAPI", sizeof(struct AsyncPutStoredBlockAPIProxy));
    api->m_Context = context;
    return Longtail_MakeAsyncPutStoredBlockAPI(
        api,
        AsyncPutStoredBlockAPIProxy_Dispose,
        AsyncPutStoredBlockAPIProxy_OnComplete);
}

////////////// Longtail_AsyncGetStoredBlockAPI

struct AsyncGetStoredBlockAPIProxy
{
    struct Longtail_AsyncGetStoredBlockAPI m_API;
    void* m_Context;
};

static void* AsyncGetStoredBlockAPIProxy_GetContext(void* api) { return ((struct AsyncGetStoredBlockAPIProxy*)api)->m_Context; }
void AsyncGetStoredBlockAPIProxy_OnComplete(struct Longtail_AsyncGetStoredBlockAPI* async_complete_api, struct Longtail_StoredBlock* stored_block, int err);
void AsyncGetStoredBlockAPIProxy_Dispose(struct Longtail_API* api);

static struct Longtail_AsyncGetStoredBlockAPI* CreateAsyncGetStoredBlockAPI(void* context)
{
    struct AsyncGetStoredBlockAPIProxy* api = (struct AsyncGetStoredBlockAPIProxy*)Longtail_Alloc("AsyncGetStoredBlockAPI", sizeof(struct AsyncGetStoredBlockAPIProxy));
    api->m_Context = context;
    return Longtail_MakeAsyncGetStoredBlockAPI(
        api,
        AsyncGetStoredBlockAPIProxy_Dispose,
        AsyncGetStoredBlockAPIProxy_OnComplete);
}

////////////// Longtail_AsyncGetExistingContentAPI

struct AsyncGetExistingContentAPIProxy
{
    struct Longtail_AsyncGetExistingContentAPI m_API;
    void* m_Context;
};

static void* AsyncGetExistingContentAPIProxy_GetContext(void* api) { return ((struct AsyncGetExistingContentAPIProxy*)api)->m_Context; }
void AsyncGetExistingContentAPIProxy_OnComplete(struct Longtail_AsyncGetExistingContentAPI* async_complete_api, struct Longtail_ContentIndex* content_index, int err);
void AsyncGetExistingContentAPIProxy_Dispose(struct Longtail_API* api);

static struct Longtail_AsyncGetExistingContentAPI* CreateAsyncGetExistingContentAPI(void* context)
{
    struct AsyncGetExistingContentAPIProxy* api    = (struct AsyncGetExistingContentAPIProxy*)Longtail_Alloc("AsyncGetExistingContentAPI", sizeof(struct AsyncGetExistingContentAPIProxy));
    api->m_Context = context;
    return Longtail_MakeAsyncGetExistingContentAPI(
        api,
        AsyncGetExistingContentAPIProxy_Dispose,
        AsyncGetExistingContentAPIProxy_OnComplete);
}

////////////// Longtail_AsyncFlushAPI

struct AsyncFlushAPIProxy
{
    struct Longtail_AsyncFlushAPI m_API;
    void* m_Context;
};

static void* AsyncFlushAPIProxy_GetContext(void* api) { return ((struct AsyncFlushAPIProxy*)api)->m_Context; }
void AsyncFlushAPIProxy_OnComplete(struct Longtail_AsyncFlushAPI* async_complete_api, int err);
void AsyncFlushAPIProxy_Dispose(struct Longtail_API* api);

static struct Longtail_AsyncFlushAPI* CreateAsyncFlushAPI(void* context)
{
    struct AsyncFlushAPIProxy* api    = (struct AsyncFlushAPIProxy*)Longtail_Alloc("AsyncFlushAPI", sizeof(struct AsyncFlushAPIProxy));
    api->m_Context = context;
    return Longtail_MakeAsyncFlushAPI(
        api,
        AsyncFlushAPIProxy_Dispose,
        AsyncFlushAPIProxy_OnComplete);
}

static const char* GetVersionIndexPath(struct Longtail_VersionIndex* version_index, uint32_t asset_index)
{
    return &version_index->m_NameData[version_index->m_NameOffsets[asset_index]];
}

static uint64_t GetVersionAssetSize(struct Longtail_VersionIndex* version_index, uint32_t asset_index)
{
    return version_index->m_AssetSizes[asset_index];
}

static uint16_t GetVersionAssetPermissions(struct Longtail_VersionIndex* version_index, uint32_t asset_index)
{
    return version_index->m_Permissions[asset_index];
}

static void EnableMemtrace() {
	Longtail_MemTracer_Init();
	Longtail_SetAllocAndFree(Longtail_MemTracer_Alloc, Longtail_MemTracer_Free);
}

static void DisableMemtrace() {
	Longtail_SetAllocAndFree(0,  0);
	Longtail_MemTracer_Dispose();
}

#ifdef __cplusplus
}
#endif
