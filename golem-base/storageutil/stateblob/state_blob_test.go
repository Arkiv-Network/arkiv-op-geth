package stateblob_test

import (
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
	"github.com/stretchr/testify/require"
)

func TestPacker(t *testing.T) {

	t.Run("short slice", func(t *testing.T) {

		hashes := stateblob.BytesTo32ByteSequence([]byte("hello"))

		// require.Equal(t, len(hashes), 1)
		require.Equal(
			t,
			[]common.Hash{common.HexToHash("0x68656c6c6f00000000000000000000000000000000000000000000000000000a")},
			slices.Collect(hashes),
		)
	})

	t.Run("long slice", func(t *testing.T) {
		hashes := stateblob.BytesTo32ByteSequence([]byte("lorem ipsum dolor sit amet consectetur adipiscing elit"))

		require.Equal(
			t,
			[]common.Hash{
				common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000006d"),
				common.HexToHash("0x6c6f72656d20697073756d20646f6c6f722073697420616d657420636f6e7365"),
				common.HexToHash("0x6374657475722061646970697363696e6720656c697400000000000000000000"),
			},
			slices.Collect(hashes),
		)
	})

}

type mockStateAccess struct {
	storage map[common.Address]map[common.Hash]common.Hash
}

func newMockStateAccess() *mockStateAccess {
	return &mockStateAccess{
		storage: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func (m *mockStateAccess) GetState(addr common.Address, key common.Hash) common.Hash {
	if m.storage[addr] == nil {
		return common.Hash{}
	}
	return m.storage[addr][key]
}

func (m *mockStateAccess) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	if m.storage[addr] == nil {
		m.storage[addr] = make(map[common.Hash]common.Hash)
	}
	if value == (common.Hash{}) {
		delete(m.storage[addr], key)
		if len(m.storage[addr]) == 0 {
			delete(m.storage, addr)
		}
	} else {
		m.storage[addr][key] = value
	}
	return value
}

func (m *mockStateAccess) IsEmpty() bool {
	return len(m.storage) == 0
}

func TestGolemDBState(t *testing.T) {
	t.Run("small payload (≤31 bytes)", func(t *testing.T) {
		db := newMockStateAccess()
		key := common.HexToHash("0x1234")
		value := []byte("small payload")

		// Test Set
		stateblob.SetBlob(db, key, value)

		// Test Get
		retrieved := stateblob.GetBlob(db, key)
		require.Equal(t, value, retrieved)

		// Test Delete
		stateblob.DeleteBlob(db, key)
		retrieved = stateblob.GetBlob(db, key)
		require.Empty(t, retrieved)
		require.True(t, db.IsEmpty())
	})

	t.Run("large payload (>31 bytes)", func(t *testing.T) {
		db := newMockStateAccess()
		key := common.HexToHash("0x5678")
		value := []byte("this is a large payload that definitely exceeds thirty one bytes in length")

		// Test Set
		stateblob.SetBlob(db, key, value)

		// Test Get
		retrieved := stateblob.GetBlob(db, key)
		require.Equal(t, value, retrieved)

		// Test Delete
		stateblob.DeleteBlob(db, key)
		retrieved = stateblob.GetBlob(db, key)
		require.Empty(t, retrieved)
		require.True(t, db.IsEmpty())
	})

	t.Run("empty payload", func(t *testing.T) {
		db := newMockStateAccess()
		key := common.HexToHash("0x9abc")
		value := []byte{}

		// Test Set
		stateblob.SetBlob(db, key, value)

		// Test Get
		retrieved := stateblob.GetBlob(db, key)
		require.Equal(t, value, retrieved)

		// Test Delete
		stateblob.DeleteBlob(db, key)
		retrieved = stateblob.GetBlob(db, key)
		require.Empty(t, retrieved)
		require.True(t, db.IsEmpty())
	})

	t.Run("exactly 31 bytes", func(t *testing.T) {
		db := newMockStateAccess()
		key := common.HexToHash("0xdef0")
		value := []byte("this-is-exactly-31-bytes-long!!")
		require.Equal(t, 31, len(value))

		// Test Set
		stateblob.SetBlob(db, key, value)

		// Test Get
		retrieved := stateblob.GetBlob(db, key)
		require.Equal(t, value, retrieved)

		// Test Delete
		stateblob.DeleteBlob(db, key)
		retrieved = stateblob.GetBlob(db, key)
		require.Empty(t, retrieved)
		require.True(t, db.IsEmpty())
	})

	t.Run("exactly 32 bytes", func(t *testing.T) {
		db := newMockStateAccess()
		key := common.HexToHash("0xdef1")
		value := []byte("this-is-exactly-32-bytes-long!!!")
		require.Equal(t, 32, len(value))

		// Test Set
		stateblob.SetBlob(db, key, value)

		// Test Get
		retrieved := stateblob.GetBlob(db, key)
		require.Equal(t, value, retrieved)

		// Test Delete
		stateblob.DeleteBlob(db, key)
		retrieved = stateblob.GetBlob(db, key)
		require.Empty(t, retrieved)
		require.True(t, db.IsEmpty())
	})
}
