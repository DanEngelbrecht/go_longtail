package longtaillib

// #cgo CFLAGS: -g -std=gnu99 -m64 -msse4.1 -maes -pthread -O3
// #include "golongtail.h"
import "C"
import (
	"errors"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const EPERM = 1       /* Not super-user */
const ENOENT = 2      /* No such file or directory */
const ESRCH = 3       /* No such process */
const EINTR = 4       /* Interrupted system call */
const EIO = 5         /* I/O error */
const ENXIO = 6       /* No such device or address */
const E2BIG = 7       /* Arg list too long */
const ENOEXEC = 8     /* Exec format error */
const EBADF = 9       /* Bad file number */
const ECHILD = 10     /* No children */
const EAGAIN = 11     /* No more processes */
const ENOMEM = 12     /* Not enough core */
const EACCES = 13     /* Permission denied */
const EFAULT = 14     /* Bad address */
const ENOTBLK = 15    /* Block device required */
const EBUSY = 16      /* Mount device busy */
const EEXIST = 17     /* File exists */
const EXDEV = 18      /* Cross-device link */
const ENODEV = 19     /* No such device */
const ENOTDIR = 20    /* Not a directory */
const EISDIR = 21     /* Is a directory */
const EINVAL = 22     /* Invalid argument */
const ENFILE = 23     /* Too many open files in system */
const EMFILE = 24     /* Too many open files */
const ENOTTY = 25     /* Not a typewriter */
const ETXTBSY = 26    /* Text file busy */
const EFBIG = 27      /* File too large */
const ENOSPC = 28     /* No space left on device */
const ESPIPE = 29     /* Illegal seek */
const EROFS = 30      /* Read only file system */
const EMLINK = 31     /* Too many links */
const EPIPE = 32      /* Broken pipe */
const EDOM = 33       /* Math arg out of domain of func */
const ERANGE = 34     /* Math result not representable */
const ECANCELED = 105 /* Operation canceled */

var (
	//ErrEPERM Not super-user
	ErrEPERM = errors.New("Not super-user")
	//ErrENOENT No such file or directory
	ErrENOENT = errors.New("No such file or directory")
	//ErrESRCH No such process
	ErrESRCH = errors.New("No such process")
	//ErrEINTR Interrupted system call
	ErrEINTR = errors.New("Interrupted system call")
	//ErrEIO I/O error
	ErrEIO = errors.New("I/O error")
	//ErrENXIO No such device or address
	ErrENXIO = errors.New("No such device or address")
	//ErrE2BIG Arg list too long
	ErrE2BIG = errors.New("Arg list too long")
	//ErrENOEXEC Exec format error
	ErrENOEXEC = errors.New("Exec format error")
	//ErrEBADF Bad file number
	ErrEBADF = errors.New("Bad file number")
	//ErrECHILD No children
	ErrECHILD = errors.New("No children")
	//ErrEAGAIN No more processes
	ErrEAGAIN = errors.New("No more processes")
	//ErrENOMEM Not enough core
	ErrENOMEM = errors.New("Not enough core")
	//ErrEACCES Permission denied
	ErrEACCES = errors.New("Permission denied")
	//ErrEFAULT Bad address
	ErrEFAULT = errors.New("Bad address")
	//ErrENOTBLK Block device required
	ErrENOTBLK = errors.New("Block device required")
	//ErrEBUSY Mount device busy
	ErrEBUSY = errors.New("Mount device busy")
	//ErrEEXIST File exists
	ErrEEXIST = errors.New("File exists")
	//ErrEXDEV Cross-device link
	ErrEXDEV = errors.New("Cross-device link")
	//ErrENODEV No such device
	ErrENODEV = errors.New("No such device")
	//ErrENOTDIR Not a directory
	ErrENOTDIR = errors.New("Not a directory")
	//ErrEISDIR Is a directory
	ErrEISDIR = errors.New("Is a directory")
	//ErrEINVAL Invalid argument
	ErrEINVAL = errors.New("Invalid argument")
	//ErrENFILE Too many open files in system
	ErrENFILE = errors.New("Too many open files in system")
	//ErrEMFILE Too many open files
	ErrEMFILE = errors.New("Too many open files")
	//ErrENOTTY Not a typewriter
	ErrENOTTY = errors.New("Not a typewriter")
	//ErrETXTBSY Text file busy
	ErrETXTBSY = errors.New("Text file busy")
	//ErrEFBIG File too large
	ErrEFBIG = errors.New("File too large")
	//ErrENOSPC No space left on device
	ErrENOSPC = errors.New("No space left on device")
	//ErrESPIPE Illegal seek
	ErrESPIPE = errors.New("Illegal seek")
	//ErrEROFS Read only file system
	ErrEROFS = errors.New("Read only file system")
	//ErrEMLINK Too many links
	ErrEMLINK = errors.New("Too many links")
	//ErrEPIPE Broken pipe
	ErrEPIPE = errors.New("Broken pipe")
	//ErrEDOM Math arg out of domain of func
	ErrEDOM = errors.New("Math arg out of domain of func")
	//ErrERANGE Math result not representable
	ErrERANGE = errors.New("Math result not representable")
	//ErrECANCELED Operation canceled
	ErrECANCELED = errors.New("Operation canceled")
)

var errnoToError = map[int]error{
	EPERM:     ErrEPERM,
	ENOENT:    ErrENOENT,
	ESRCH:     ErrESRCH,
	EINTR:     ErrEINTR,
	EIO:       ErrEIO,
	ENXIO:     ErrENXIO,
	E2BIG:     ErrE2BIG,
	ENOEXEC:   ErrENOEXEC,
	EBADF:     ErrEBADF,
	ECHILD:    ErrECHILD,
	EAGAIN:    ErrEAGAIN,
	ENOMEM:    ErrENOMEM,
	EACCES:    ErrEACCES,
	EFAULT:    ErrEFAULT,
	ENOTBLK:   ErrENOTBLK,
	EBUSY:     ErrEBUSY,
	EEXIST:    ErrEEXIST,
	EXDEV:     ErrEXDEV,
	ENODEV:    ErrENODEV,
	ENOTDIR:   ErrENOTDIR,
	EISDIR:    ErrEISDIR,
	EINVAL:    ErrEINVAL,
	ENFILE:    ErrENFILE,
	EMFILE:    ErrEMFILE,
	ENOTTY:    ErrENOTTY,
	ETXTBSY:   ErrETXTBSY,
	EFBIG:     ErrEFBIG,
	ENOSPC:    ErrENOSPC,
	ESPIPE:    ErrESPIPE,
	EROFS:     ErrEROFS,
	EMLINK:    ErrEMLINK,
	EPIPE:     ErrEPIPE,
	EDOM:      ErrEDOM,
	ERANGE:    ErrERANGE,
	ECANCELED: ErrECANCELED,
}

var errorToErrno = map[error]int{
	ErrEPERM:     EPERM,
	ErrENOENT:    ENOENT,
	ErrESRCH:     ESRCH,
	ErrEINTR:     EINTR,
	ErrEIO:       EIO,
	ErrENXIO:     ENXIO,
	ErrE2BIG:     E2BIG,
	ErrENOEXEC:   ENOEXEC,
	ErrEBADF:     EBADF,
	ErrECHILD:    ECHILD,
	ErrEAGAIN:    EAGAIN,
	ErrENOMEM:    ENOMEM,
	ErrEACCES:    EACCES,
	ErrEFAULT:    EFAULT,
	ErrENOTBLK:   ENOTBLK,
	ErrEBUSY:     EBUSY,
	ErrEEXIST:    EEXIST,
	ErrEXDEV:     EXDEV,
	ErrENODEV:    ENODEV,
	ErrENOTDIR:   ENOTDIR,
	ErrEISDIR:    EISDIR,
	ErrEINVAL:    EINVAL,
	ErrENFILE:    ENFILE,
	ErrEMFILE:    EMFILE,
	ErrENOTTY:    ENOTTY,
	ErrETXTBSY:   ETXTBSY,
	ErrEFBIG:     EFBIG,
	ErrENOSPC:    ENOSPC,
	ErrESPIPE:    ESPIPE,
	ErrEROFS:     EROFS,
	ErrEMLINK:    EMLINK,
	ErrEPIPE:     EPIPE,
	ErrEDOM:      EDOM,
	ErrERANGE:    ERANGE,
	ErrECANCELED: ECANCELED,
}

// ErrnoToError Converts a longtail int errno to golang error
func ErrnoToError(errno int, fallback error) error {
	if errno == 0 {
		return nil
	}
	error, exists := errnoToError[errno]
	if exists {
		return error
	}
	return fallback //errors.New(fmt.Sprintf("Error: %d", errno))
}

// ErrorToErrno Converts a golang error to a longtail int errno
func ErrorToErrno(err error, fallback int) int {
	if err == nil {
		return 0
	}
	unwrappedError := errors.Unwrap(err)
	if unwrappedError == nil {
		unwrappedError = err
	}
	errno, exists := errorToErrno[unwrappedError]
	if exists {
		return errno
	}
	return fallback //ENOENT // Bad catchall
}

type ProgressAPI interface {
	OnProgress(totalCount uint32, doneCount uint32)
}
type PathFilterAPI interface {
	Include(rootPath string, assetPath string, assetName string, isDir bool, size uint64, permissions uint16) bool
}

type AsyncPutStoredBlockAPI interface {
	OnComplete(errno int)
}

type AsyncGetStoredBlockAPI interface {
	OnComplete(stored_block Longtail_StoredBlock, errno int)
}

type AsyncGetExistingContentAPI interface {
	OnComplete(content_index Longtail_ContentIndex, errno int)
}

type AsyncFlushAPI interface {
	OnComplete(errno int)
}

type Assert interface {
	OnAssert(expression string, file string, line int)
}

type LogField struct {
	Name  string
	Value string
}

type Logger interface {
	OnLog(file string, function string, line int, level int, logFields []LogField, message string)
}

type Longtail_AsyncGetStoredBlockAPI struct {
	cAsyncCompleteAPI *C.struct_Longtail_AsyncGetStoredBlockAPI
}

type Longtail_AsyncPutStoredBlockAPI struct {
	cAsyncCompleteAPI *C.struct_Longtail_AsyncPutStoredBlockAPI
}

type Longtail_AsyncGetExistingContentAPI struct {
	cAsyncCompleteAPI *C.struct_Longtail_AsyncGetExistingContentAPI
}

type Longtail_AsyncFlushAPI struct {
	cAsyncCompleteAPI *C.struct_Longtail_AsyncFlushAPI
}

const (
	Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Count       = 0
	Longtail_BlockStoreAPI_StatU64_GetStoredBlock_RetryCount  = 1
	Longtail_BlockStoreAPI_StatU64_GetStoredBlock_FailCount   = 2
	Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Chunk_Count = 3
	Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Byte_Count  = 4

	Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Count       = 5
	Longtail_BlockStoreAPI_StatU64_PutStoredBlock_RetryCount  = 6
	Longtail_BlockStoreAPI_StatU64_PutStoredBlock_FailCount   = 7
	Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Chunk_Count = 8
	Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Byte_Count  = 9

	Longtail_BlockStoreAPI_StatU64_GetExistingContent_Count      = 10
	Longtail_BlockStoreAPI_StatU64_GetExistingContent_RetryCount = 11
	Longtail_BlockStoreAPI_StatU64_GetExistingContent_FailCount  = 12

	Longtail_BlockStoreAPI_StatU64_PreflightGet_Count      = 13
	Longtail_BlockStoreAPI_StatU64_PreflightGet_RetryCount = 14
	Longtail_BlockStoreAPI_StatU64_PreflightGet_FailCount  = 15

	Longtail_BlockStoreAPI_StatU64_Flush_Count     = 16
	Longtail_BlockStoreAPI_StatU64_Flush_FailCount = 17

	Longtail_BlockStoreAPI_StatU64_GetStats_Count = 18
	Longtail_BlockStoreAPI_StatU64_Count          = 19
)

type BlockStoreStats struct {
	StatU64 [Longtail_BlockStoreAPI_StatU64_Count]uint64
}

type BlockStoreAPI interface {
	PutStoredBlock(storedBlock Longtail_StoredBlock, asyncCompleteAPI Longtail_AsyncPutStoredBlockAPI) int
	PreflightGet(chunkHashes []uint64) int
	GetStoredBlock(blockHash uint64, asyncCompleteAPI Longtail_AsyncGetStoredBlockAPI) int
	GetExistingContent(chunkHashes []uint64, minBlockUsagePercent uint32, asyncCompleteAPI Longtail_AsyncGetExistingContentAPI) int
	GetStats() (BlockStoreStats, int)
	Flush(asyncCompleteAPI Longtail_AsyncFlushAPI) int
	Close()
}

type Longtail_FileInfos struct {
	cFileInfos *C.struct_Longtail_FileInfos
}

type Longtail_ContentIndex struct {
	cContentIndex *C.struct_Longtail_ContentIndex
}

type Longtail_StoreIndex struct {
	cStoreIndex *C.struct_Longtail_StoreIndex
}

type Longtail_VersionIndex struct {
	cVersionIndex *C.struct_Longtail_VersionIndex
}

type Longtail_VersionDiff struct {
	cVersionDiff *C.struct_Longtail_VersionDiff
}

type Longtail_ProgressAPI struct {
	cProgressAPI *C.struct_Longtail_ProgressAPI
}

type Longtail_PathFilterAPI struct {
	cPathFilterAPI *C.struct_Longtail_PathFilterAPI
}

type Longtail_JobAPI struct {
	cJobAPI *C.struct_Longtail_JobAPI
}

type Longtail_CompressionRegistryAPI struct {
	cCompressionRegistryAPI *C.struct_Longtail_CompressionRegistryAPI
}

type Longtail_HashRegistryAPI struct {
	cHashRegistryAPI *C.struct_Longtail_HashRegistryAPI
}

type Longtail_CompressionAPI struct {
	cCompressionAPI *C.struct_Longtail_CompressionAPI
}

type Longtail_BlockStoreAPI struct {
	cBlockStoreAPI *C.struct_Longtail_BlockStoreAPI
}

type Longtail_StoredBlock struct {
	cStoredBlock *C.struct_Longtail_StoredBlock
}

type Longtail_BlockIndex struct {
	cBlockIndex *C.struct_Longtail_BlockIndex
}

type Longtail_StorageAPI struct {
	cStorageAPI *C.struct_Longtail_StorageAPI
}

type Longtail_StorageAPI_HOpenFile struct {
	cOpenFile C.Longtail_StorageAPI_HOpenFile
}

type Longtail_StorageAPI_Iterator struct {
	cIterator C.Longtail_StorageAPI_HIterator
}

type Longtail_StorageAPI_EntryProperties struct {
	Name        string
	Size        uint64
	Permissions uint16
	IsDir       bool
}

type Longtail_HashAPI struct {
	cHashAPI *C.struct_Longtail_HashAPI
}

type Longtail_ChunkerAPI struct {
	cChunkerAPI *C.struct_Longtail_ChunkerAPI
}

var pointerIndex uint32
var pointerStore [1024]interface{}
var pointerIndexer = (*[1 << 30]C.uint32_t)(C.malloc(4 * 1024))

func SavePointer(v interface{}) unsafe.Pointer {
	if v == nil {
		return nil
	}

	newPointerIndex := (atomic.AddUint32(&pointerIndex, 1)) % 1024
	startPointerIndex := newPointerIndex
	for pointerStore[newPointerIndex] != nil {
		newPointerIndex = (atomic.AddUint32(&pointerIndex, 1)) % 1024
		if newPointerIndex == startPointerIndex {
			return nil
		}
	}
	pointerIndexer[newPointerIndex] = C.uint32_t(newPointerIndex)
	pointerStore[newPointerIndex] = v
	return unsafe.Pointer(&pointerIndexer[newPointerIndex])
}

func RestorePointer(ptr unsafe.Pointer) (v interface{}) {
	if ptr == nil {
		return nil
	}

	p := (*C.uint32_t)(ptr)
	index := uint32(*p)

	return pointerStore[index]
}

func UnrefPointer(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	p := (*C.uint32_t)(ptr)
	index := uint32(*p)
	pointerStore[index] = nil
}

// ReadFromStorage ...
func (storageAPI *Longtail_StorageAPI) ReadFromStorage(rootPath string, path string) ([]byte, int) {
	cRootPath := C.CString(rootPath)
	defer C.free(unsafe.Pointer(cRootPath))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cFullPath := C.Longtail_Storage_ConcatPath(storageAPI.cStorageAPI, cRootPath, cPath)
	defer C.Longtail_Free(unsafe.Pointer(cFullPath))

	var cOpenFile C.Longtail_StorageAPI_HOpenFile
	errno := C.Longtail_Storage_OpenReadFile(storageAPI.cStorageAPI, cFullPath, &cOpenFile)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Storage_CloseFile(storageAPI.cStorageAPI, cOpenFile)

	var cSize C.uint64_t
	errno = C.Longtail_Storage_GetSize(storageAPI.cStorageAPI, cOpenFile, &cSize)
	if errno != 0 {
		return nil, int(errno)
	}

	blockData := make([]byte, int(cSize))
	if cSize > 0 {
		errno = C.Longtail_Storage_Read(storageAPI.cStorageAPI, cOpenFile, 0, cSize, unsafe.Pointer(&blockData[0]))
		if errno != 0 {
			return nil, int(errno)
		}
	}
	return blockData, 0
}

// WriteToStorage ...
func (storageAPI *Longtail_StorageAPI) WriteToStorage(rootPath string, path string, blockData []byte) int {
	cRootPath := C.CString(rootPath)
	defer C.free(unsafe.Pointer(cRootPath))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cFullPath := C.Longtail_Storage_ConcatPath(storageAPI.cStorageAPI, cRootPath, cPath)
	defer C.Longtail_Free(unsafe.Pointer(cFullPath))

	errno := C.EnsureParentPathExists(storageAPI.cStorageAPI, cFullPath)
	if errno != 0 {
		return int(errno)
	}

	blockSize := C.uint64_t(len(blockData))

	var cOpenFile C.Longtail_StorageAPI_HOpenFile
	errno = C.Longtail_Storage_OpenWriteFile(storageAPI.cStorageAPI, cFullPath, blockSize, &cOpenFile)
	if errno != 0 {
		return int(errno)
	}
	defer C.Longtail_Storage_CloseFile(storageAPI.cStorageAPI, cOpenFile)

	if blockSize > 0 {
		data := unsafe.Pointer(&blockData[0])
		errno = C.Longtail_Storage_Write(storageAPI.cStorageAPI, cOpenFile, 0, blockSize, data)
	}

	if errno != 0 {
		return int(errno)
	}
	return 0
}

func (storageAPI *Longtail_StorageAPI) OpenReadFile(path string) (Longtail_StorageAPI_HOpenFile, int) {
	cPath := C.CString(path)
	var cOpenFile C.Longtail_StorageAPI_HOpenFile
	errno := C.Longtail_Storage_OpenReadFile(storageAPI.cStorageAPI, cPath, &cOpenFile)
	if errno != 0 {
		return Longtail_StorageAPI_HOpenFile{}, int(errno)
	}
	return Longtail_StorageAPI_HOpenFile{cOpenFile: cOpenFile}, 0
}

func (storageAPI *Longtail_StorageAPI) GetSize(f Longtail_StorageAPI_HOpenFile) (uint64, int) {
	var size C.uint64_t
	errno := C.Longtail_Storage_GetSize(storageAPI.cStorageAPI, f.cOpenFile, &size)
	if errno != 0 {
		return 0, int(errno)
	}
	return uint64(size), 0
}

func (storageAPI *Longtail_StorageAPI) Read(f Longtail_StorageAPI_HOpenFile, offset uint64, size uint64) ([]byte, int) {

	blockData := make([]byte, size)
	errno := C.Longtail_Storage_Read(storageAPI.cStorageAPI, f.cOpenFile, C.uint64_t(offset), C.uint64_t(size), unsafe.Pointer(&blockData[0]))
	if errno != 0 {
		return nil, int(errno)
	}
	return blockData, 0
}

func (storageAPI *Longtail_StorageAPI) CloseFile(f Longtail_StorageAPI_HOpenFile) {
	C.Longtail_Storage_CloseFile(storageAPI.cStorageAPI, f.cOpenFile)
}

func (storageAPI *Longtail_StorageAPI) StartFind(path string) (Longtail_StorageAPI_Iterator, int) {
	cPath := C.CString(path)
	var cIterator C.Longtail_StorageAPI_HIterator
	errno := C.Longtail_Storage_StartFind(storageAPI.cStorageAPI, cPath, &cIterator)
	if errno != 0 {
		return Longtail_StorageAPI_Iterator{}, int(errno)
	}
	return Longtail_StorageAPI_Iterator{cIterator: cIterator}, 0
}

func (storageAPI *Longtail_StorageAPI) FindNext(iterator Longtail_StorageAPI_Iterator) int {
	errno := C.Longtail_Storage_FindNext(storageAPI.cStorageAPI, iterator.cIterator)
	return int(errno)
}

func (storageAPI *Longtail_StorageAPI) CloseFind(iterator Longtail_StorageAPI_Iterator) {
	C.Longtail_Storage_CloseFind(storageAPI.cStorageAPI, iterator.cIterator)
	iterator.cIterator = nil
}

func (storageAPI *Longtail_StorageAPI) GetEntryProperties(iterator Longtail_StorageAPI_Iterator) (Longtail_StorageAPI_EntryProperties, int) {
	var cProperties C.struct_Longtail_StorageAPI_EntryProperties
	errno := C.Longtail_Storage_GetEntryProperties(storageAPI.cStorageAPI, iterator.cIterator, &cProperties)
	if errno != 0 {
		return Longtail_StorageAPI_EntryProperties{}, int(errno)
	}
	return Longtail_StorageAPI_EntryProperties{
		Name:        C.GoString(cProperties.m_Name),
		Size:        uint64(cProperties.m_Size),
		Permissions: uint16(cProperties.m_Permissions),
		IsDir:       cProperties.m_IsDir != 0}, 0
}

func (fileInfos *Longtail_FileInfos) Dispose() {
	if fileInfos.cFileInfos != nil {
		C.Longtail_Free(unsafe.Pointer(fileInfos.cFileInfos))
		fileInfos.cFileInfos = nil
	}
}

func carray2slice64(array *C.uint64_t, len int) []uint64 {
	var list []uint64
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&list)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
	return list
}

