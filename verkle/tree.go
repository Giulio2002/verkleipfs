package verkle

import (
	"encoding/binary"
	"time"

	"github.com/gballet/go-verkle"
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/turbo/trie/vtree"
	"github.com/ledgerwatch/log/v3"
)

const maxInsertions = 2_000_000

func int256ToVerkleFormat(x *uint256.Int, buffer []byte) {
	bbytes := x.ToBig().Bytes()
	if len(bbytes) > 0 {
		for i, b := range bbytes {
			buffer[len(bbytes)-i-1] = b
		}
	}
}

func flushVerkleNode(db kv.RwTx, node verkle.VerkleNode, logInterval *time.Ticker, key []byte) error {
	var err error
	totalInserted := 0
	node.(*verkle.InternalNode).Flush(func(node verkle.VerkleNode) {
		if err != nil {
			return
		}

		err = rawdb.WriteVerkleNode(db, node)
		if err != nil {
			return
		}
		totalInserted++
		select {
		case <-logInterval.C:
			log.Info("Flushing Verkle nodes", "inserted", totalInserted, "key", common.Bytes2Hex(key))
		default:
		}
	})
	return err
}

type Tree struct {
	db         kv.RwTx
	node       verkle.VerkleNode
	insertions uint64
}

func NewTree(db kv.RwTx, node verkle.VerkleNode) *Tree {
	return &Tree{
		db:   db,
		node: node,
	}
}

func (tree *Tree) UpdateAccount(versionKey []byte, codeSize uint64, isContract bool, acc accounts.Account) error {
	defer tree.commitIfPossible()

	resolverFunc := func(root []byte) ([]byte, error) {
		return tree.db.GetOne(kv.VerkleTrie, root)
	}

	var codeHashKey, nonceKey, balanceKey, codeSizeKey, nonce, balance, cs [32]byte
	copy(codeHashKey[:], versionKey[:31])
	copy(nonceKey[:], versionKey[:31])
	copy(balanceKey[:], versionKey[:31])
	copy(codeSizeKey[:], versionKey[:31])
	codeHashKey[31] = vtree.CodeKeccakLeafKey
	nonceKey[31] = vtree.NonceLeafKey
	balanceKey[31] = vtree.BalanceLeafKey
	codeSizeKey[31] = vtree.CodeSizeLeafKey
	// Process values
	int256ToVerkleFormat(&acc.Balance, balance[:])
	binary.LittleEndian.PutUint64(nonce[:], acc.Nonce)

	// Insert in the tree
	if err := tree.node.Insert(versionKey, []byte{0}, resolverFunc); err != nil {
		return err
	}

	if err := tree.node.Insert(nonceKey[:], nonce[:], resolverFunc); err != nil {
		return err
	}

	if err := tree.node.Insert(balanceKey[:], balance[:], resolverFunc); err != nil {
		return err
	}

	if isContract {
		binary.LittleEndian.PutUint64(cs[:], codeSize)
		if err := tree.node.Insert(codeHashKey[:], acc.CodeHash[:], resolverFunc); err != nil {
			return err
		}
		if err := tree.node.Insert(codeSizeKey[:], cs[:], resolverFunc); err != nil {
			return err
		}
	}
	tree.insertions += 6
	return nil
}

func (tree *Tree) DeleteAccount(versionKey []byte, isContract bool) error {
	defer tree.commitIfPossible()

	resolverFunc := func(root []byte) ([]byte, error) {
		return tree.db.GetOne(kv.VerkleTrie, root)
	}

	var codeHashKey, nonceKey, balanceKey, codeSizeKey [32]byte
	copy(codeHashKey[:], versionKey[:31])
	copy(nonceKey[:], versionKey[:31])
	copy(balanceKey[:], versionKey[:31])
	copy(codeSizeKey[:], versionKey[:31])
	codeHashKey[31] = vtree.CodeKeccakLeafKey
	nonceKey[31] = vtree.NonceLeafKey
	balanceKey[31] = vtree.BalanceLeafKey
	codeSizeKey[31] = vtree.CodeSizeLeafKey

	// Insert in the tree
	if err := tree.node.Insert(versionKey, []byte{0}, resolverFunc); err != nil {
		return err
	}

	if err := tree.node.Insert(nonceKey[:], []byte{0}, resolverFunc); err != nil {
		return err
	}

	if err := tree.node.Insert(balanceKey[:], []byte{0}, resolverFunc); err != nil {
		return err
	}

	if isContract {
		if err := tree.node.Insert(codeHashKey[:], []byte{0}, resolverFunc); err != nil {
			return err
		}
		if err := tree.node.Insert(codeSizeKey[:], []byte{0}, resolverFunc); err != nil {
			return err
		}
		tree.insertions += 2
	}
	tree.insertions += 5
	return nil
}

func (tree *Tree) Insert(key, value []byte) error {
	resolverFunc := func(root []byte) ([]byte, error) {
		return tree.db.GetOne(kv.VerkleTrie, root)
	}
	defer tree.commitIfPossible()
	tree.insertions++
	return tree.node.Insert(key, value, resolverFunc)
}

func (tree *Tree) WriteContractCodeChunks(codeKeys [][]byte, chunks [][]byte) error {
	resolverFunc := func(root []byte) ([]byte, error) {
		return tree.db.GetOne(kv.VerkleTrie, root)
	}
	defer tree.commitIfPossible()
	for i, codeKey := range codeKeys {
		if err := tree.node.Insert(codeKey, chunks[i], resolverFunc); err != nil {
			return err
		}
		tree.insertions++
	}
	return nil
}

func (tree *Tree) Commit() (common.Hash, error) {
	logInterval := time.NewTicker(30 * time.Second)
	defer logInterval.Stop()

	commitment := tree.node.ComputeCommitment().Bytes()
	return common.BytesToHash(commitment[:]), flushVerkleNode(tree.db, tree.node, logInterval, nil)
}

func (tree *Tree) commitIfPossible() {
	if tree.insertions >= maxInsertions {
		if _, err := tree.Commit(); err != nil {
			panic(err)
		}
		tree.insertions = 0
	}

}
