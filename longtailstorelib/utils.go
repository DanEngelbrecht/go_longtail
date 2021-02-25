package longtailstorelib

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/DanEngelbrecht/golongtail/longtaillib"
	"github.com/pkg/errors"
)

const (
	publicStoreIndexName      = "store.lsi"
	privateTempStoreIndexPath = "index/"
)

func readBlobWithRetry(
	client BlobClient,
	key string) ([]byte, int, error) {
	retryCount := 0
	objHandle, err := client.NewObject(key)
	if err != nil {
		return nil, retryCount, err
	}
	exists, err := objHandle.Exists()
	if err != nil {
		return nil, retryCount, err
	}
	if !exists {
		return nil, retryCount, longtaillib.ErrENOENT
	}
	blobData, err := objHandle.Read()
	if err != nil {
		log.Printf("Retrying getBlob %s using %s\n", key, client.String())
		retryCount++
		blobData, err = objHandle.Read()
	} else if blobData == nil {
		return nil, retryCount, longtaillib.ErrENOENT
	}
	if err != nil {
		log.Printf("Retrying 500 ms delayed getBlob %s using %s\n", key, client.String())
		time.Sleep(500 * time.Millisecond)
		retryCount++
		blobData, err = objHandle.Read()
	} else if blobData == nil {
		return nil, retryCount, longtaillib.ErrENOENT
	}
	if err != nil {
		log.Printf("Retrying 2 s delayed getBlob %s using %s\n", key, client.String())
		time.Sleep(2 * time.Second)
		retryCount++
		blobData, err = objHandle.Read()
	} else if blobData == nil {
		return nil, retryCount, longtaillib.ErrENOENT
	}

	if err != nil {
		return nil, retryCount, err
	}

	return blobData, retryCount, nil
}

func writeBlobWithRetry(
	client BlobClient,
	key string,
	forceWrite bool,
	blob []byte) (int, error) {

	retryCount := 0
	objHandle, err := client.NewObject(key)
	if err != nil {
		return retryCount, err
	}
	write := forceWrite
	if !forceWrite {
		exists, err := objHandle.Exists()
		if err != nil {
			return retryCount, err
		}
		write = !exists
	}
	if !write {
		return 0, nil
	}

	_, err = objHandle.Write(blob)
	if err != nil {
		log.Printf("Retrying putBlob %s in store %s\n", key, client.String())
		retryCount++
		_, err = objHandle.Write(blob)
	}
	if err != nil {
		log.Printf("Retrying 500 ms delayed putBlob %s in store %s\n", key, client.String())
		time.Sleep(500 * time.Millisecond)
		retryCount++
		_, err = objHandle.Write(blob)
	}
	if err != nil {
		log.Printf("Retrying 2 s delayed putBlob %s in store %s\n", key, client.String())
		time.Sleep(2 * time.Second)
		retryCount++
		_, err = objHandle.Write(blob)
	}

	if err != nil {
		return retryCount, err
	}

	return retryCount, nil
}

func readStoreIndex(
	client BlobClient) (longtaillib.Longtail_StoreIndex, error) {

	var storeIndex longtaillib.Longtail_StoreIndex
	storeIndexBuffer, _, err := readBlobWithRetry(client, publicStoreIndexName)
	if err != nil {
		if !errors.Is(err, longtaillib.ErrENOENT) {
			return longtaillib.Longtail_StoreIndex{}, err
		}
		errno := 0
		storeIndex, errno = longtaillib.CreateStoreIndexFromBlocks([]longtaillib.Longtail_BlockIndex{})
		if errno != 0 {
			return longtaillib.Longtail_StoreIndex{}, errors.Wrapf(longtaillib.ErrnoToError(errno, longtaillib.ErrEIO), "readStoreStoreIndex: longtaillib.CreateStoreIndexFromBlocks() failed")
		}
	} else {
		errno := 0
		storeIndex, errno = longtaillib.ReadStoreIndexFromBuffer(storeIndexBuffer)
		if errno != 0 {
			return longtaillib.Longtail_StoreIndex{}, errors.Wrapf(longtaillib.ErrnoToError(errno, longtaillib.ErrEIO), "readStoreStoreIndex: longtaillib.CreateStoreIndexFromBlocks() failed")
		}
	}
	return storeIndex, nil
}