func carray2slice32(array *C.uint32_t, len int) []uint32 {
	var list []uint32
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&list)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
	return list
}

func carray2slice16(array *C.uint16_t, len int) []uint16 {
	var list []uint16
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&list)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
	return list
}

func carray2sliceByte(array *C.char, len int) []byte {
	var list []byte
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&list)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
	return list
}

func (fileInfos *Longtail_FileInfos) GetFileCount() uint32 {
	return uint32(fileInfos.cFileInfos.m_Count)
}

func (fileInfos *Longtail_FileInfos) GetFileSizes() []uint64 {
	size := int(fileInfos.cFileInfos.m_Count)
	return carray2slice64(fileInfos.cFileInfos.m_Sizes, size)
}

func (fileInfos *Longtail_FileInfos) GetFilePermissions() []uint16 {
	size := int(fileInfos.cFileInfos.m_Count)
	return carray2slice16(fileInfos.cFileInfos.m_Permissions, size)
}

func (contentIndex *Longtail_ContentIndex) IsValid() bool {
	return contentIndex.cContentIndex != nil
}

func (contentIndex *Longtail_ContentIndex) Dispose() {
	if contentIndex.cContentIndex != nil {
		C.Longtail_Free(unsafe.Pointer(contentIndex.cContentIndex))
		contentIndex.cContentIndex = nil
	}
}

