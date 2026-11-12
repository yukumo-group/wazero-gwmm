package storage

import (
	"testing"
)

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
}
