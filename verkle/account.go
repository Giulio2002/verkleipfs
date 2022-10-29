package verkle

import (
	"time"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/changeset"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/turbo/trie/vtree"
	"github.com/ledgerwatch/log/v3"
)

func DoAccountAndCode(tx kv.Tx, tree *Tree, from, to uint64) error {
	cursor, err := tx.CursorDupSort(kv.AccountChangeSet)
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
		block, dbKey, _, err := changeset.DecodeAccounts(key, value)
		if err != nil {
			return err
		}
		if block > to {
			break
		}
		address := common.BytesToAddress(dbKey)
		if visited.Visited(address) {
			continue
		}
		visited.Visit(address)

		var account accounts.Account
		decoded, err := rawdb.ReadAccount(tx, address, &account)
		if err != nil {
			return err
		}
		versionKey := vtree.GetTreeKeyVersion(dbKey)
		if !decoded {
			if err := tree.DeleteAccount(versionKey, true); err != nil {
				return err
			}
			continue
		}
		isContract := account.Incarnation > 0
		codeSize := uint64(0)
		if isContract {
			code, err := tx.GetOne(kv.Code, account.CodeHash[:])
			if err != nil {
				return err
			}
			codeSize = uint64(len(code))
			// Chunkify contract code and build keys for each chunks and insert them in the tree
			chunkedCode := vtree.ChunkifyCode(code)
			offset := byte(0)
			offsetOverflow := false
			currentKey := vtree.GetTreeKeyCodeChunk(address[:], uint256.NewInt(0))
			var chunks [][]byte
			var chunkKeys [][]byte
			// Write code chunks
			for i := 0; i < len(chunkedCode); i += 32 {
				chunks = append(chunks, common.CopyBytes(chunkedCode[i:i+32]))
				if currentKey[31]+offset < currentKey[31] || offsetOverflow {
					currentKey = vtree.GetTreeKeyCodeChunk(dbKey, uint256.NewInt(uint64(i)/32))
					chunkKeys = append(chunkKeys, common.CopyBytes(currentKey))
					offset = 1
					offsetOverflow = false
				} else {
					codeKey := common.CopyBytes(currentKey)
					codeKey[31] += offset
					chunkKeys = append(chunkKeys, common.CopyBytes(codeKey))
					offset += 1
					// If offset overflows, handle it.
					offsetOverflow = offset == 0
				}
			}
			if err := tree.WriteContractCodeChunks(chunkKeys, chunks); err != nil {
				return err
			}
		}

		if err := tree.UpdateAccount(versionKey, codeSize, isContract, account); err != nil {
			return err
		}
		select {
		case <-logInterval.C:
			log.Info("Current Account Progress", "num", block, "remaining", to-block)
		default:
		}
	}
	return nil
}