func (contentIndex *Longtail_ContentIndex) GetVersion() uint32 {
	return uint32(*contentIndex.cContentIndex.m_Version)
}

func (contentIndex *Longtail_ContentIndex) GetHashIdentifier() uint32 {
	return uint32(*contentIndex.cContentIndex.m_HashIdentifier)
}

func (contentIndex *Longtail_ContentIndex) GetMaxBlockSize() uint32 {
	return uint32(*contentIndex.cContentIndex.m_MaxBlockSize)
}

func (contentIndex *Longtail_ContentIndex) GetMaxChunksPerBlock() uint32 {
	return uint32(*contentIndex.cContentIndex.m_MaxChunksPerBlock)
}

func (hashAPI *Longtail_HashAPI) GetIdentifier() uint32 {
	return uint32(C.Longtail_Hash_GetIdentifier(hashAPI.cHashAPI))
}

func (contentIndex *Longtail_ContentIndex) GetBlockCount() uint64 {
	return uint64(*contentIndex.cContentIndex.m_BlockCount)
}

func (contentIndex *Longtail_ContentIndex) GetChunkCount() uint64 {
	return uint64(*contentIndex.cContentIndex.m_ChunkCount)
}

func (contentIndex *Longtail_ContentIndex) GetBlockHashes() []uint64 {
	size := int(C.Longtail_ContentIndex_GetBlockCount(contentIndex.cContentIndex))
	return carray2slice64(C.Longtail_ContentIndex_BlockHashes(contentIndex.cContentIndex), size)
}

func (contentIndex *Longtail_ContentIndex) GetChunkHashes() []uint64 {
	size := int(*contentIndex.cContentIndex.m_ChunkCount)
	return carray2slice64(contentIndex.cContentIndex.m_ChunkHashes, size)
}

func (storeIndex *Longtail_StoreIndex) IsValid() bool {
	return storeIndex.cStoreIndex != nil
}

func (storeIndex *Longtail_StoreIndex) Dispose() {
	if storeIndex.cStoreIndex != nil {
		C.Longtail_Free(unsafe.Pointer(storeIndex.cStoreIndex))
		storeIndex.cStoreIndex = nil
	}
}

func (storeIndex *Longtail_StoreIndex) GetVersion() uint32 {
	return uint32(*storeIndex.cStoreIndex.m_Version)
}

func (storeIndex *Longtail_StoreIndex) GetHashIdentifier() uint32 {
	return uint32(*storeIndex.cStoreIndex.m_HashIdentifier)
}

func (storeIndex *Longtail_StoreIndex) GetBlockCount() uint64 {
	return uint64(*storeIndex.cStoreIndex.m_BlockCount)
}

func (storeIndex *Longtail_StoreIndex) GetChunkCount() uint64 {
	return uint64(*storeIndex.cStoreIndex.m_ChunkCount)
}

func (storeIndex *Longtail_StoreIndex) GetBlockHashes() []uint64 {
	size := int(C.Longtail_StoreIndex_GetBlockCount(storeIndex.cStoreIndex))
	return carray2slice64(C.Longtail_StoreIndex_GetBlockHashes(storeIndex.cStoreIndex), size)
}

func (versionIndex *Longtail_VersionIndex) Dispose() {
	if versionIndex.cVersionIndex != nil {
		C.Longtail_Free(unsafe.Pointer(versionIndex.cVersionIndex))
		versionIndex.cVersionIndex = nil
	}
}

func (versionIndex *Longtail_VersionIndex) GetVersion() uint32 {
	return uint32(*versionIndex.cVersionIndex.m_Version)
}

func (versionIndex *Longtail_VersionIndex) GetHashIdentifier() uint32 {
	return uint32(*versionIndex.cVersionIndex.m_HashIdentifier)
}

func (versionIndex *Longtail_VersionIndex) GetTargetChunkSize() uint32 {
	return uint32(*versionIndex.cVersionIndex.m_TargetChunkSize)
}

func (versionIndex *Longtail_VersionIndex) GetAssetCount() uint32 {
	return uint32(*versionIndex.cVersionIndex.m_AssetCount)
}

func (versionIndex *Longtail_VersionIndex) GetAssetPath(assetIndex uint32) string {
	cPath := C.GetVersionIndexPath(versionIndex.cVersionIndex, C.uint32_t(assetIndex))
	return C.GoString(cPath)
}

func (versionIndex *Longtail_VersionIndex) GetAssetHashes() []uint64 {
	size := int(*versionIndex.cVersionIndex.m_AssetCount)
	return carray2slice64(versionIndex.cVersionIndex.m_ContentHashes, size)
}

func (versionIndex *Longtail_VersionIndex) GetAssetSize(assetIndex uint32) uint64 {
	cSize := C.GetVersionAssetSize(versionIndex.cVersionIndex, C.uint32_t(assetIndex))
	return uint64(cSize)
}

func (versionIndex *Longtail_VersionIndex) GetAssetPermissions(assetIndex uint32) uint16 {
	cPermissions := C.GetVersionAssetPermissions(versionIndex.cVersionIndex, C.uint32_t(assetIndex))
	return uint16(cPermissions)
}

func (versionIndex *Longtail_VersionIndex) GetAssetChunkCounts() []uint32 {
	size := int(*versionIndex.cVersionIndex.m_AssetCount)
	return carray2slice32(versionIndex.cVersionIndex.m_AssetChunkCounts, size)
}

func (versionIndex *Longtail_VersionIndex) GetAssetChunkIndexStarts() []uint32 {
	size := int(*versionIndex.cVersionIndex.m_AssetCount)
	return carray2slice32(versionIndex.cVersionIndex.m_AssetChunkIndexStarts, size)
}

func (versionIndex *Longtail_VersionIndex) GetAssetChunkIndexes() []uint32 {
	size := int(*versionIndex.cVersionIndex.m_AssetChunkIndexCount)
	return carray2slice32(versionIndex.cVersionIndex.m_AssetChunkIndexes, size)
}

func (versionIndex *Longtail_VersionIndex) GetChunkCount() uint32 {
	return uint32(*versionIndex.cVersionIndex.m_ChunkCount)
}

func (versionIndex *Longtail_VersionIndex) GetChunkHashes() []uint64 {
	size := int(*versionIndex.cVersionIndex.m_ChunkCount)
	return carray2slice64(versionIndex.cVersionIndex.m_ChunkHashes, size)
}

func (versionIndex *Longtail_VersionIndex) GetChunkSizes() []uint32 {
	size := int(*versionIndex.cVersionIndex.m_ChunkCount)
	return carray2slice32(versionIndex.cVersionIndex.m_ChunkSizes, size)
}

func (versionIndex *Longtail_VersionIndex) GetAssetSizes() []uint64 {
	size := int(*versionIndex.cVersionIndex.m_AssetCount)
	return carray2slice64(versionIndex.cVersionIndex.m_AssetSizes, size)
}

func (versionIndex *Longtail_VersionIndex) GetChunkTags() []uint32 {
	size := int(*versionIndex.cVersionIndex.m_ChunkCount)
	return carray2slice32(versionIndex.cVersionIndex.m_ChunkTags, size)
}

func (versionDiff *Longtail_VersionDiff) Dispose() {
	if versionDiff.cVersionDiff != nil {
		C.Longtail_Free(unsafe.Pointer(versionDiff.cVersionDiff))
		versionDiff.cVersionDiff = nil
	}
}

// CreateFullHashRegistry ...
func CreateFullHashRegistry() Longtail_HashRegistryAPI {
	return Longtail_HashRegistryAPI{cHashRegistryAPI: C.Longtail_CreateFullHashRegistry()}
}

// CreateBlake3HashRegistry ...
func CreateBlake3HashRegistry() Longtail_HashRegistryAPI {
	return Longtail_HashRegistryAPI{cHashRegistryAPI: C.Longtail_CreateBlake3HashRegistry()}
}

// Longtail_HashRegistryAPI ...
func (hashRegistry *Longtail_HashRegistryAPI) Dispose() {
	if hashRegistry.cHashRegistryAPI != nil {
		C.Longtail_DisposeAPI(&hashRegistry.cHashRegistryAPI.m_API)
		hashRegistry.cHashRegistryAPI = nil
	}
}

// Longtail_HashRegistryAPI ...
func (hashRegistry *Longtail_HashRegistryAPI) GetHashAPI(hashIdentifier uint32) (Longtail_HashAPI, int) {
	var hash_api *C.struct_Longtail_HashAPI
	errno := C.Longtail_GetHashRegistry_GetHashAPI(hashRegistry.cHashRegistryAPI, C.uint32_t(hashIdentifier), &hash_api)
	if errno != 0 {
		return Longtail_HashAPI{cHashAPI: nil}, int(errno)
	}
	return Longtail_HashAPI{cHashAPI: hash_api}, 0
}

// CreateBlake2HashAPI ...
func CreateBlake2HashAPI() Longtail_HashAPI {
	return Longtail_HashAPI{cHashAPI: C.Longtail_CreateBlake2HashAPI()}
}

// CreateBlake3HashAPI ...
func CreateBlake3HashAPI() Longtail_HashAPI {
	return Longtail_HashAPI{cHashAPI: C.Longtail_CreateBlake3HashAPI()}
}

// CreateMeowHashAPI ...
func CreateMeowHashAPI() Longtail_HashAPI {
	return Longtail_HashAPI{cHashAPI: C.Longtail_CreateMeowHashAPI()}
}

