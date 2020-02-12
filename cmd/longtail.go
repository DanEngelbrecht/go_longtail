package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/DanEngelbrecht/golongtail/lib"
	"github.com/DanEngelbrecht/golongtail/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type loggerData struct {
}

func (l *loggerData) OnLog(level int, message string) {
	switch level {
	case 0:
		log.Printf("DEBUG: %s", message)
	case 1:
		log.Printf("INFO: %s", message)
	case 2:
		log.Printf("WARNING: %s", message)
	case 3:
		log.Fatal(message)
	}
}

type progressData struct {
	inited     bool
	oldPercent uint32
	task       string
}

func (p *progressData) OnProgress(totalCount uint32, doneCount uint32) {
	if doneCount < totalCount {
		if !p.inited {
			fmt.Fprintf(os.Stderr, "%s: ", p.task)
			p.inited = true
		}
		percentDone := (100 * doneCount) / totalCount
		if (percentDone - p.oldPercent) >= 5 {
			fmt.Fprintf(os.Stderr, "%d%% ", percentDone)
			p.oldPercent = percentDone
		}
		return
	}
	if p.inited {
		if p.oldPercent != 100 {
			fmt.Fprintf(os.Stderr, "100%%")
		}
		fmt.Fprintf(os.Stderr, " Done\n")
	}
}

func trace(s string) (string, time.Time) {
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	elapsed := time.Since(startTime)
	log.Printf("%s: elapsed %f secs\n", s, elapsed.Seconds())
}

func createBlobStoreForURI(uri string) (store.BlobStore, error) {
	blobStoreURL, err := url.Parse(*storageURI)
	if err == nil {
		switch blobStoreURL.Scheme {
		case "gs":
			return store.NewGCSBlobStore(blobStoreURL)
		case "s3":
			return nil, fmt.Errorf("AWS storage not yet implemented")
		case "abfs":
			return nil, fmt.Errorf("Azure Gen1 storage not yet implemented")
		case "abfss":
			return nil, fmt.Errorf("Azure Gen2 storage not yet implemented")
		case "file":
			return store.NewFSBlobStore(blobStoreURL.Path[1:])
		}
	}
	return store.NewFSBlobStore(uri)
}

func createBlockStoreForURI(uri string, hashIdentifier uint32) (lib.Longtail_BlockStoreAPI, error) {
	blobStoreURL, err := url.Parse(*storageURI)
	if err == nil {
		switch blobStoreURL.Scheme {
		case "gs":
			return store.NewGCSBlockStore(blobStoreURL, hashIdentifier)
		case "s3":
			return lib.Longtail_BlockStoreAPI{}, fmt.Errorf("AWS storage not yet implemented")
		case "abfs":
			return lib.Longtail_BlockStoreAPI{}, fmt.Errorf("Azure Gen1 storage not yet implemented")
		case "abfss":
			return lib.Longtail_BlockStoreAPI{}, fmt.Errorf("Azure Gen2 storage not yet implemented")
		case "file":
			return lib.CreateFSBlockStore(lib.CreateFSStorageAPI(), blobStoreURL.Path[1:]), nil
		}
	}
	return lib.CreateFSBlockStore(lib.CreateFSStorageAPI(), blobStoreURL.Path[1:]), nil
}

const noCompressionType = uint32(0)

func getCompressionType(compressionAlgorithm *string) (uint32, error) {
	switch *compressionAlgorithm {
	case "none":
		return noCompressionType, nil
	case "brotli":
		return lib.GetBrotliGenericDefaultCompressionType(), nil
	case "brotli_min":
		return lib.GetBrotliGenericMinCompressionType(), nil
	case "brotli_max":
		return lib.GetBrotliGenericMaxCompressionType(), nil
	case "brotli_text":
		return lib.GetBrotliTextDefaultCompressionType(), nil
	case "brotli_text_min":
		return lib.GetBrotliTextMinCompressionType(), nil
	case "brotli_text_max":
		return lib.GetBrotliTextMaxCompressionType(), nil
	case "lz4":
		return lib.GetLZ4DefaultCompressionType(), nil
	case "zstd":
		return lib.GetZStdMaxCompressionType(), nil
	case "zstd_min":
		return lib.GetZStdMinCompressionType(), nil
	case "zstd_max":
		return lib.GetZStdMaxCompressionType(), nil
	}
	return 0, fmt.Errorf("Unsupported compression algorithm: `%s`", *compressionAlgorithm)
}

