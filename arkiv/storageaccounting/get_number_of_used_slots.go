package storageaccounting

import (
	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/arkiv/storageutil"
	"github.com/holiman/uint256"
)

func GetNumberOfUsedSlots(db storageutil.StateAccess) *uint256.Int {

	counter := uint256.NewInt(0)
	counter.SetBytes32(db.GetState(address.ArkivProcessorAddress, UsedSlotsKey).Bytes())

	return counter
}
