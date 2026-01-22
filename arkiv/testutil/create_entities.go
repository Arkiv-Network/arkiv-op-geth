package testutil

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/arkiv/compression"
	"github.com/ethereum/go-ethereum/arkiv/storagetx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func (w *World) SubmitStorageTransaction(
	ctx context.Context,
	storageTx *storagetx.ArkivTransaction,
) error {

	client := w.GethInstance.ETHClient

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Get the current nonce for the sender address
	nonce, err := client.PendingNonceAt(ctx, w.FundedAccount.Address)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// RLP encode the storage transaction
	rlpData, err := rlp.EncodeToBytes(storageTx)
	if err != nil {
		return fmt.Errorf("failed to encode storage transaction: %w", err)
	}

	// Create UpdateStorageTx instance with the RLP encoded data
	txdata := &types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  big.NewInt(1e9), // 1 Gwei
		GasFeeCap:  big.NewInt(5e9), // 5 Gwei
		Gas:        12_800_000,
		To:         &address.ArkivProcessorAddress,
		Value:      big.NewInt(0), // No ETH transfer needed
		Data:       compression.MustBrotliCompress(rlpData),
		AccessList: types.AccessList{},
	}

	// Use the London signer since we're using a dynamic fee transaction
	signer := types.LatestSignerForChainID(chainID)

	// return fmt.Errorf("signer: %#v", signer)

	// Create and sign the transaction
	signedTx, err := types.SignNewTx(w.FundedAccount.PrivateKey, signer, txdata)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	return client.SendTransaction(ctx, signedTx)

}
