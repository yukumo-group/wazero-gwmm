package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero/api"
)

// StoredData defines the data that will be stored in the memory of wasm
type StoredData struct {
	startPos int32
	length   int32
}

// StoredDataManager defines all the datas that will be stored in the memory of wasm
type StoredDataManager struct {
	lock  sync.RWMutex
	datas map[string]*StoredData
}

// NewStoredDataManager creates new manager for storing stored data in wasm memory
func NewStoredDataManager() *StoredDataManager {
	return &StoredDataManager{
		datas: map[string]*StoredData{},
	}
}

// addData adds data directly
func (manager *StoredDataManager) addData(
	startPos int32,
	length int32,
	dataName string,
) {
	manager.datas[dataName] = &StoredData{
		startPos: max(0, startPos),
		length:   length,
	}
}

// getData gets data directly
func (manager *StoredDataManager) getData(
	dataName string,
) (*StoredData, bool) {
	result, exists := manager.datas[dataName]
	return result, exists
}

// getHighestPos gets the highest pos occupied
func (manager *StoredDataManager) getHighestPos() int32 {
	var highestPos int32 = 0
	for _, data := range manager.datas {
		highestPos = max(highestPos, (data.startPos + data.length))
	}
	return highestPos
}

// getNextPos gets the next pos
func (manager *StoredDataManager) getNextPos() int32 {
	highestPos := manager.getHighestPos()
	nextPos := ((highestPos + 16 + 3) / 4) * 4
	return nextPos
}

// isPosAvailable checks if certain position is available
func (manager *StoredDataManager) isPosAvailable(
	dataCheck *StoredData,
) bool {
	dataCheckEndPos := dataCheck.startPos + dataCheck.length
	for _, data := range manager.datas {
		dataEndPos := data.startPos + data.length
		// Cover
		if dataCheck.startPos >= data.startPos && dataCheckEndPos <= dataEndPos {
			return false
		}
		if dataCheck.startPos <= data.startPos && dataCheckEndPos >= dataEndPos {
			return false
		}
		// Intersect
		if dataCheck.startPos <= data.startPos && dataCheckEndPos >= data.startPos {
			return false
		}
		if dataCheck.startPos <= dataEndPos && dataCheckEndPos >= dataEndPos {
			return false
		}
	}
	return true
}

// free frees data stored
func (manager *StoredDataManager) free(
	dataName string,
) {
	delete(
		manager.datas,
		dataName,
	)
}

// GetDataDirectly gets the data stored struct
func (manager *StoredDataManager) GetDataDirectly(
	dataName string,
) (*StoredData, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	result, exists := manager.getData(
		dataName,
	)
	return result, exists
}

// AddDataDirectly directly (passin length and start position) adds new data for storing in wasm
func (manager *StoredDataManager) AddDataDirectly(
	startPos int32,
	length int32,
	dataName string,
) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	isPosAvailable := manager.isPosAvailable(
		&StoredData{
			startPos: max(0, startPos),
			length:   length,
		},
	)
	if !isPosAvailable {
		return fmt.Errorf(
			"position from %d to %d is already occupied",
			startPos,
			startPos+length,
		)
	}
	manager.addData(
		startPos,
		length,
		dataName,
	)
	return nil
}

// AddData adds data with check
func (manager *StoredDataManager) AddData(
	dataName string,
	data []byte,
	memoryAPI api.Memory,
) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	startPos := manager.getNextPos()
	length := int32(len(data))
	ok := memoryAPI.Write(
		uint32(startPos),
		data,
	)
	if !ok {
		return fmt.Errorf(
			"",
		)
	}
	manager.addData(
		startPos,
		length,
		dataName,
	)
	return nil
}

// GetData directly gets data
func (manager *StoredDataManager) GetData(
	dataName string,
	memoryApi api.Memory,
) ([]byte, error) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	storedData, exists := manager.getData(dataName)
	if !exists {
		return nil, fmt.Errorf(
			"the data with name %s does not exists",
			dataName,
		)
	}
	result, ok := memoryApi.Read(
		uint32(storedData.startPos),
		uint32(storedData.length),
	)
	if !ok {
		return nil, errors.New(
			"failed to read data",
		)
	}
	return result, nil
}

// Free frees data stored
func (manager *StoredDataManager) Free(
	dataName string,
) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.free(dataName)
}

// FreeAll frees all the data
func (manager *StoredDataManager) FreeAll() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	for dataName := range manager.datas {
		manager.free(
			dataName,
		)
	}
}