// Longtail_HashAPI.Dispose() ...
func (hashAPI *Longtail_HashAPI) Dispose() {
	if hashAPI.cHashAPI != nil {
		C.Longtail_DisposeAPI(&hashAPI.cHashAPI.m_API)
		hashAPI.cHashAPI = nil
	}
}

// CreateHPCDCChunkerAPI ...
func CreateHPCDCChunkerAPI() Longtail_ChunkerAPI {
	return Longtail_ChunkerAPI{cChunkerAPI: C.Longtail_CreateHPCDCChunkerAPI()}
}

// Longtail_ChunkerAPI.Dispose() ...
func (chunkerAPI *Longtail_ChunkerAPI) Dispose() {
	if chunkerAPI.cChunkerAPI != nil {
		C.Longtail_DisposeAPI(&chunkerAPI.cChunkerAPI.m_API)
		chunkerAPI.cChunkerAPI = nil
	}
}

// GetBlake2HashIdentifier() ...
func GetBlake2HashIdentifier() uint32 {
	return uint32(C.Longtail_GetBlake2HashType())
}

// GetBlake3HashIdentifier() ...
func GetBlake3HashIdentifier() uint32 {
	return uint32(C.Longtail_GetBlake3HashType())
}

// GetMeowHashIdentifier() ...
func GetMeowHashIdentifier() uint32 {
	return uint32(C.Longtail_GetMeowHashType())
}

//// Longtail_AsyncPutStoredBlockAPI::OnComplete() ...
func (asyncCompleteAPI *Longtail_AsyncPutStoredBlockAPI) OnComplete(errno int) {
	C.Longtail_AsyncPutStoredBlock_OnComplete(asyncCompleteAPI.cAsyncCompleteAPI, C.int(errno))
}

//// Longtail_AsyncGetStoredBlockAPI::OnComplete() ...
func (asyncCompleteAPI *Longtail_AsyncGetStoredBlockAPI) OnComplete(stored_block Longtail_StoredBlock, errno int) {
	C.Longtail_AsyncGetStoredBlock_OnComplete(asyncCompleteAPI.cAsyncCompleteAPI, stored_block.cStoredBlock, C.int(errno))
}

//// Longtail_AsyncGetExistingContentAPI::OnComplete() ...
func (asyncCompleteAPI *Longtail_AsyncGetExistingContentAPI) OnComplete(content_index Longtail_ContentIndex, errno int) {
	C.Longtail_AsyncGetExistingContent_OnComplete(asyncCompleteAPI.cAsyncCompleteAPI, content_index.cContentIndex, C.int(errno))
}

//// Longtail_AsyncFlushAPI::OnComplete() ...
func (asyncCompleteAPI *Longtail_AsyncFlushAPI) OnComplete(errno int) {
	C.Longtail_AsyncFlush_OnComplete(asyncCompleteAPI.cAsyncCompleteAPI, C.int(errno))
}

// CreateFSBlockStore() ...
func CreateFSBlockStore(jobAPI Longtail_JobAPI, storageAPI Longtail_StorageAPI, contentPath string, defaultMaxBlockSize uint32, defaultMaxChunksPerBlock uint32) Longtail_BlockStoreAPI {
	cContentPath := C.CString(contentPath)
	defer C.free(unsafe.Pointer(cContentPath))
	return Longtail_BlockStoreAPI{cBlockStoreAPI: C.Longtail_CreateFSBlockStoreAPI(jobAPI.cJobAPI, storageAPI.cStorageAPI, cContentPath, C.uint32_t(defaultMaxBlockSize), C.uint32_t(defaultMaxChunksPerBlock), nil)}
}

// CreateCacheBlockStore() ...
func CreateCacheBlockStore(jobAPI Longtail_JobAPI, cacheBlockStore Longtail_BlockStoreAPI, persistentBlockStore Longtail_BlockStoreAPI) Longtail_BlockStoreAPI {
	return Longtail_BlockStoreAPI{cBlockStoreAPI: C.Longtail_CreateCacheBlockStoreAPI(jobAPI.cJobAPI, cacheBlockStore.cBlockStoreAPI, persistentBlockStore.cBlockStoreAPI)}
}

// CreateCompressBlockStore() ...
func CreateCompressBlockStore(backingBlockStore Longtail_BlockStoreAPI, compressionRegistry Longtail_CompressionRegistryAPI) Longtail_BlockStoreAPI {
	return Longtail_BlockStoreAPI{cBlockStoreAPI: C.Longtail_CreateCompressBlockStoreAPI(backingBlockStore.cBlockStoreAPI, compressionRegistry.cCompressionRegistryAPI)}
}

// CreateShareBlockStore() ...
func CreateShareBlockStore(backingBlockStore Longtail_BlockStoreAPI) Longtail_BlockStoreAPI {
	return Longtail_BlockStoreAPI{cBlockStoreAPI: C.Longtail_CreateShareBlockStoreAPI(backingBlockStore.cBlockStoreAPI)}
}

// CreateLRUBlockStoreAPI() ...
func CreateLRUBlockStoreAPI(blockStore Longtail_BlockStoreAPI, cache_block_count uint) Longtail_BlockStoreAPI {
	return Longtail_BlockStoreAPI{cBlockStoreAPI: C.Longtail_CreateLRUBlockStoreAPI(blockStore.cBlockStoreAPI, (C.uint32_t)(cache_block_count))}
}

// CreateBlockStoreStorageAPI() ...
func CreateBlockStoreStorageAPI(
	hashAPI Longtail_HashAPI,
	jobAPI Longtail_JobAPI,
	blockStore Longtail_BlockStoreAPI,
	contentIndex Longtail_ContentIndex,
	versionIndex Longtail_VersionIndex) Longtail_StorageAPI {
	return Longtail_StorageAPI{cStorageAPI: C.Longtail_CreateBlockStoreStorageAPI(
		hashAPI.cHashAPI,
		jobAPI.cJobAPI,
		blockStore.cBlockStoreAPI,
		contentIndex.cContentIndex,
		versionIndex.cVersionIndex)}
}

// Longtail_BlockStoreAPI.Dispose() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) Dispose() {
	if blockStoreAPI.cBlockStoreAPI != nil {
		C.Longtail_DisposeAPI(&blockStoreAPI.cBlockStoreAPI.m_API)
		blockStoreAPI.cBlockStoreAPI = nil
	}
}

//// PutStoredBlock() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) PutStoredBlock(
	storedBlock Longtail_StoredBlock,
	asyncCompleteAPI Longtail_AsyncPutStoredBlockAPI) int {
	errno := C.Longtail_BlockStore_PutStoredBlock(
		blockStoreAPI.cBlockStoreAPI,
		storedBlock.cStoredBlock,
		asyncCompleteAPI.cAsyncCompleteAPI)
	return int(errno)
}

// GetStoredBlock() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) GetStoredBlock(
	blockHash uint64,
	asyncCompleteAPI Longtail_AsyncGetStoredBlockAPI) int {

	errno := C.Longtail_BlockStore_GetStoredBlock(
		blockStoreAPI.cBlockStoreAPI,
		C.uint64_t(blockHash),
		asyncCompleteAPI.cAsyncCompleteAPI)
	return int(errno)
}

// GetExistingContent() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) GetExistingContent(
	chunkHashes []uint64,
	minBlockUsagePercent uint32,
	asyncCompleteAPI Longtail_AsyncGetExistingContentAPI) int {

	chunkCount := len(chunkHashes)
	cChunkHashes := (*C.TLongtail_Hash)(unsafe.Pointer(nil))
	if chunkCount > 0 {
		cChunkHashes = (*C.TLongtail_Hash)(unsafe.Pointer(&chunkHashes[0]))
	}
	errno := C.Longtail_BlockStore_GetExistingContent(
		blockStoreAPI.cBlockStoreAPI,
		C.uint64_t(chunkCount),
		cChunkHashes,
		C.uint32_t(minBlockUsagePercent),
		asyncCompleteAPI.cAsyncCompleteAPI)
	return int(errno)
}

// GetStats() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) GetStats() (BlockStoreStats, int) {
	if blockStoreAPI.cBlockStoreAPI == nil {
		return BlockStoreStats{}, EINVAL
	}
	var cStats C.struct_Longtail_BlockStore_Stats
	errno := C.Longtail_BlockStore_GetStats(
		blockStoreAPI.cBlockStoreAPI,
		&cStats)

	stats := BlockStoreStats{}
	for s := 0; s < Longtail_BlockStoreAPI_StatU64_Count; s++ {
		stats.StatU64[s] = uint64(cStats.m_StatU64[s])
	}
	return stats, int(errno)
}

// Flush() ...
func (blockStoreAPI *Longtail_BlockStoreAPI) Flush(asyncCompleteAPI Longtail_AsyncFlushAPI) int {
	if blockStoreAPI.cBlockStoreAPI == nil {
		asyncCompleteAPI.OnComplete(0)
		return 0
	}
	errno := C.Longtail_BlockStore_Flush(
		blockStoreAPI.cBlockStoreAPI,
		asyncCompleteAPI.cAsyncCompleteAPI)

	return int(errno)
}

func (blockIndex *Longtail_BlockIndex) GetBlockHash() uint64 {
	return uint64(*blockIndex.cBlockIndex.m_BlockHash)
}

func (blockIndex *Longtail_BlockIndex) GetHashIdentifier() uint32 {
	return uint32(*blockIndex.cBlockIndex.m_HashIdentifier)
}

func (blockIndex *Longtail_BlockIndex) GetChunkCount() uint32 {
	return uint32(*blockIndex.cBlockIndex.m_ChunkCount)
}

func (blockIndex *Longtail_BlockIndex) GetTag() uint32 {
	return uint32(*blockIndex.cBlockIndex.m_Tag)
}

func (blockIndex *Longtail_BlockIndex) GetChunkHashes() []uint64 {
	size := int(*blockIndex.cBlockIndex.m_ChunkCount)
	return carray2slice64(blockIndex.cBlockIndex.m_ChunkHashes, size)
}

func (blockIndex *Longtail_BlockIndex) GetChunkSizes() []uint32 {
	size := int(*blockIndex.cBlockIndex.m_ChunkCount)
	return carray2slice32(blockIndex.cBlockIndex.m_ChunkSizes, size)
}

func (blockIndex *Longtail_BlockIndex) IsValid() bool {
	return blockIndex.cBlockIndex != nil
}

func (storedBlock *Longtail_StoredBlock) IsValid() bool {
	return storedBlock.cStoredBlock != nil
}

func (storedBlock *Longtail_StoredBlock) GetBlockIndex() Longtail_BlockIndex {
	return Longtail_BlockIndex{cBlockIndex: storedBlock.cStoredBlock.m_BlockIndex}
}

func (storedBlock *Longtail_StoredBlock) GetBlockSize() int {
	blockIndex := storedBlock.GetBlockIndex()
	chunkCount := C.uint32_t(blockIndex.GetChunkCount())
	blockIndexSize := int(C.Longtail_GetBlockIndexSize(chunkCount))
	blockDataSize := int(storedBlock.cStoredBlock.m_BlockChunksDataSize)
	return blockIndexSize + blockDataSize
}

