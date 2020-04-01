@echo off

reg Query "HKLM\Hardware\Description\System\CentralProcessor\0" | find /i "x86" > NUL && set ARCH_NAME=386 || set ARCH_NAME=amd64
set OS_NAME=windows

set LIB_TARGET_FOLDER=%OS_NAME%_%ARCH_NAME%
set LIB_TARGET=%LIB_TARGET_FOLDER%/longtail_lib.a

echo Building %LIB_TARGET%

if not exist obj mkdir obj
if not exist "%LIB_TARGET_FOLDER%" mkdir "%LIB_TARGET_FOLDER%"
pushd obj
set BIKESHED_SRC=..\lib\bikeshed\*.c
set BLAKE2_SRC=..\lib\blake2\*.c ..\lib\blake2\ext\*.c
set BLAKE3_SRC=..\lib\blake3\*.c ..\lib\blake3\ext\*.c
set CACHEBLOCKSTORE_SRC=..\lib\cacheblockstore\*.c
set COMPRESSBLOCKSTORE_SRC=..\lib\compressblockstore\*.c
set FULL_COMPRESSION_REGISTRY_SRC=..\lib\compressionregistry\longtail_full_compression_registry.c
set FILESTORAGE_SRC=..\lib\filestorage\*.c
set FSBLOCKSTORE_SRC=..\lib\fsblockstore\*.c
set MEMSTORAGE_SRC=..\lib\memstorage\*.c
set MEOWHASH_SRC=..\lib\meowhash\*.c
set LIZARD_SRC=..\lib\lizard\*.c ..\lib\lizard\ext\*.c ..\lib\lizard\ext\entropy\*.c
set LZ4_SRC=..\lib\lz4\*.c ..\lib\lz4\ext\*.c
set BROTLI_SRC=..\lib\brotli\*.c ..\lib\brotli\ext\common\*.c ..\lib\brotli\ext\dec\*.c ..\lib\brotli\ext\enc\*.c ..\lib\brotli\ext\fuzz\*.c
set ZSTD_SRC=..\lib\zstd\*.c ..\lib\zstd\ext\common\*.c ..\lib\zstd\ext\compress\*.c ..\lib\zstd\ext\decompress\*.c
del /Q *.o
gcc -c -std=gnu99 -g -m64 -pthread -msse4.1 -maes -DWINVER=0x0A00 -D_WIN32_WINNT=0x0A00 -DLONGTAIL_ASSERTS ..\src\*.c ..\src\ext\*.c ..\lib\*.c %BIKESHED_SRC% %BLAKE2_SRC% %BLAKE3_SRC% %CACHEBLOCKSTORE_SRC% %COMPRESSBLOCKSTORE_SRC% %FULL_COMPRESSION_REGISTRY_SRC% %FILESTORAGE_SRC% %FSBLOCKSTORE_SRC% %MEMSTORAGE_SRC% %MEOWHASH_SRC% %LIZARD_SRC% %LZ4_SRC% %BROTLI_SRC% %ZSTD_SRC%
popd
ar rc %LIB_TARGET% obj/*.o