func updateStoreIndex(
	storeIndex longtaillib.Longtail_StoreIndex,
	addedBlockIndexes []longtaillib.Longtail_BlockIndex) (longtaillib.Longtail_StoreIndex, error) {
	addedStoreIndex, errno := longtaillib.CreateStoreIndexFromBlocks(addedBlockIndexes)
	if errno != 0 {
		return longtaillib.Longtail_StoreIndex{}, errors.Wrap(longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM), "contentIndexWorker: longtaillib.CreateStoreIndexFromBlocks() failed")
	}

	updatedStoreIndex, errno := longtaillib.MergeStoreIndex(storeIndex, addedStoreIndex)
	addedStoreIndex.Dispose()
	if errno != 0 {
		updatedStoreIndex.Dispose()
		return longtaillib.Longtail_StoreIndex{}, errors.Wrap(longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM), "contentIndexWorker: longtaillib.MergeStoreIndex() failed")
	}
	return updatedStoreIndex, nil
}

func getPartialStoreIndexName(storeIndex longtaillib.Longtail_StoreIndex) string {
	blockHashes := storeIndex.GetBlockHashes()
	sort.Slice(blockHashes, func(i, j int) bool { return blockHashes[i] < blockHashes[j] })
	buf := make([]byte, len(blockHashes)*8)
	for i, h := range blockHashes {
		buf[i*8+0] = byte(h & 255)
		buf[i*8+1] = byte((h >> 8) & 255)
		buf[i*8+2] = byte((h >> 16) & 255)
		buf[i*8+3] = byte((h >> 24) & 255)
		buf[i*8+4] = byte((h >> 32) & 255)
		buf[i*8+5] = byte((h >> 40) & 255)
		buf[i*8+6] = byte((h >> 48) & 255)
		buf[i*8+7] = byte((h >> 56) & 255)
	}
	storeIndexSha := sha1.Sum(buf)
	storeIndexName := fmt.Sprintf("%s%x.lsi", privateTempStoreIndexPath, storeIndexSha)
	return storeIndexName
}

