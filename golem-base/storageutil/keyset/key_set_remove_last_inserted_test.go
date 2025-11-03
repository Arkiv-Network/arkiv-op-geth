package keyset_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/allentities"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/keyset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestRemoveLastInserted(t *testing.T) {
	db := newMockStateAccess()

	k1 := newHash("0xa")
	k2 := newHash("0xb")
	k3 := newHash("0xc")
	k4 := newHash("0xd")

	keyset.AddValue(db, allentities.AllEntitiesKey, k1)
	keyset.AddValue(db, allentities.AllEntitiesKey, k2)
	keyset.AddValue(db, allentities.AllEntitiesKey, k3)
	keyset.AddValue(db, allentities.AllEntitiesKey, k4)

	require.Equal(t, uint256.NewInt(4), keyset.Size(db, allentities.AllEntitiesKey))

	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	err := keyset.RemoveValue(db, allentities.AllEntitiesKey, k4)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	require.Equal(t, uint256.NewInt(3), keyset.Size(db, allentities.AllEntitiesKey))

	keyset.AddValue(db, allentities.AllEntitiesKey, k4)
	require.Equal(t, uint256.NewInt(4), keyset.Size(db, allentities.AllEntitiesKey))

	err = keyset.RemoveValue(db, allentities.AllEntitiesKey, k4)
	require.NoError(t, err)

}

func TestRemoveSecondButLastInserted(t *testing.T) {
	db := newMockStateAccess()

	k1 := newHash("0xa")
	k2 := newHash("0xb")
	k3 := newHash("0xc")
	k4 := newHash("0xd")

	keyset.AddValue(db, allentities.AllEntitiesKey, k1)
	keyset.AddValue(db, allentities.AllEntitiesKey, k2)
	keyset.AddValue(db, allentities.AllEntitiesKey, k3)
	keyset.AddValue(db, allentities.AllEntitiesKey, k4)

	require.Equal(t, uint256.NewInt(4), keyset.Size(db, allentities.AllEntitiesKey))

	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))

	err := keyset.RemoveValue(db, allentities.AllEntitiesKey, k3)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))

	require.Equal(t, uint256.NewInt(3), keyset.Size(db, allentities.AllEntitiesKey))

	keyset.AddValue(db, allentities.AllEntitiesKey, k3)
	require.Equal(t, uint256.NewInt(4), keyset.Size(db, allentities.AllEntitiesKey))

	err = keyset.RemoveValue(db, allentities.AllEntitiesKey, k3)
	require.NoError(t, err)

}

func TestRemoveInOrder(t *testing.T) {
	db := newMockStateAccess()

	k1 := newHash("0xa")
	k2 := newHash("0xb")
	k3 := newHash("0xc")
	k4 := newHash("0xd")

	keyset.AddValue(db, allentities.AllEntitiesKey, k1)
	keyset.AddValue(db, allentities.AllEntitiesKey, k2)
	keyset.AddValue(db, allentities.AllEntitiesKey, k3)
	keyset.AddValue(db, allentities.AllEntitiesKey, k4)

	require.Equal(t, uint256.NewInt(4), keyset.Size(db, allentities.AllEntitiesKey))

	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k1))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k2))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	err := keyset.RemoveValue(db, allentities.AllEntitiesKey, k1)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k1))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k2))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	require.Equal(t, uint256.NewInt(3), keyset.Size(db, allentities.AllEntitiesKey))

	err = keyset.RemoveValue(db, allentities.AllEntitiesKey, k2)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k1))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k2))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	require.Equal(t, uint256.NewInt(2), keyset.Size(db, allentities.AllEntitiesKey))

	err = keyset.RemoveValue(db, allentities.AllEntitiesKey, k3)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k1))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k2))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))
	require.True(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	require.Equal(t, uint256.NewInt(1), keyset.Size(db, allentities.AllEntitiesKey))

	err = keyset.RemoveValue(db, allentities.AllEntitiesKey, k4)
	require.NoError(t, err)

	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k1))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k2))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k3))
	require.False(t, keyset.ContainsValue(db, allentities.AllEntitiesKey, k4))

	require.Equal(t, uint256.NewInt(0), keyset.Size(db, allentities.AllEntitiesKey))

}
