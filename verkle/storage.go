package verkle

import (
	"time"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common/changeset"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/turbo/trie/vtree"
	"github.com/ledgerwatch/log/v3"
)

func DoStorage(tx kv.Tx, tree *Tree, from, to uint64) error {
	cursor, err := tx.CursorDupSort(kv.StorageChangeSet)
	if err != nil {
		return err
	}
	defer cursor.Close()
	startKey := dbutils.EncodeBlockNumber(from)
	visited := NewVisitedMap()
	var key, value []byte
	logInterval := time.NewTicker(30 * time.Second)
	defer logInterval.Stop()
	for key, value, err = cursor.Seek(startKey); key != nil; key, value, err = cursor.Next() {
		if err != nil {
			return err
		}
		block, dbKey, _, err := changeset.DecodeStorage(key, value)
		if err != nil {
			return err
		}
		if block > to {
			break
		}
		if visited.Visited(string(dbKey)) {
			continue
		}
		visited.Visit(string(dbKey))
		storageValue, err := tx.GetOne(kv.PlainState, dbKey)
		if err != nil {
			return err
		}

		if block%7000 == 0 {
			_, err := tree.Commit()
			if err != nil {
				return err
			}
		}

		storageKey := new(uint256.Int).SetBytes(dbKey[28:])
		var storageValueFormatted []byte

		if len(storageValue) > 0 {
			storageValueFormatted = make([]byte, 32)
			int256ToVerkleFormat(new(uint256.Int).SetBytes(storageValue), storageValueFormatted)
			if err := tree.Insert(vtree.GetTreeKeyStorageSlot(dbKey[:20], storageKey), storageValueFormatted); err != nil {
				return err
			}
		} else {
			if err := tree.Insert(vtree.GetTreeKeyStorageSlot(dbKey[:20], storageKey), []byte{0}); err != nil {
				return err
			}
		}

		select {
		case <-logInterval.C:
			log.Info("Current Storage Progress", "num", block, "remaining", to-block)
		default:
		}
	}
	return nil
}