func (storedBlock *Longtail_StoredBlock) GetChunksBlockData() []byte {
	size := int(storedBlock.cStoredBlock.m_BlockChunksDataSize)
	return carray2sliceByte((*C.char)(storedBlock.cStoredBlock.m_BlockData), size)
}

func (storedBlock *Longtail_StoredBlock) Dispose() {
	if storedBlock.cStoredBlock != nil {
		C.Longtail_StoredBlock_Dispose(storedBlock.cStoredBlock)
		storedBlock.cStoredBlock = nil
	}
}

func (blockIndex *Longtail_BlockIndex) Dispose() {
	if blockIndex.cBlockIndex != nil {
		C.Longtail_Free(unsafe.Pointer(blockIndex.cBlockIndex))
		blockIndex.cBlockIndex = nil
	}
}

func WriteStoredBlockToBuffer(storedBlock Longtail_StoredBlock) ([]byte, int) {
	var buffer unsafe.Pointer
	var size C.size_t
	errno := C.Longtail_WriteStoredBlockToBuffer(storedBlock.cStoredBlock, &buffer, &size)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Free(buffer)
	bytes := C.GoBytes(buffer, C.int(size))
	return bytes, 0
}

func ReadStoredBlockFromBuffer(buffer []byte) (Longtail_StoredBlock, int) {
	cBuffer := unsafe.Pointer(&buffer[0])
	size := C.size_t(len(buffer))
	var stored_block *C.struct_Longtail_StoredBlock
	errno := C.Longtail_ReadStoredBlockFromBuffer(cBuffer, size, &stored_block)
	if errno != 0 {
		return Longtail_StoredBlock{cStoredBlock: nil}, int(errno)
	}
	if stored_block == nil {
		return Longtail_StoredBlock{cStoredBlock: nil}, EBADF
	}
	return Longtail_StoredBlock{cStoredBlock: stored_block}, 0
}

func ValidateContent(contentIndex Longtail_ContentIndex, versionIndex Longtail_VersionIndex) int {
	errno := C.Longtail_ValidateContent(contentIndex.cContentIndex, versionIndex.cVersionIndex)
	return int(errno)
}

// CreateStoredBlock() ...
func CreateStoredBlock(
	blockHash uint64,
	hashIdentifier uint32,
	compressionType uint32,
	chunkHashes []uint64,
	chunkSizes []uint32,
	blockData []uint8,
	blockDataIncludesIndex bool) (Longtail_StoredBlock, int) {
	chunkCount := len(chunkHashes)
	if chunkCount != len(chunkSizes) {
		return Longtail_StoredBlock{cStoredBlock: nil}, EINVAL
	}
	cChunkHashes := (*C.TLongtail_Hash)(unsafe.Pointer(nil))
	if chunkCount > 0 {
		cChunkHashes = (*C.TLongtail_Hash)(unsafe.Pointer(&chunkHashes[0]))
	}
	cChunkSizes := (*C.uint32_t)(unsafe.Pointer(nil))
	if chunkCount > 0 {
		cChunkSizes = (*C.uint32_t)(unsafe.Pointer(&chunkSizes[0]))
	}
	blockByteCount := len(blockData)
	blockByteOffset := 0
	if blockDataIncludesIndex {
		blockByteOffset = int(C.Longtail_GetBlockIndexSize((C.uint32_t)(chunkCount)))
		blockByteCount = blockByteCount - blockByteOffset
	}
	var cStoredBlock *C.struct_Longtail_StoredBlock
	errno := C.Longtail_CreateStoredBlock(
		C.TLongtail_Hash(blockHash),
		C.uint32_t(hashIdentifier),
		C.uint32_t(chunkCount),
		C.uint32_t(compressionType),
		cChunkHashes,
		cChunkSizes,
		C.uint32_t(blockByteCount),
		&cStoredBlock)
	if errno != 0 {
		return Longtail_StoredBlock{}, int(errno)
	}
	cBlockBytes := unsafe.Pointer(nil)
	if blockByteCount > 0 {
		cBlockBytes = unsafe.Pointer(&blockData[blockByteOffset])
	}
	C.memmove(cStoredBlock.m_BlockData, cBlockBytes, C.size_t(blockByteCount))
	return Longtail_StoredBlock{cStoredBlock: cStoredBlock}, 0
}

// CreateFSStorageAPI ...
func CreateFSStorageAPI() Longtail_StorageAPI {
	return Longtail_StorageAPI{cStorageAPI: C.Longtail_CreateFSStorageAPI()}
}

// CreateInMemStorageAPI ...
func CreateInMemStorageAPI() Longtail_StorageAPI {
	return Longtail_StorageAPI{cStorageAPI: C.Longtail_CreateInMemStorageAPI()}
}

// Longtail_StorageAPI.Dispose() ...
func (storageAPI *Longtail_StorageAPI) Dispose() {
	if storageAPI.cStorageAPI != nil {
		C.Longtail_DisposeAPI(&storageAPI.cStorageAPI.m_API)
		storageAPI.cStorageAPI = nil
	}
}

// CreateBrotliCompressionAPI ...
func CreateBrotliCompressionAPI() Longtail_CompressionAPI {
	return Longtail_CompressionAPI{cCompressionAPI: C.Longtail_CreateBrotliCompressionAPI()}
}

// CreateLZ4CompressionAPI ...
func CreateLZ4CompressionAPI() Longtail_CompressionAPI {
	return Longtail_CompressionAPI{cCompressionAPI: C.Longtail_CreateLZ4CompressionAPI()}
}

// CreateZStdCompressionAPI ...
func CreateZStdCompressionAPI() Longtail_CompressionAPI {
	return Longtail_CompressionAPI{cCompressionAPI: C.Longtail_CreateZStdCompressionAPI()}
}

// Longtail_CompressionAPI.Dispose() ...
func (compressionAPI *Longtail_CompressionAPI) Dispose() {
	if compressionAPI.cCompressionAPI != nil {
		C.Longtail_DisposeAPI(&compressionAPI.cCompressionAPI.m_API)
		compressionAPI.cCompressionAPI = nil
	}
}

// CreateBikeshedJobAPI ...
func CreateBikeshedJobAPI(workerCount uint32, workerPriority int) Longtail_JobAPI {
	return Longtail_JobAPI{cJobAPI: C.Longtail_CreateBikeshedJobAPI(C.uint32_t(workerCount), C.int(workerPriority))}
}

// CreateRateLimitedProgressAPI ...
func CreateRateLimitedProgressAPI(progressAPI Longtail_ProgressAPI, percentRateLimit uint32) Longtail_ProgressAPI {
	return Longtail_ProgressAPI{cProgressAPI: C.Longtail_CreateRateLimitedProgress(progressAPI.cProgressAPI, C.uint32_t(percentRateLimit))}
}

// OnProgress ...
func (progressAPI *Longtail_ProgressAPI) OnProgress(totalCount uint32, doneCount uint32) {
	C.Longtail_Progress_OnProgress(progressAPI.cProgressAPI, C.uint32_t(totalCount), C.uint32_t(doneCount))
}

// Longtail_ProgressAPI.Dispose() ...
func (progressAPI *Longtail_ProgressAPI) Dispose() {
	if progressAPI.cProgressAPI != nil {
		C.Longtail_DisposeAPI(&progressAPI.cProgressAPI.m_API)
		progressAPI.cProgressAPI = nil
	}
}

// Longtail_JobAPI.Dispose() ...
func (jobAPI *Longtail_JobAPI) Dispose() {
	if jobAPI.cJobAPI != nil {
		C.Longtail_DisposeAPI(&jobAPI.cJobAPI.m_API)
		jobAPI.cJobAPI = nil
	}
}

// CreateFullCompressionRegistry ...
func CreateFullCompressionRegistry() Longtail_CompressionRegistryAPI {
	return Longtail_CompressionRegistryAPI{cCompressionRegistryAPI: C.Longtail_CreateFullCompressionRegistry()}
}

// CreateZStdCompressionRegistry ...
func CreateZStdCompressionRegistry() Longtail_CompressionRegistryAPI {
	return Longtail_CompressionRegistryAPI{cCompressionRegistryAPI: C.Longtail_CreateZStdCompressionRegistry()}
}

// Longtail_CompressionRegistryAPI ...
func (compressionRegistry *Longtail_CompressionRegistryAPI) Dispose() {
	if compressionRegistry.cCompressionRegistryAPI != nil {
		C.Longtail_DisposeAPI(&compressionRegistry.cCompressionRegistryAPI.m_API)
		compressionRegistry.cCompressionRegistryAPI = nil
	}
}

// GetNoCompressionType ...
func GetNoCompressionType() uint32 {
	return uint32(0)
}

// GetBrotliGenericMinCompressionType ...
func GetBrotliGenericMinCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliGenericMinQuality())
}

// GetBrotliGenericDefaultCompressionType ...
func GetBrotliGenericDefaultCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliGenericDefaultQuality())
}

// GetBrotliGenericMaxCompressionType ...
func GetBrotliGenericMaxCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliGenericMaxQuality())
}

// GetBrotliTextMinCompressionType ...
func GetBrotliTextMinCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliTextMinQuality())
}

// GetBrotliTextDefaultCompressionType ...
func GetBrotliTextDefaultCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliTextDefaultQuality())
}

// GetBrotliTextMaxCompressionType ...
func GetBrotliTextMaxCompressionType() uint32 {
	return uint32(C.Longtail_GetBrotliTextMaxQuality())
}

// GetLZ4DefaultCompressionType ...
func GetLZ4DefaultCompressionType() uint32 {
	return uint32(C.Longtail_GetLZ4DefaultQuality())
}

// GetZStdMinCompressionType ...
func GetZStdMinCompressionType() uint32 {
	return uint32(C.Longtail_GetZStdMinQuality())
}

// GetZStdDefaultCompressionType ...
func GetZStdDefaultCompressionType() uint32 {
	return uint32(C.Longtail_GetZStdDefaultQuality())
}

// GetZStdMaxCompressionType ...
func GetZStdMaxCompressionType() uint32 {
	return uint32(C.Longtail_GetZStdMaxQuality())
}

// LongtailAlloc ...
func LongtailAlloc(context string, size uint64) unsafe.Pointer {
	return C.Longtail_Alloc(C.CString(context), C.size_t(size))
}

// LongtailFree ...
func LongtailFree(data unsafe.Pointer) {
	C.Longtail_Free(data)
}

// GetFilesRecursively ...
func GetFilesRecursively(storageAPI Longtail_StorageAPI, pathFilter Longtail_PathFilterAPI, rootPath string) (Longtail_FileInfos, int) {
	cFolderPath := C.CString(rootPath)
	defer C.free(unsafe.Pointer(cFolderPath))
	var fileInfos *C.struct_Longtail_FileInfos
	errno := C.Longtail_GetFilesRecursively(storageAPI.cStorageAPI, pathFilter.cPathFilterAPI, nil, nil, cFolderPath, &fileInfos)
	if errno != 0 {
		return Longtail_FileInfos{cFileInfos: nil}, int(errno)
	}
	return Longtail_FileInfos{cFileInfos: fileInfos}, 0
}