func getCompressionTypesForFiles(fileInfos lib.Longtail_FileInfos, compressionType uint32) []uint32 {
	pathCount := fileInfos.GetFileCount()
	compressionTypes := make([]uint32, pathCount)
	for i := uint32(0); i < pathCount; i++ {
		compressionTypes[i] = compressionType
	}
	return compressionTypes
}

func createHashAPIFromIdentifier(hashIdentifier uint32) (lib.Longtail_HashAPI, error) {
	if hashIdentifier == lib.GetMeowHashIdentifier() {
		return lib.CreateMeowHashAPI(), nil
	}
	if hashIdentifier == lib.GetBlake2HashIdentifier() {
		return lib.CreateBlake2HashAPI(), nil
	}
	if hashIdentifier == lib.GetBlake3HashIdentifier() {
		return lib.CreateBlake3HashAPI(), nil
	}
	return lib.Longtail_HashAPI{}, fmt.Errorf("not a supported hash identifier: `%d`", hashIdentifier)
}

func getHashIdentifier(hashAlgorithm *string) (uint32, error) {
	switch *hashAlgorithm {
	case "meow":
		return lib.GetMeowHashIdentifier(), nil
	case "blake2":
		return lib.GetBlake2HashIdentifier(), nil
	case "blake3":
		return lib.GetBlake3HashIdentifier(), nil
	}
	return 0, fmt.Errorf("not a supportd hash api: `%s`", *hashAlgorithm)
}

func createHashAPI(hashAlgorithm *string) (lib.Longtail_HashAPI, error) {
	hashIdentifier, err := getHashIdentifier(hashAlgorithm)
	if err != nil {
		return lib.Longtail_HashAPI{}, err
	}
	return createHashAPIFromIdentifier(hashIdentifier)
}

func upSyncVersion(
	blobStoreURI string,
	sourceFolderPath string,
	sourceIndexPath *string,
	targetFilePath string,
	targetChunkSize uint32,
	targetBlockSize uint32,
	maxChunksPerBlock uint32,
	compressionAlgorithm *string,
	hashAlgorithm *string) error {
	fs := lib.CreateFSStorageAPI()
	defer fs.Dispose()
	jobs := lib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()))
	defer jobs.Dispose()
	creg := lib.CreateDefaultCompressionRegistry()
	defer creg.Dispose()

	hashIdentifier, err := getHashIdentifier(hashAlgorithm)
	if err != nil {
		return err
	}

	indexStore, err := createBlockStoreForURI(blobStoreURI, hashIdentifier)
	if err != nil {
		return err
	}
	defer indexStore.Dispose()

	getRemoteIndexProgress := lib.CreateProgressAPI(&progressData{task: "Get remote index"})
	defer getRemoteIndexProgress.Dispose()
	remoteContentIndex, err := indexStore.GetIndex(hashIdentifier, jobs, &getRemoteIndexProgress)
	if err != nil {
		return err
	}

	hash, err := createHashAPIFromIdentifier(remoteContentIndex.GetHashAPI())
	if err != nil {
		return err
	}
	defer hash.Dispose()

	var vindex lib.Longtail_VersionIndex
	if sourceIndexPath == nil || len(*sourceIndexPath) == 0 {
		fileInfos, err := lib.GetFilesRecursively(fs, sourceFolderPath)
		if err != nil {
			return err
		}
		defer fileInfos.Dispose()

		compressionType, err := getCompressionType(compressionAlgorithm)
		if err != nil {
			return err
		}
		compressionTypes := getCompressionTypesForFiles(fileInfos, compressionType)

		createVersionIndexProgress := lib.CreateProgressAPI(&progressData{task: "Indexing version"})
		defer createVersionIndexProgress.Dispose()
		vindex, err = lib.CreateVersionIndex(
			fs,
			hash,
			jobs,
			&createVersionIndexProgress,
			sourceFolderPath,
			fileInfos.GetPaths(),
			fileInfos.GetFileSizes(),
			fileInfos.GetFilePermissions(),
			compressionTypes,
			targetChunkSize)
		if err != nil {
			return err
		}
	} else {
		vindex, err = lib.ReadVersionIndex(fs, *sourceIndexPath)
		if err != nil {
			return err
		}
	}
	defer vindex.Dispose()

	//	versionBlob, err := lib.WriteVersionIndexToBuffer(vindex)
	//	if err != nil {
	//		return err
	//	}

	missingContentIndex, err := lib.CreateMissingContent(
		hash,
		remoteContentIndex,
		vindex,
		targetBlockSize,
		maxChunksPerBlock)
	if err != nil {
		return err
	}
	defer missingContentIndex.Dispose()
	if missingContentIndex.GetBlockCount() > 0 {
		writeContentProgress := lib.CreateProgressAPI(&progressData{task: "Writing content blocks"})
		defer writeContentProgress.Dispose()
		err = lib.WriteContent(
			fs,
			indexStore,
			creg,
			jobs,
			&writeContentProgress,
			missingContentIndex,
			vindex,
			sourceFolderPath)
		if err != nil {
			return err
		}
	}

	localFS := lib.CreateFSStorageAPI()
	defer localFS.Dispose()

	err = lib.WriteVersionIndex(localFS, vindex, targetFilePath)
	if err != nil {
		return err
	}

	//	err = indexStore.PutBlob(
	//		context.Background(),
	//		targetFilePath,
	//		"application/octet-stream",
	//		versionBlob)
	//	if err != nil {
	//		return err
	//	}

	return nil
}