func writeStoreIndexBlob(
	client BlobClient,
	storeIndex longtaillib.Longtail_StoreIndex,
	storeIndexName string,
	force bool) (bool, error) {

	storeIndexBlob, errno := longtaillib.WriteStoreIndexToBuffer(storeIndex)
	if errno != 0 {
		log.Printf("writeStoreIndexBlob: longtaillib.WriteStoreIndexToBuffer() for %s failed with %v\n", storeIndexName, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return false, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}
	_, err := writeBlobWithRetry(client, storeIndexName, force, storeIndexBlob)
	if err != nil {
		log.Printf("writeStoreIndexBlob: writeBlobWithRetry(%s) failed with %v\n", storeIndexName, err)
		return false, err
	}
	return false, nil
}

func readStoreIndexBlob(
	client BlobClient,
	key string) (longtaillib.Longtail_StoreIndex, error) {
	storeIndexBuffer, _, err := readBlobWithRetry(client, key)
	if err != nil {
		if err != longtaillib.ErrENOENT {
			log.Printf("readStoreIndexBlob: readBlobWithRetry(%s) failed with %v\n", key, err)
		}
		return longtaillib.Longtail_StoreIndex{}, err
	}
	storeIndex, errno := longtaillib.ReadStoreIndexFromBuffer(storeIndexBuffer)
	if errno != 0 {
		log.Printf("readStoreIndexBlob: ReadStoreIndexFromBuffer() for %s failed with %v\n", key, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return longtaillib.Longtail_StoreIndex{}, errors.Wrapf(longtaillib.ErrnoToError(errno, longtaillib.ErrEIO), "readStoreIndexBlob: ReadStoreIndexFromBuffer() failed")
	}
	return storeIndex, nil
}

func copyStoreIndex(storeIndex longtaillib.Longtail_StoreIndex) (longtaillib.Longtail_StoreIndex, error) {
	storeIndexBlob, errno := longtaillib.WriteStoreIndexToBuffer(storeIndex)
	if errno != 0 {
		log.Printf("copyStoreIndex: longtaillib.WriteStoreIndexToBuffer() failed with %v\n", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return longtaillib.Longtail_StoreIndex{}, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}
	copyIndex, errno := longtaillib.ReadStoreIndexFromBuffer(storeIndexBlob)
	if errno != 0 {
		log.Printf("copyStoreIndex: longtaillib.ReadStoreIndexFromBuffer() failed with %v\n", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return longtaillib.Longtail_StoreIndex{}, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}
	return copyIndex, nil
}

func scanPartialIndexNames(
	client BlobClient) ([]string, error) {
	partialIndexBlobs, err := client.GetObjects(privateTempStoreIndexPath)
	if err != nil {
		log.Printf("scanPartialIndexNames: client.GetObjects(%s) failed with %v\n", privateTempStoreIndexPath, err)
		return []string{}, err
	}
	newIndexNames := []string{}
	for _, blob := range partialIndexBlobs {
		newIndexNames = append(newIndexNames, blob.Name)
	}
	return newIndexNames, nil
}

func storeIndexContains(
	storeIndex longtaillib.Longtail_StoreIndex,
	addedStoreIndex longtaillib.Longtail_StoreIndex) bool {
	lookup := map[uint64]bool{}
	for _, b := range storeIndex.GetBlockHashes() {
		lookup[b] = true
	}
	for _, b := range addedStoreIndex.GetBlockHashes() {
		if _, exists := lookup[b]; !exists {
			return false
		}
	}
	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func mergeSwapStoreIndex(
	a longtaillib.Longtail_StoreIndex,
	b longtaillib.Longtail_StoreIndex) (longtaillib.Longtail_StoreIndex, error) {

	c, errno := longtaillib.MergeStoreIndex(a, b)
	if errno != 0 {
		log.Printf("mergeSwapStoreIndex: longtaillib.MergeStoreIndex failed with %v\n", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return a, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}
	a.Dispose()
	return c, nil
}

func tryDelete(
	client BlobClient,
	key string) error {
	objHandle, err := client.NewObject(key)
	if err != nil {
		log.Printf("tryDelete: client.NewObject(%s) with %v\n", key, err)
		return err
	}
	_ = objHandle.Delete()
	return nil
}

func consolidatePartialIndexes(
	client BlobClient) (longtaillib.Longtail_StoreIndex, []string, error) {
	storeIndex, errno := longtaillib.CreateStoreIndexFromBlocks([]longtaillib.Longtail_BlockIndex{})
	if errno != 0 {
		log.Printf("writeStoreIndex: CreateStoreIndexFromBlocks failed with %v\n", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return longtaillib.Longtail_StoreIndex{}, []string{}, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}
	newIndexNames, err := scanPartialIndexNames(client)
	if err != nil {
		storeIndex.Dispose()
		log.Printf("writeStoreIndex: scanPartialIndexNames failed with %v\n", err)
		return longtaillib.Longtail_StoreIndex{}, []string{}, err
	}
	log.Printf("writeStoreIndex: found %d indexes\n", len(newIndexNames))

	consolidatedNames := []string{}
	for _, name := range newIndexNames {
		partialStoreIndex, err := readStoreIndexBlob(client, name)
		if err != nil {
			log.Printf("writeStoreIndex: failed reading %s\n", name)
			continue
		}
		storeIndex, err = mergeSwapStoreIndex(storeIndex, partialStoreIndex)
		partialStoreIndex.Dispose()
		if err != nil {
			storeIndex.Dispose()
			log.Printf("writeStoreIndex: mergeSwapStoreIndex for %s failed with %v\n", name, err)
			return longtaillib.Longtail_StoreIndex{}, []string{}, err
		}
		consolidatedNames = append(consolidatedNames, name)
	}
	if len(consolidatedNames) > 0 {
		log.Printf("writeStoreIndex: consolidated %d indexes into %s\n", len(consolidatedNames), getPartialStoreIndexName(storeIndex))
	}
	return storeIndex, consolidatedNames, nil
}

func writeStoreIndex(
	client BlobClient,
	addedStoreIndex longtaillib.Longtail_StoreIndex) error {

	addedIndexStoreName := getPartialStoreIndexName(addedStoreIndex)
	_, err := writeStoreIndexBlob(client, addedStoreIndex, addedIndexStoreName, false)
	if err != nil {
		log.Printf("writeStoreIndex: writePartialStoreIndex failed with %v\n", err)
		return err
	}
	log.Printf("writeStoreIndex: wrote added index into %s\n", addedIndexStoreName)

	for {
		consolidatedStoreIndex, newConsolidatedNames, err := consolidatePartialIndexes(client)
		consolidatedStoreIndexName := getPartialStoreIndexName(consolidatedStoreIndex)
		_, err = writeStoreIndexBlob(client, consolidatedStoreIndex, consolidatedStoreIndexName, false)
		if err != nil {
			log.Printf("writeStoreIndex: writeStoreIndexBlob(%s) failed with %v\n", consolidatedStoreIndexName, err)
			return err
		}
		if len(newConsolidatedNames) > 1 {
			log.Printf("writeStoreIndex: wrote %d consolidated indexes into %s\n", len(newConsolidatedNames), consolidatedStoreIndexName)
		}

		storeIndex, err := readStoreIndexBlob(client, publicStoreIndexName)
		if err != nil {
			errno := 0
			storeIndex, errno = longtaillib.CreateStoreIndexFromBlocks([]longtaillib.Longtail_BlockIndex{})
			if errno != 0 {
				log.Printf("writeStoreIndex: CreateStoreIndexFromBlocks failed with %v\n", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
				return longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
			}
		}
		oldStoreIndexName := getPartialStoreIndexName(storeIndex)

		storeIndex, err = mergeSwapStoreIndex(storeIndex, consolidatedStoreIndex)
		consolidatedStoreIndex.Dispose()
		if err != nil {
			log.Printf("writeStoreIndex: mergeSwapStoreIndex(%s) failed with %v\n", publicStoreIndexName, err)
			return err
		}
		fullStoreIndexName := getPartialStoreIndexName(storeIndex)

		if oldStoreIndexName != fullStoreIndexName {
			_, err = writeStoreIndexBlob(client, storeIndex, publicStoreIndexName, true)
			if err != nil {
				storeIndex.Dispose()
				log.Printf("writeStoreIndex: writeStoreIndexBlob(%s) failed with %v\n", consolidatedStoreIndexName, err)
				return err
			}
			log.Printf("writeStoreIndex: updated `store.lsi` to %s\n", consolidatedStoreIndexName)
			storeIndex.Dispose()
			continue
		}

		storeIndex.Dispose()
		for _, name := range newConsolidatedNames {
			if name == consolidatedStoreIndexName {
				continue
			}
			err = tryDelete(client, name)
			if err != nil {
				log.Printf("writeStoreIndex: client.NewObject with %v\n", err)
				return err
			}
		}

		if len(newConsolidatedNames) == 1 {
			log.Printf("writeStoreIndex: completed on %s\n", fullStoreIndexName)
			return nil
		}
	}
}

func getStoreIndexFromBlocks(
	ctx context.Context,
	s *remoteStore,
	blobClient BlobClient,
	blockKeys []string) (longtaillib.Longtail_StoreIndex, error) {

	storeIndex, errno := longtaillib.CreateStoreIndexFromBlocks([]longtaillib.Longtail_BlockIndex{})
	if errno != 0 {
		return longtaillib.Longtail_StoreIndex{}, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
	}

	batchCount := runtime.NumCPU()
	batchStart := 0

	if batchCount > len(blockKeys) {
		batchCount = len(blockKeys)
	}
	clients := make([]BlobClient, batchCount)
	for c := 0; c < batchCount; c++ {
		client, err := s.blobStore.NewClient(ctx)
		if err != nil {
			storeIndex.Dispose()
			return longtaillib.Longtail_StoreIndex{}, err
		}
		clients[c] = client
	}

	var wg sync.WaitGroup

	for batchStart < len(blockKeys) {
		batchLength := batchCount
		if batchStart+batchLength > len(blockKeys) {
			batchLength = len(blockKeys) - batchStart
		}
		batchBlockIndexes := make([]longtaillib.Longtail_BlockIndex, batchLength)
		wg.Add(batchLength)
		for batchPos := 0; batchPos < batchLength; batchPos++ {
			i := batchStart + batchPos
			blockKey := blockKeys[i]
			go func(client BlobClient, batchPos int, blockKey string) {
				storedBlockData, _, err := readBlobWithRetry(
					client,
					blockKey)

				if err != nil {
					wg.Done()
					return
				}

				blockIndex, errno := longtaillib.ReadBlockIndexFromBuffer(storedBlockData)
				if errno != 0 {
					wg.Done()
					return
				}

				blockPath := GetBlockPath("chunks", blockIndex.GetBlockHash())
				if blockPath == blockKey {
					batchBlockIndexes[batchPos] = blockIndex
				} else {
					log.Printf("Block %s name does not match content hash, expected name %s\n", blockKey, blockPath)
				}

				wg.Done()
			}(clients[batchPos], batchPos, blockKey)
		}
		wg.Wait()
		writeIndex := 0
		for i, blockIndex := range batchBlockIndexes {
			if !blockIndex.IsValid() {
				continue
			}
			if i > writeIndex {
				batchBlockIndexes[writeIndex] = blockIndex
			}
			writeIndex++
		}
		batchBlockIndexes = batchBlockIndexes[:writeIndex]
		batchStoreIndex, errno := longtaillib.CreateStoreIndexFromBlocks(batchBlockIndexes)
		for _, blockIndex := range batchBlockIndexes {
			blockIndex.Dispose()
		}
		if errno != 0 {
			storeIndex.Dispose()
			return longtaillib.Longtail_StoreIndex{}, longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM)
		}
		newStoreIndex, errno := longtaillib.MergeStoreIndex(storeIndex, batchStoreIndex)
		batchStoreIndex.Dispose()
		storeIndex.Dispose()
		storeIndex = newStoreIndex
		batchStart += batchLength
		log.Printf("Scanned %d/%d blocks in %s\n", batchStart, len(blockKeys), blobClient.String())
	}

	for c := 0; c < batchCount; c++ {
		clients[c].Close()
	}

	return storeIndex, nil
}
