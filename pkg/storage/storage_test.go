package storage

import (
	"context"
	"os"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// loadWASM loads wasm file
func loadWASM(
	ctx context.Context,
	r wazero.Runtime,
	wasmFilePath string,
) (api.Module, error) {
	wasmData, err := os.ReadFile(wasmFilePath)
	if err != nil {
		return nil, nil
	}
	return r.Instantiate(
		ctx,
		wasmData,
	)
}

// assertBytesSame tests if the bytes are same
func assertBytesSame(
	t *testing.T,
	byteData1 []byte,
	byteData2 []byte,
) {
	len1 := len(byteData1)
	len2 := len(byteData2)
	if len1 != len2 {
		t.Fatalf(
			"the two datas doesn't have the same length: 1: %d, 2: %d",
			len1,
			len2,
		)
	}
	testedLen := min(len1, len2)
	for i := range testedLen {
		if byteData1[i] != byteData2[i] {
			t.Fatalf(
				"on %d, the data should be %d, got %d",
				i,
				byteData1[i],
				byteData2[i],
			)
		}
	}
}

// TestGetNextPos tests getting next position
func TestGetNextPos(t *testing.T) {
	t.Parallel()
	newManager := NewStoredDataManager()
	nextPos1 := newManager.getNextPos()
	if nextPos1%4 != 0 {
		t.Fatalf(
			"the next pos 1 fetched %d cannot get divided by 4",
			nextPos1,
		)
	}
	err := newManager.AddDataDirectly(
		nextPos1,
		65,
		"114514",
	)
	if err != nil {
		t.Fatal(err)
	}
	nextPos2 := newManager.getNextPos()
	if nextPos2%4 != 0 {
		t.Fatalf(
			"the next pos 2 fetched %d cannot get divided by 4",
			nextPos2,
		)
	}
	if nextPos2 == nextPos1 {
		t.Fatalf(
			"position 1 and 2 should not be same: %d, %d",
			nextPos1,
			nextPos2,
		)
	}
	err = newManager.AddDataDirectly(
		nextPos2,
		114514,
		"1919810",
	)
	nextPos3 := newManager.getNextPos()
	if nextPos3%4 != 0 {
		t.Fatalf(
			"the next pos 2 fetched %d cannot get divided by 4",
			nextPos3,
		)
	}
	if nextPos3 == nextPos1 {
		t.Fatalf(
			"position 1 and 3 should not be same: %d, %d",
			nextPos1,
			nextPos3,
		)
	}
}

// TestDataCRUD tests data crud
func TestDataCRUD(t *testing.T) {
	t.Parallel()
	const wasmFilePath string = "testdata/add.wasm"
	const testDataPath string = "testdata/README.md"
	const dataName string = "testData"
	// Prepare data
	ctx := t.Context()
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(
		ctx,
		r,
	)
	wasmModule, err := loadWASM(
		ctx,
		r,
		wasmFilePath,
	)
	if err != nil {
		t.Fatal(err)
	}
	memory := wasmModule.Memory()
	if memory == nil {
		t.Fatal(
			"the wasm file does not supports memory export",
		)
	}
	testData, err := os.ReadFile(
		testDataPath,
	)
	if err != nil {
		t.Fatal(err)
	}
	// Test the manager
	testManager := NewStoredDataManager()
	err = testManager.AddData(
		dataName,
		testData,
		memory,
	)
	if err != nil {
		t.Fatal(err)
	}
	fetchedData, err := testManager.GetData(
		dataName,
		memory,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertBytesSame(
		t,
		testData,
		fetchedData,
	)
	// Close the module and runtime
	err = wasmModule.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