// GetPathCount ...
func (fileInfos *Longtail_FileInfos) GetPathCount() uint32 {
	return uint32(fileInfos.cFileInfos.m_Count)
}

// GetPath ...
func (fileInfos Longtail_FileInfos) GetPath(index uint32) string {
	cPath := C.Longtail_FileInfos_GetPath(fileInfos.cFileInfos, C.uint32_t(index))
	return C.GoString(cPath)
}

// WriteBlockIndexToBuffer ...
func WriteBlockIndexToBuffer(index Longtail_BlockIndex) ([]byte, int) {
	var buffer unsafe.Pointer
	size := C.size_t(0)
	errno := C.Longtail_WriteBlockIndexToBuffer(index.cBlockIndex, &buffer, &size)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Free(buffer)
	bytes := C.GoBytes(buffer, C.int(size))
	return bytes, 0
}

// ReadBlockIndexFromBuffer ...
func ReadBlockIndexFromBuffer(buffer []byte) (Longtail_BlockIndex, int) {
	cBuffer := unsafe.Pointer(&buffer[0])
	cSize := C.size_t(len(buffer))
	var bindex *C.struct_Longtail_BlockIndex
	errno := C.Longtail_ReadBlockIndexFromBuffer(cBuffer, cSize, &bindex)
	if errno != 0 {
		return Longtail_BlockIndex{cBlockIndex: nil}, int(errno)
	}
	return Longtail_BlockIndex{cBlockIndex: bindex}, 0
}

// CreateVersionIndex ...
func CreateVersionIndex(
	storageAPI Longtail_StorageAPI,
	hashAPI Longtail_HashAPI,
	chunkerAPI Longtail_ChunkerAPI,
	jobAPI Longtail_JobAPI,
	progressAPI *Longtail_ProgressAPI,
	rootPath string,
	fileInfos Longtail_FileInfos,
	assetCompressionTypes []uint32,
	maxChunkSize uint32) (Longtail_VersionIndex, int) {

	var cProgressAPI *C.struct_Longtail_ProgressAPI
	if progressAPI != nil {
		cProgressAPI = progressAPI.cProgressAPI
	}

	cRootPath := C.CString(rootPath)
	defer C.free(unsafe.Pointer(cRootPath))

	cCompressionTypes := unsafe.Pointer(nil)
	if len(assetCompressionTypes) > 0 {
		cCompressionTypes = unsafe.Pointer(&assetCompressionTypes[0])
	}

	var vindex *C.struct_Longtail_VersionIndex
	errno := C.Longtail_CreateVersionIndex(
		storageAPI.cStorageAPI,
		hashAPI.cHashAPI,
		chunkerAPI.cChunkerAPI,
		jobAPI.cJobAPI,
		cProgressAPI,
		nil,
		nil,
		cRootPath,
		fileInfos.cFileInfos,
		(*C.uint32_t)(cCompressionTypes),
		C.uint32_t(maxChunkSize),
		&vindex)

	if errno != 0 {
		return Longtail_VersionIndex{cVersionIndex: nil}, int(errno)
	}

	return Longtail_VersionIndex{cVersionIndex: vindex}, 0
}

// WriteVersionIndexToBuffer ...
func WriteVersionIndexToBuffer(index Longtail_VersionIndex) ([]byte, int) {
	var buffer unsafe.Pointer
	size := C.size_t(0)
	errno := C.Longtail_WriteVersionIndexToBuffer(index.cVersionIndex, &buffer, &size)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Free(buffer)
	bytes := C.GoBytes(buffer, C.int(size))
	return bytes, 0
}

// WriteVersionIndex ...
func WriteVersionIndex(storageAPI Longtail_StorageAPI, index Longtail_VersionIndex, path string) int {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	errno := C.Longtail_WriteVersionIndex(storageAPI.cStorageAPI, index.cVersionIndex, cPath)
	if errno != 0 {
		return int(errno)
	}
	return 0
}

// ReadVersionIndexFromBuffer ...
func ReadVersionIndexFromBuffer(buffer []byte) (Longtail_VersionIndex, int) {
	cBuffer := unsafe.Pointer(&buffer[0])
	cSize := C.size_t(len(buffer))
	var vindex *C.struct_Longtail_VersionIndex
	errno := C.Longtail_ReadVersionIndexFromBuffer(cBuffer, cSize, &vindex)
	if errno != 0 {
		return Longtail_VersionIndex{cVersionIndex: nil}, int(errno)
	}
	return Longtail_VersionIndex{cVersionIndex: vindex}, 0
}

// ReadVersionIndex ...
func ReadVersionIndex(storageAPI Longtail_StorageAPI, path string) (Longtail_VersionIndex, int) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var vindex *C.struct_Longtail_VersionIndex
	errno := C.Longtail_ReadVersionIndex(storageAPI.cStorageAPI, cPath, &vindex)
	if errno != 0 {
		return Longtail_VersionIndex{cVersionIndex: nil}, int(errno)
	}
	return Longtail_VersionIndex{cVersionIndex: vindex}, 0
}

