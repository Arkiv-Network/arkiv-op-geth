package dbevents

import (
	"fmt"

	"github.com/Arkiv-Network/arkiv-events/events"
	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/arkiv/logs"
	"github.com/ethereum/go-ethereum/arkiv/storagetx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func blockToEvents(rawBlock *types.Block, rawReceipts []*types.Receipt) (*events.Block, error) {

	bl := &events.Block{
		Number:     rawBlock.NumberU64(),
		Operations: []events.Operation{},
	}

	if len(rawReceipts) == 0 {
		return bl, nil
	}

	firstReceipt := rawReceipts[0]

	for opIndex, log := range firstReceipt.Logs {
		if len(log.Topics) == 0 {
			continue
		}
		if log.Topics[0] == logs.ArkivEntityExpired && len(log.Data) >= 32 {
			entityKey := common.BytesToHash(log.Data[:32])
			expire := events.OPExpire(entityKey.Bytes())
			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: 0,
				OpIndex: uint64(opIndex),
				Expire:  &expire,
			})
		}
	}

	for i, transaction := range rawBlock.Transactions() {
		transactionTo := transaction.To()
		if transactionTo == nil {
			continue
		}

		if *transactionTo != address.ArkivProcessorAddress {
			continue
		}

		receipt := rawReceipts[i]

		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}

		atx, err := storagetx.UnpackArkivTransaction(transaction.Data())
		if err != nil {
			return nil, fmt.Errorf("failed to unpack arkiv transaction: %w", err)
		}

		signer := types.LatestSignerForChainID(transaction.ChainId())
		from, err := signer.Sender(transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender from transaction: %w", err)
		}

		createdEntities := createdEntities(receipt)

		for opIndex, create := range atx.Create {
			createdEntityKey := createdEntities[0]
			createdEntities = createdEntities[1:]

			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: uint64(i),
				OpIndex: uint64(opIndex),
				Create: &events.OPCreate{
					Key:               createdEntityKey,
					ContentType:       create.ContentType,
					BTL:               create.BTL,
					Owner:             from,
					Content:           create.Payload,
					StringAttributes:  stringAnnotationsToMap(create.StringAnnotations),
					NumericAttributes: numericAnnotationsToMap(create.NumericAnnotations),
				},
			})
		}

		for opIndex, update := range atx.Update {

			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: uint64(i),
				OpIndex: uint64(opIndex),
				Update: &events.OPUpdate{
					Key:               update.EntityKey,
					ContentType:       update.ContentType,
					BTL:               update.BTL,
					Owner:             from,
					Content:           update.Payload,
					StringAttributes:  stringAnnotationsToMap(update.StringAnnotations),
					NumericAttributes: numericAnnotationsToMap(update.NumericAnnotations),
				},
			})
		}

		for opIndex, extendBTL := range atx.Extend {

			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: uint64(i),
				OpIndex: uint64(opIndex),
				ExtendBTL: &events.OPExtendBTL{
					Key: extendBTL.EntityKey,
					BTL: extendBTL.NumberOfBlocks,
				},
			})

		}
		for opIndex, changeOwner := range atx.ChangeOwner {

			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: uint64(i),
				OpIndex: uint64(opIndex),
				ChangeOwner: &events.OPChangeOwner{
					Key:   changeOwner.EntityKey,
					Owner: changeOwner.NewOwner,
				},
			})

		}
		for opIndex, delete := range atx.Delete {
			event := events.OPDelete(delete)

			bl.Operations = append(bl.Operations, events.Operation{
				TxIndex: uint64(i),
				OpIndex: uint64(opIndex),
				Delete:  &event,
			})
		}

	}

	return bl, nil
}

func createdEntities(r *types.Receipt) []common.Hash {
	entities := []common.Hash{}
	for _, log := range r.Logs {
		if log.Topics[0] == logs.ArkivEntityCreated {
			entityKey := log.Topics[1]
			entities = append(entities, entityKey)
		}
	}
	return entities
}

func stringAnnotationsToMap(annotations []storagetx.StringAnnotation) map[string]string {
	annotationsMap := make(map[string]string)
	for _, annotation := range annotations {
		annotationsMap[annotation.Key] = annotation.Value
	}
	return annotationsMap
}

func numericAnnotationsToMap(annotations []storagetx.NumericAnnotation) map[string]uint64 {
	annotationsMap := make(map[string]uint64)
	for _, annotation := range annotations {
		annotationsMap[annotation.Key] = annotation.Value
	}
	return annotationsMap
}
