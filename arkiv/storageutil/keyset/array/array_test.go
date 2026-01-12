package array_test

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/arkiv/storageutil/keyset/array"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestEmptyArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))

	size := array.Size()
	require.Equal(t, uint256.NewInt(0), size)
}

func TestAppendToEmptyArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))

	v := common.HexToHash("0xa")

	array.Append(v)

	got, err := array.Get(uint256.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, v, got)

	require.Equal(t, uint256.NewInt(1), array.Size())
}

func TestAppendToNonEmptyArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))
	array.Append(common.HexToHash("0xb"))

	got, err := array.Get(uint256.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xa"), got)

	got, err = array.Get(uint256.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xb"), got)

	require.Equal(t, uint256.NewInt(2), array.Size())
}

func TestRemoveLastFromNonEmptyArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))
	array.Append(common.HexToHash("0xb"))

	err := array.RemoveLast()
	require.NoError(t, err)

	got, err := array.Get(uint256.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xa"), got)

	require.Equal(t, uint256.NewInt(1), array.Size())

}

func TestRemoveLastFromArrayWithOneElement(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))

	err := array.RemoveLast()
	require.NoError(t, err)

	require.Equal(t, uint256.NewInt(0), array.Size())
}

func TestSetElementForOneElementArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))

	err := array.Set(uint256.NewInt(0), common.HexToHash("0xb"))
	require.NoError(t, err)

	got, err := array.Get(uint256.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xb"), got)
}

func TestSetElementForNonEmptyArray(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))
	array.Append(common.HexToHash("0xb"))

	err := array.Set(uint256.NewInt(1), common.HexToHash("0xc"))
	require.NoError(t, err)

	got, err := array.Get(uint256.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xc"), got)

	got, err = array.Get(uint256.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xa"), got)

}

// mockStateAccess implements StateAccess interface for testing
type mockStateAccess struct {
	storage map[common.Address]map[common.Hash]common.Hash
}

func newMockStateAccess() *mockStateAccess {
	return &mockStateAccess{
		storage: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func (m *mockStateAccess) GetState(addr common.Address, key common.Hash) common.Hash {
	if _, exists := m.storage[addr]; !exists {
		return common.Hash{}
	}
	if val, exists := m.storage[addr][key]; exists {
		return val
	}
	return common.Hash{}
}

func (m *mockStateAccess) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	zeroHash := common.Hash{}

	// If value is zero, delete the entry instead of storing it
	if value == zeroHash {
		if storageMap, exists := m.storage[addr]; exists {
			delete(storageMap, key)

			// If address map is now empty, delete it too
			if len(storageMap) == 0 {
				delete(m.storage, addr)
			}
		}
		return zeroHash
	}

	// Otherwise store the non-zero value
	if _, exists := m.storage[addr]; !exists {
		m.storage[addr] = make(map[common.Hash]common.Hash)
	}
	m.storage[addr][key] = value
	return value
}

// Helper method to get the number of entries in storage for testing
func (m *mockStateAccess) GetStorageEntryCount(addr common.Address) int {
	if storageMap, exists := m.storage[addr]; exists {
		return len(storageMap)
	}
	return 0
}

func (m *mockStateAccess) Print(addr common.Address) {
	keys := []common.Hash{}

	for key := range m.storage[addr] {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Big().Cmp(keys[j].Big()) < 0
	})

	for _, key := range keys {
		value := m.storage[addr][key]
		fmt.Printf("%s: %s\n", key.Hex(), value.Hex())

	}
}

func TestIterate(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))
	array.Append(common.HexToHash("0xb"))

	values := slices.Collect(array.Iterate)

	require.Equal(t, []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xb")}, values)
}

func TestClear(t *testing.T) {
	db := newMockStateAccess()
	array := array.NewArray(db, common.HexToHash("0xabc"))
	array.Append(common.HexToHash("0xa"))
	array.Append(common.HexToHash("0xb"))

	array.Clear()

	require.Equal(t, uint256.NewInt(0), array.Size())
}