// CreateContentIndex
func CreateContentIndex(
	hashAPI Longtail_HashAPI,
	versionIndex Longtail_VersionIndex,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {

	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_CreateContentIndex(
		hashAPI.cHashAPI,
		versionIndex.cVersionIndex,
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}

	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

/*
// CreateContentIndexFromDiff
func CreateContentIndexFromDiff(
	hashAPI Longtail_HashAPI,
	versionIndex Longtail_VersionIndex,
	versionDiff Longtail_VersionDiff,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {

	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_CreateContentIndexFromDiff(
		hashAPI.cHashAPI,
		versionIndex.cVersionIndex,
		versionDiff.cVersionDiff,
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}

	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}
*/
// CreateContentIndexRaw ...
func CreateContentIndexRaw(
	hashAPI Longtail_HashAPI,
	chunkHashes []uint64,
	chunkSizes []uint32,
	compressionTypes []uint32,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {

	chunkCount := uint64(len(chunkHashes))
	var cindex *C.struct_Longtail_ContentIndex
	if chunkCount == 0 {
		errno := C.Longtail_CreateContentIndexRaw(
			hashAPI.cHashAPI,
			0,
			nil,
			nil,
			nil,
			C.uint32_t(maxBlockSize),
			C.uint32_t(maxChunksPerBlock),
			&cindex)
		if errno != 0 {
			return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
		}
	}
	var cChunkHashes *C.TLongtail_Hash
	var cChunkSizes *C.uint32_t
	var cCompressionTypes *C.uint32_t
	if chunkCount > 0 {
		cChunkHashes = (*C.TLongtail_Hash)(unsafe.Pointer(&chunkHashes[0]))
		cChunkSizes = (*C.uint32_t)(unsafe.Pointer(&chunkSizes[0]))
		cCompressionTypes = (*C.uint32_t)(unsafe.Pointer(&compressionTypes[0]))
	}

	errno := C.Longtail_CreateContentIndexRaw(
		hashAPI.cHashAPI,
		C.uint64_t(chunkCount),
		cChunkHashes,
		cChunkSizes,
		cCompressionTypes,
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&cindex)

	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}

	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

// CreateContentIndexFromBlocks ...
func CreateContentIndexFromBlocks(
	maxBlockSize uint32,
	maxChunksPerBlock uint32,
	blockIndexes []Longtail_BlockIndex) (Longtail_ContentIndex, int) {
	rawBlockIndexes := make([]*C.struct_Longtail_BlockIndex, len(blockIndexes))
	blockCount := len(blockIndexes)
	for index, blockIndex := range blockIndexes {
		rawBlockIndexes[index] = blockIndex.cBlockIndex
	}
	var cBlockIndexes unsafe.Pointer
	if blockCount > 0 {
		cBlockIndexes = unsafe.Pointer(&rawBlockIndexes[0])
	}

	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_CreateContentIndexFromBlocks(
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		C.uint64_t(blockCount),
		(**C.struct_Longtail_BlockIndex)(cBlockIndexes),
		&cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

func CreateContentIndexFromStoreIndex(
	storeIndex Longtail_StoreIndex,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {
	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_CreateContentIndexFromStoreIndex(
		storeIndex.cStoreIndex,
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

// CreateStoreIndexFromBlocks ...
func CreateStoreIndexFromBlocks(blockIndexes []Longtail_BlockIndex) (Longtail_StoreIndex, int) {
	rawBlockIndexes := make([]*C.struct_Longtail_BlockIndex, len(blockIndexes))
	blockCount := len(blockIndexes)
	for index, blockIndex := range blockIndexes {
		rawBlockIndexes[index] = blockIndex.cBlockIndex
	}
	var cBlockIndexes unsafe.Pointer
	if blockCount > 0 {
		cBlockIndexes = unsafe.Pointer(&rawBlockIndexes[0])
	}
	var sindex *C.struct_Longtail_StoreIndex
	errno := C.Longtail_CreateStoreIndexFromBlocks(
		C.uint32_t(blockCount),
		(**C.struct_Longtail_BlockIndex)(cBlockIndexes),
		&sindex)
	if errno != 0 {
		return Longtail_StoreIndex{cStoreIndex: nil}, int(errno)
	}
	return Longtail_StoreIndex{cStoreIndex: sindex}, 0
}

func GetExistingContentIndex(
	storeIndex Longtail_StoreIndex,
	chunkHashes []uint64,
	minBlockUsagePercent uint32,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {
	chunkCount := uint64(len(chunkHashes))
	var cChunkHashes *C.TLongtail_Hash
	if chunkCount > 0 {
		cChunkHashes = (*C.TLongtail_Hash)(unsafe.Pointer(&chunkHashes[0]))
	}
	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_GetExistingContentIndex(
		storeIndex.cStoreIndex,
		C.uint32_t(chunkCount),
		cChunkHashes,
		C.uint32_t(minBlockUsagePercent),
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

func GetRequiredChunkHashes(
	versionIndex Longtail_VersionIndex,
	versionDiff Longtail_VersionDiff) ([]uint64, int) {
	maxChunkCount := uint64(versionIndex.GetChunkCount())
	outChunkHashes := make([]uint64, maxChunkCount)
	var cChunkHashes *C.TLongtail_Hash
	if maxChunkCount > 0 {
		cChunkHashes = (*C.TLongtail_Hash)(unsafe.Pointer(&outChunkHashes[0]))
	}
	var outChunkCount C.uint64_t
	errno := C.Longtail_GetRequiredChunkHashes(
		versionIndex.cVersionIndex,
		versionDiff.cVersionDiff,
		&outChunkCount,
		cChunkHashes)
	if errno != 0 {
		return []uint64{}, int(errno)
	}
	return outChunkHashes[:int(outChunkCount)], 0
}

func MergeStoreIndex(local_store_index Longtail_StoreIndex, remote_store_index Longtail_StoreIndex) (Longtail_StoreIndex, int) {
	var sIndex *C.struct_Longtail_StoreIndex
	errno := C.Longtail_MergeStoreIndex(
		local_store_index.cStoreIndex,
		remote_store_index.cStoreIndex,
		&sIndex)
	if errno != 0 {
		return Longtail_StoreIndex{cStoreIndex: nil}, int(errno)
	}
	return Longtail_StoreIndex{cStoreIndex: sIndex}, 0
}

// WriteStoreIndexToBuffer ...
func WriteStoreIndexToBuffer(index Longtail_StoreIndex) ([]byte, int) {
	var buffer unsafe.Pointer
	size := C.size_t(0)
	errno := C.Longtail_WriteStoreIndexToBuffer(index.cStoreIndex, &buffer, &size)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Free(buffer)
	bytes := C.GoBytes(buffer, C.int(size))
	return bytes, 0
}

// ReadStoreIndexFromBuffer ...
func ReadStoreIndexFromBuffer(buffer []byte) (Longtail_StoreIndex, int) {
	cBuffer := unsafe.Pointer(&buffer[0])
	cSize := C.size_t(len(buffer))
	var cindex *C.struct_Longtail_StoreIndex
	errno := C.Longtail_ReadStoreIndexFromBuffer(cBuffer, cSize, &cindex)
	if errno != 0 {
		return Longtail_StoreIndex{cStoreIndex: nil}, int(errno)
	}
	return Longtail_StoreIndex{cStoreIndex: cindex}, 0
}

// WriteContentIndexToBuffer ...
func WriteContentIndexToBuffer(index Longtail_ContentIndex) ([]byte, int) {
	var buffer unsafe.Pointer
	size := C.size_t(0)
	errno := C.Longtail_WriteContentIndexToBuffer(index.cContentIndex, &buffer, &size)
	if errno != 0 {
		return nil, int(errno)
	}
	defer C.Longtail_Free(buffer)
	bytes := C.GoBytes(buffer, C.int(size))
	return bytes, 0
}

// WriteContentIndex ...
func WriteContentIndex(storageAPI Longtail_StorageAPI, index Longtail_ContentIndex, path string) int {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	errno := C.Longtail_WriteContentIndex(storageAPI.cStorageAPI, index.cContentIndex, cPath)
	if errno != 0 {
		return int(errno)
	}
	return 0
}

// ReadContentIndexFromBuffer ...
func ReadContentIndexFromBuffer(buffer []byte) (Longtail_ContentIndex, int) {
	cBuffer := unsafe.Pointer(&buffer[0])
	cSize := C.size_t(len(buffer))
	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_ReadContentIndexFromBuffer(cBuffer, cSize, &cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

// ReadContentIndex ...
func ReadContentIndex(storageAPI Longtail_StorageAPI, path string) (Longtail_ContentIndex, int) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cindex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_ReadContentIndex(storageAPI.cStorageAPI, cPath, &cindex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: cindex}, 0
}

// CreateProgressAPI ...
func CreateProgressAPI(progress ProgressAPI) Longtail_ProgressAPI {
	cContext := SavePointer(progress)
	progressAPIProxy := C.CreateProgressProxyAPI(cContext)
	return Longtail_ProgressAPI{cProgressAPI: progressAPIProxy}
}

//export ProgressAPIProxy_OnProgress
func ProgressAPIProxy_OnProgress(progress_api *C.struct_Longtail_ProgressAPI, total_count C.uint32_t, done_count C.uint32_t) {
	context := C.ProgressAPIProxy_GetContext(unsafe.Pointer(progress_api))
	progress := RestorePointer(context).(ProgressAPI)
	progress.OnProgress(uint32(total_count), uint32(done_count))
}

//export ProgressAPIProxy_Dispose
func ProgressAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.ProgressAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// CreatePathFilterAPI ...
func CreatePathFilterAPI(pathFilter PathFilterAPI) Longtail_PathFilterAPI {
	cContext := SavePointer(pathFilter)
	pathFilterAPIProxy := C.CreatePathFilterProxyAPI(cContext)
	return Longtail_PathFilterAPI{cPathFilterAPI: pathFilterAPIProxy}
}

//export PathFilterAPIProxy_Include
func PathFilterAPIProxy_Include(path_filter_api *C.struct_Longtail_PathFilterAPI, root_path *C.char, asset_path *C.char, asset_name *C.char, is_dir C.int, size C.uint64_t, permissions C.uint16_t) C.int {
	context := C.PathFilterAPIProxy_GetContext(unsafe.Pointer(path_filter_api))
	pathFilter := RestorePointer(context).(PathFilterAPI)
	isDir := is_dir != 0
	if pathFilter.Include(C.GoString(root_path), C.GoString(asset_path), C.GoString(asset_name), isDir, uint64(size), uint16(permissions)) {
		return 1
	}
	return 0
}

//export PathFilterAPIProxy_Dispose
func PathFilterAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.PathFilterAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// CreateAsyncPutStoredBlockAPI ...
func CreateAsyncPutStoredBlockAPI(asyncComplete AsyncPutStoredBlockAPI) Longtail_AsyncPutStoredBlockAPI {
	cContext := SavePointer(asyncComplete)
	asyncCompleteAPIProxy := C.CreateAsyncPutStoredBlockAPI(cContext)
	return Longtail_AsyncPutStoredBlockAPI{cAsyncCompleteAPI: asyncCompleteAPIProxy}
}

//export AsyncPutStoredBlockAPIProxy_OnComplete
func AsyncPutStoredBlockAPIProxy_OnComplete(async_complete_api *C.struct_Longtail_AsyncPutStoredBlockAPI, err C.int) {
	context := C.AsyncPutStoredBlockAPIProxy_GetContext(unsafe.Pointer(async_complete_api))
	asyncComplete := RestorePointer(context).(AsyncPutStoredBlockAPI)
	asyncComplete.OnComplete(int(err))
	C.Longtail_DisposeAPI(&async_complete_api.m_API)
}

//export AsyncPutStoredBlockAPIProxy_Dispose
func AsyncPutStoredBlockAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.AsyncPutStoredBlockAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// CreateAsyncGetStoredBlockAPI ...
func CreateAsyncGetStoredBlockAPI(asyncComplete AsyncGetStoredBlockAPI) Longtail_AsyncGetStoredBlockAPI {
	cContext := SavePointer(asyncComplete)
	asyncCompleteAPIProxy := C.CreateAsyncGetStoredBlockAPI(cContext)
	return Longtail_AsyncGetStoredBlockAPI{cAsyncCompleteAPI: asyncCompleteAPIProxy}
}

//export AsyncGetStoredBlockAPIProxy_OnComplete
func AsyncGetStoredBlockAPIProxy_OnComplete(async_complete_api *C.struct_Longtail_AsyncGetStoredBlockAPI, stored_block *C.struct_Longtail_StoredBlock, err C.int) {
	context := C.AsyncGetStoredBlockAPIProxy_GetContext(unsafe.Pointer(async_complete_api))
	asyncComplete := RestorePointer(context).(AsyncGetStoredBlockAPI)
	asyncComplete.OnComplete(Longtail_StoredBlock{cStoredBlock: stored_block}, int(err))
	C.Longtail_DisposeAPI(&async_complete_api.m_API)
}

//export AsyncGetStoredBlockAPIProxy_Dispose
func AsyncGetStoredBlockAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.AsyncGetStoredBlockAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// CreateAsyncGetExistingContentAPI ...
func CreateAsyncGetExistingContentAPI(asyncComplete AsyncGetExistingContentAPI) Longtail_AsyncGetExistingContentAPI {
	cContext := SavePointer(asyncComplete)
	asyncCompleteAPIProxy := C.CreateAsyncGetExistingContentAPI(cContext)
	return Longtail_AsyncGetExistingContentAPI{cAsyncCompleteAPI: asyncCompleteAPIProxy}
}

//export AsyncGetExistingContentAPIProxy_OnComplete
func AsyncGetExistingContentAPIProxy_OnComplete(async_complete_api *C.struct_Longtail_AsyncGetExistingContentAPI, content_index *C.struct_Longtail_ContentIndex, err C.int) {
	context := C.AsyncGetExistingContentAPIProxy_GetContext(unsafe.Pointer(async_complete_api))
	asyncComplete := RestorePointer(context).(AsyncGetExistingContentAPI)
	asyncComplete.OnComplete(Longtail_ContentIndex{cContentIndex: content_index}, int(err))
	C.Longtail_DisposeAPI(&async_complete_api.m_API)
}

//export AsyncGetExistingContentAPIProxy_Dispose
func AsyncGetExistingContentAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.AsyncGetExistingContentAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// CreateAsyncFlushAPI ...
func CreateAsyncFlushAPI(asyncComplete AsyncFlushAPI) Longtail_AsyncFlushAPI {
	cContext := SavePointer(asyncComplete)
	asyncCompleteAPIProxy := C.CreateAsyncFlushAPI(cContext)
	return Longtail_AsyncFlushAPI{cAsyncCompleteAPI: asyncCompleteAPIProxy}
}

//export AsyncFlushAPIProxy_OnComplete
func AsyncFlushAPIProxy_OnComplete(async_complete_api *C.struct_Longtail_AsyncFlushAPI, err C.int) {
	context := C.AsyncFlushAPIProxy_GetContext(unsafe.Pointer(async_complete_api))
	asyncComplete := RestorePointer(context).(AsyncFlushAPI)
	asyncComplete.OnComplete(int(err))
	C.Longtail_DisposeAPI(&async_complete_api.m_API)
}

//export AsyncFlushAPIProxy_Dispose
func AsyncFlushAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.AsyncFlushAPIProxy_GetContext(unsafe.Pointer(api))
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

// WriteContent ...
func WriteContent(
	sourceStorageAPI Longtail_StorageAPI,
	targetBlockStoreAPI Longtail_BlockStoreAPI,
	jobAPI Longtail_JobAPI,
	progressAPI *Longtail_ProgressAPI,
	content_index Longtail_ContentIndex,
	versionIndex Longtail_VersionIndex,
	versionFolderPath string) int {

	var cProgressAPI *C.struct_Longtail_ProgressAPI
	if progressAPI != nil {
		cProgressAPI = progressAPI.cProgressAPI
	}

	cVersionFolderPath := C.CString(versionFolderPath)
	defer C.free(unsafe.Pointer(cVersionFolderPath))

	errno := C.Longtail_WriteContent(
		sourceStorageAPI.cStorageAPI,
		targetBlockStoreAPI.cBlockStoreAPI,
		jobAPI.cJobAPI,
		cProgressAPI,
		nil,
		nil,
		content_index.cContentIndex,
		versionIndex.cVersionIndex,
		cVersionFolderPath)
	if errno != 0 {
		return int(errno)
	}
	return 0
}

// CreateMissingContent ...
func CreateMissingContent(
	hashAPI Longtail_HashAPI,
	contentIndex Longtail_ContentIndex,
	versionIndex Longtail_VersionIndex,
	maxBlockSize uint32,
	maxChunksPerBlock uint32) (Longtail_ContentIndex, int) {

	var missingContentIndex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_CreateMissingContent(
		hashAPI.cHashAPI,
		contentIndex.cContentIndex,
		versionIndex.cVersionIndex,
		C.uint32_t(maxBlockSize),
		C.uint32_t(maxChunksPerBlock),
		&missingContentIndex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: missingContentIndex}, 0
}

/*
// GetExistingContent ...
func GetExistingContent(
	referenceContentIndex Longtail_ContentIndex,
	contentIndex Longtail_ContentIndex) (Longtail_ContentIndex, int) {
	var retargetedContentIndex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_GetExistingContent(referenceContentIndex.cContentIndex, contentIndex.cContentIndex, &retargetedContentIndex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: retargetedContentIndex}, 0
}
*/
// MergeContentIndex ...
func MergeContentIndex(
	jobAPI Longtail_JobAPI,
	localContentIndex Longtail_ContentIndex,
	remoteContentIndex Longtail_ContentIndex) (Longtail_ContentIndex, int) {
	var mergedContentIndex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_MergeContentIndex(jobAPI.cJobAPI, localContentIndex.cContentIndex, remoteContentIndex.cContentIndex, &mergedContentIndex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: mergedContentIndex}, 0
}

// AddContentIndex ...
func AddContentIndex(
	localContentIndex Longtail_ContentIndex,
	remoteContentIndex Longtail_ContentIndex) (Longtail_ContentIndex, int) {
	var mergedContentIndex *C.struct_Longtail_ContentIndex
	errno := C.Longtail_AddContentIndex(localContentIndex.cContentIndex, remoteContentIndex.cContentIndex, &mergedContentIndex)
	if errno != 0 {
		return Longtail_ContentIndex{cContentIndex: nil}, int(errno)
	}
	return Longtail_ContentIndex{cContentIndex: mergedContentIndex}, 0
}

// WriteVersion ...
func WriteVersion(
	contentBlockStoreAPI Longtail_BlockStoreAPI,
	versionStorageAPI Longtail_StorageAPI,
	jobAPI Longtail_JobAPI,
	progressAPI *Longtail_ProgressAPI,
	contentIndex Longtail_ContentIndex,
	versionIndex Longtail_VersionIndex,
	versionFolderPath string,
	retainPermissions bool) int {

	var cProgressAPI *C.struct_Longtail_ProgressAPI
	if progressAPI != nil {
		cProgressAPI = progressAPI.cProgressAPI
	}

	cVersionFolderPath := C.CString(versionFolderPath)
	defer C.free(unsafe.Pointer(cVersionFolderPath))

	cRetainPermissions := C.int(0)
	if retainPermissions {
		cRetainPermissions = C.int(1)
	}

	errno := C.Longtail_WriteVersion(
		contentBlockStoreAPI.cBlockStoreAPI,
		versionStorageAPI.cStorageAPI,
		jobAPI.cJobAPI,
		cProgressAPI,
		nil,
		nil,
		contentIndex.cContentIndex,
		versionIndex.cVersionIndex,
		cVersionFolderPath,
		cRetainPermissions)
	if errno != 0 {
		return int(errno)
	}
	return 0
}

//CreateVersionDiff do we really need this? Maybe ChangeVersion should create one on the fly?
func CreateVersionDiff(
	hashAPI Longtail_HashAPI,
	sourceVersionIndex Longtail_VersionIndex,
	targetVersionIndex Longtail_VersionIndex) (Longtail_VersionDiff, int) {
	var versionDiff *C.struct_Longtail_VersionDiff
	errno := C.Longtail_CreateVersionDiff(
		hashAPI.cHashAPI,
		sourceVersionIndex.cVersionIndex,
		targetVersionIndex.cVersionIndex,
		&versionDiff)
	if errno != 0 {
		return Longtail_VersionDiff{cVersionDiff: nil}, int(errno)
	}
	return Longtail_VersionDiff{cVersionDiff: versionDiff}, 0
}

//ChangeVersion ...
func ChangeVersion(
	contentBlockStoreAPI Longtail_BlockStoreAPI,
	versionStorageAPI Longtail_StorageAPI,
	hashAPI Longtail_HashAPI,
	jobAPI Longtail_JobAPI,
	progressAPI *Longtail_ProgressAPI,
	contentIndex Longtail_ContentIndex,
	sourceVersionIndex Longtail_VersionIndex,
	targetVersionIndex Longtail_VersionIndex,
	versionDiff Longtail_VersionDiff,
	versionFolderPath string,
	retainPermissions bool) int {

	var cProgressAPI *C.struct_Longtail_ProgressAPI
	if progressAPI != nil {
		cProgressAPI = progressAPI.cProgressAPI
	}

	cVersionFolderPath := C.CString(versionFolderPath)
	defer C.free(unsafe.Pointer(cVersionFolderPath))

	cRetainPermissions := C.int(0)
	if retainPermissions {
		cRetainPermissions = C.int(1)
	}

	errno := C.Longtail_ChangeVersion(
		contentBlockStoreAPI.cBlockStoreAPI,
		versionStorageAPI.cStorageAPI,
		hashAPI.cHashAPI,
		jobAPI.cJobAPI,
		cProgressAPI,
		nil,
		nil,
		contentIndex.cContentIndex,
		sourceVersionIndex.cVersionIndex,
		targetVersionIndex.cVersionIndex,
		versionDiff.cVersionDiff,
		cVersionFolderPath,
		cRetainPermissions)
	if errno != 0 {
		return int(errno)
	}
	return 0
}

//export LogProxy_Log
func LogProxy_Log(log_context *C.struct_Longtail_LogContext, log *C.char) {
	logger := RestorePointer(log_context.context).(Logger)
	logFields := make([]LogField, log_context.field_count)
	for i := 0; i < int(log_context.field_count); i++ {
		logField := C.GetLogField(log_context, C.int(i))
		logFields[i].Name = C.GoString(logField.name)
		logFields[i].Value = C.GoString(logField.value)
	}
	logger.OnLog(C.GoString(log_context.file), C.GoString(log_context.function), int(log_context.line), int(log_context.level), logFields, C.GoString(log))
}

//export BlockStoreAPIProxy_Dispose
func BlockStoreAPIProxy_Dispose(api *C.struct_Longtail_API) {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	blockStore.Close()
	UnrefPointer(context)
	C.Longtail_Free(unsafe.Pointer(api))
}

//export BlockStoreAPIProxy_PutStoredBlock
func BlockStoreAPIProxy_PutStoredBlock(api *C.struct_Longtail_BlockStoreAPI, storedBlock *C.struct_Longtail_StoredBlock, async_complete_api *C.struct_Longtail_AsyncPutStoredBlockAPI) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	errno := blockStore.PutStoredBlock(Longtail_StoredBlock{cStoredBlock: storedBlock}, Longtail_AsyncPutStoredBlockAPI{cAsyncCompleteAPI: async_complete_api})
	return C.int(errno)
}

//export BlockStoreAPIProxy_GetStoredBlock
func BlockStoreAPIProxy_GetStoredBlock(api *C.struct_Longtail_BlockStoreAPI, blockHash C.uint64_t, async_complete_api *C.struct_Longtail_AsyncGetStoredBlockAPI) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	errno := blockStore.GetStoredBlock(uint64(blockHash), Longtail_AsyncGetStoredBlockAPI{cAsyncCompleteAPI: async_complete_api})
	return C.int(errno)
}

//export BlockStoreAPIProxy_PreflightGet
func BlockStoreAPIProxy_PreflightGet(api *C.struct_Longtail_BlockStoreAPI, chunk_count C.uint64_t, chunk_hashes *C.TLongtail_Hash) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	chunkCount := int(chunk_count)
	chunkHashes := carray2slice64(chunk_hashes, chunkCount)
	copyChunkHashes := append([]uint64{}, chunkHashes...)
	errno := blockStore.PreflightGet(copyChunkHashes)
	return C.int(errno)
}

//export BlockStoreAPIProxy_GetExistingContent
func BlockStoreAPIProxy_GetExistingContent(api *C.struct_Longtail_BlockStoreAPI, chunk_count C.uint64_t, chunk_hashes *C.TLongtail_Hash, min_block_usage_percent C.uint32_t, async_complete_api *C.struct_Longtail_AsyncGetExistingContentAPI) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	chunkCount := int(chunk_count)
	minBlockUsagePercent := uint32(min_block_usage_percent)
	chunkHashes := carray2slice64(chunk_hashes, chunkCount)
	copyChunkHashes := append([]uint64{}, chunkHashes...)
	errno := blockStore.GetExistingContent(
		copyChunkHashes,
		minBlockUsagePercent,
		Longtail_AsyncGetExistingContentAPI{cAsyncCompleteAPI: async_complete_api})
	return C.int(errno)
}

//export BlockStoreAPIProxy_GetStats
func BlockStoreAPIProxy_GetStats(api *C.struct_Longtail_BlockStoreAPI, out_stats *C.struct_Longtail_BlockStore_Stats) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	stats, errno := blockStore.GetStats()
	if errno == 0 {
		for s := 0; s < Longtail_BlockStoreAPI_StatU64_Count; s++ {
			out_stats.m_StatU64[s] = C.uint64_t(stats.StatU64[s])
		}
	}
	return C.int(errno)
}

//export BlockStoreAPIProxy_Flush
func BlockStoreAPIProxy_Flush(api *C.struct_Longtail_BlockStoreAPI, async_complete_api *C.struct_Longtail_AsyncFlushAPI) C.int {
	context := C.BlockStoreAPIProxy_GetContext(unsafe.Pointer(api))
	blockStore := RestorePointer(context).(BlockStoreAPI)
	errno := blockStore.Flush(Longtail_AsyncFlushAPI{cAsyncCompleteAPI: async_complete_api})
	return C.int(errno)
}

func CreateBlockStoreAPI(blockStore BlockStoreAPI) Longtail_BlockStoreAPI {
	cContext := SavePointer(blockStore)
	blockStoreAPIProxy := C.CreateBlockStoreProxyAPI(cContext)
	return Longtail_BlockStoreAPI{cBlockStoreAPI: blockStoreAPIProxy}
}

func getLoggerFunc(logger Logger) C.Longtail_Log {
	if logger == nil {
		return nil
	}
	return C.Longtail_Log(C.LogProxy_Log)
}

//SetLogger ...
func SetLogger(logger Logger) {
	cLoggerContext := SavePointer(logger)
	C.Longtail_SetLog(getLoggerFunc(logger), cLoggerContext)
}

//SetLogLevel ...
func SetLogLevel(level int) {
	C.Longtail_SetLogLevel(C.int(level))
}

func getAssertFunc(assert Assert) C.Longtail_Assert {
	if assert == nil {
		return nil
	}
	return C.Longtail_Assert(C.AssertProxy_Assert)
}

var activeAssert Assert

//SetAssert ...
func SetAssert(assert Assert) {
	C.Longtail_SetAssert(getAssertFunc(assert))
	activeAssert = assert
}

//export AssertProxy_Assert
func AssertProxy_Assert(expression *C.char, file *C.char, line C.int) {
	if activeAssert != nil {
		activeAssert.OnAssert(C.GoString(expression), C.GoString(file), int(line))
	}
}

//EnableMemtrace ...
func EnableMemtrace() {
	C.EnableMemtrace()
}

//EnableMemtrace ...
func DisableMemtrace() {
	C.DisableMemtrace()
}