func downSyncVersion(
	blobStoreURI string,
	sourceFilePath string,
	targetFolderPath string,
	targetIndexPath *string,
	localCachePath string,
	targetChunkSize uint32,
	targetBlockSize uint32,
	maxChunksPerBlock uint32,
	hashAlgorithm *string,
	retainPermissions bool) error {
	//	defer un(trace("downSyncVersion " + sourceFilePath))
	fs := lib.CreateFSStorageAPI()
	defer fs.Dispose()
	jobs := lib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()))
	defer jobs.Dispose()
	creg := lib.CreateDefaultCompressionRegistry()
	defer creg.Dispose()

	hashIdentifier, err := getHashIdentifier(hashAlgorithm)
	if err != nil {
		return err
	}

	remoteIndexStore, err := createBlockStoreForURI(blobStoreURI, hashIdentifier)
	if err != nil {
		return err
	}
	defer remoteIndexStore.Dispose()

	localFS := lib.CreateFSStorageAPI()
	defer localFS.Dispose()

	localIndexStore := lib.CreateFSBlockStore(localFS, localCachePath)
	if err != nil {
		return err
	}
	defer localIndexStore.Dispose()

	indexStore := lib.CreateCacheBlockStore(localIndexStore, remoteIndexStore)
	defer indexStore.Dispose()

	getRemoteIndexProgress := lib.CreateProgressAPI(&progressData{task: "Get remote index"})
	defer getRemoteIndexProgress.Dispose()
	remoteContentIndex, err := indexStore.GetIndex(hashIdentifier, jobs, &getRemoteIndexProgress)
	if err != nil {
		return err
	}

	hash, err := createHashAPIFromIdentifier(remoteContentIndex.GetHashAPI())
	if err != nil {
		return err
	}
	defer hash.Dispose()

	var remoteVersionIndex lib.Longtail_VersionIndex

	//	remoteVersionBlob, err := indexStore.GetBlob(context.Background(), sourceFilePath)
	//	if err != nil {
	//		return err
	//	}
	//	remoteVersionIndex, err = lib.ReadVersionIndexFromBuffer(remoteVersionBlob)
	remoteVersionIndex, err = lib.ReadVersionIndex(localFS, sourceFilePath)
	if err != nil {
		return err
	}

	var localVersionIndex lib.Longtail_VersionIndex
	if targetIndexPath == nil || len(*targetIndexPath) == 0 {
		fileInfos, err := lib.GetFilesRecursively(fs, targetFolderPath)
		if err != nil {
			return err
		}
		defer fileInfos.Dispose()

		compressionTypes := getCompressionTypesForFiles(fileInfos, noCompressionType)

		createVersionIndexProgress := lib.CreateProgressAPI(&progressData{task: "Indexing version"})
		defer createVersionIndexProgress.Dispose()
		localVersionIndex, err = lib.CreateVersionIndex(
			fs,
			hash,
			jobs,
			&createVersionIndexProgress,
			targetFolderPath,
			fileInfos.GetPaths(),
			fileInfos.GetFileSizes(),
			fileInfos.GetFilePermissions(),
			compressionTypes,
			targetChunkSize)
		if err != nil {
			return err
		}
	} else {
		localVersionIndex, err = lib.ReadVersionIndex(fs, *targetIndexPath)
		if err != nil {
			return err
		}
	}
	defer localVersionIndex.Dispose()

	versionDiff, err := lib.CreateVersionDiff(localVersionIndex, remoteVersionIndex)
	if err != nil {
		return err
	}
	defer versionDiff.Dispose()

	changeVersionProgress := lib.CreateProgressAPI(&progressData{task: "Updating version"})
	defer changeVersionProgress.Dispose()
	err = lib.ChangeVersion(
		indexStore,
		fs,
		hash,
		jobs,
		&changeVersionProgress,
		creg,
		remoteContentIndex,
		localVersionIndex,
		remoteVersionIndex,
		versionDiff,
		targetFolderPath,
		retainPermissions)
	if err != nil {
		return err
	}
	return nil
}

func parseLevel(lvl string) (int, error) {
	switch strings.ToLower(lvl) {
	case "debug":
		return 0, nil
	case "info":
		return 1, nil
	case "warn":
		return 2, nil
	case "error":
		return 3, nil
	case "off":
		return 4, nil
	}

	return -1, fmt.Errorf("not a valid log Level: %q", lvl)
}

var (
	logLevel          = kingpin.Flag("log-level", "Log level").Default("warn").Enum("debug", "info", "warn", "error")
	targetChunkSize   = kingpin.Flag("target-chunk-size", "Target chunk size").Default("32768").Uint32()
	targetBlockSize   = kingpin.Flag("target-block-size", "Target block size").Default("524288").Uint32()
	maxChunksPerBlock = kingpin.Flag("max-chunks-per-block", "Max chunks per block").Default("1024").Uint32()
	storageURI        = kingpin.Flag("storage-uri", "Storage URI (only GCS bucket URI supported)").String()
	hashing           = kingpin.Flag("hash-algorithm", "Hashing algorithm: blake2, blake3, meow").
				Default("blake3").
				Enum("meow", "blake2", "blake3")

	commandUpSync    = kingpin.Command("upsync", "Upload a folder")
	sourceFolderPath = commandUpSync.Flag("source-path", "Source folder path").String()
	sourceIndexPath  = commandUpSync.Flag("source-index-path", "Optional pre-computed index of source-path").String()
	targetFilePath   = commandUpSync.Flag("target-path", "Target file path relative to --storage-uri").String()
	compression      = commandUpSync.Flag("compression-algorithm", "Compression algorithm: none, brotli[_min|_max], brotli_text[_min|_max], lz4, ztd[_min|_max]").
				Default("zstd").
				Enum(
			"none",
			"brotli",
			"brotli_min",
			"brotli_max",
			"brotli_text",
			"brotli_text_min",
			"brotli_text_max",
			"lz4",
			"zstd",
			"zstd_min",
			"zstd_max")

	commandDownSync     = kingpin.Command("downsync", "Download a folder")
	downSyncContentPath = commandDownSync.Flag("content-path", "Location for downloaded/cached blocks").Default(path.Join(os.TempDir(), "longtail_block_store")).String()
	targetFolderPath    = commandDownSync.Flag("target-path", "Target folder path").String()
	targetIndexPath     = commandUpSync.Flag("target-index-path", "Optional pre-computed index of target-path").String()
	sourceFilePath      = commandDownSync.Flag("source-path", "Source file path relative to --storage-uri").String()
	noRetainPermissions = commandDownSync.Flag("no-retain-permissions", "Disable setting permission on file/directories from source").Bool()
)

type assertData struct {
}

func (a *assertData) OnAssert(expression string, file string, line int) {
	log.Fatalf("ASSERT: %s %s:%d", expression, file, line)
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.DefaultEnvars()
	kingpin.Parse()

	longtailLogLevel, err := parseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}

	lib.SetLogger(&loggerData{})
	defer lib.SetLogger(nil)
	lib.SetLogLevel(longtailLogLevel)

	lib.SetAssert(&assertData{})
	defer lib.SetAssert(nil)

	switch kingpin.Parse() {
	case commandUpSync.FullCommand():
		err := upSyncVersion(
			*storageURI,
			*sourceFolderPath,
			sourceIndexPath,
			*targetFilePath,
			*targetChunkSize,
			*targetBlockSize,
			*maxChunksPerBlock,
			compression, hashing)
		if err != nil {
			log.Fatal(err)
		}
	case commandDownSync.FullCommand():
		err := downSyncVersion(
			*storageURI,
			*sourceFilePath,
			*targetFolderPath,
			targetIndexPath,
			*downSyncContentPath,
			*targetChunkSize,
			*targetBlockSize,
			*maxChunksPerBlock,
			hashing,
			!(*noRetainPermissions))
		if err != nil {
			log.Fatal(err)
		}
	}
}
