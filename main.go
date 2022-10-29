package main

import (
	// "context"

	"context"
	"encoding/hex"
	"flag"
	"sync/atomic"
	"time"

	ipfssnapshot "github.com/Giulio2002/verkleipfs/ipfs-snapshot"
	"github.com/Giulio2002/verkleipfs/verkle"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	"github.com/ledgerwatch/log/v3"

	shell "github.com/ipfs/go-ipfs-api"
)

const workers = 3

func IncrementVerkleTree(verkleDb kv.RwDB, stateChaindata string) error {
	start := time.Now()

	stateDb, err := mdbx.Open(stateChaindata, log.Root(), true)
	if err != nil {
		log.Error("Error while opening database", "err", err.Error())
		return err
	}
	defer stateDb.Close()
	ctx := context.Background()
	vTx, err := verkleDb.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer vTx.Rollback()

	tx, err := stateDb.BeginRo(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ticker := time.NewTicker(4 * time.Second)
Loop:
	for {
		select {
		case <-ticker.C:

			from, err := stages.GetStageProgress(vTx, stages.VerkleTrie)
			if err != nil {
				return err
			}

			to, err := stages.GetStageProgress(tx, stages.Execution)
			if err != nil {
				return err
			}
			if from >= to {
				continue
			}

			rootHash, err := rawdb.ReadVerkleRoot(tx, from)
			if err != nil {
				return err
			}
			rootNode, err := rawdb.ReadVerkleNode(vTx, rootHash)
			if err != nil {
				return err
			}

			tree := verkle.NewTree(vTx, rootNode)
			if err := verkle.DoAccountAndCode(tx, tree, from, to); err != nil {
				return err
			}
			if err = verkle.DoStorage(tx, tree, from, to); err != nil {
				return err
			}
			var root common.Hash
			if root, err = tree.Commit(); err != nil {
				return err
			}
			if err := rawdb.WriteVerkleRoot(vTx, to, root); err != nil {
				return err
			}
			if err := stages.SaveStageProgress(vTx, stages.VerkleTrie, to); err != nil {
				return err
			}

			log.Info("Imported Verkle Segments", "root", root, "number", to, "elapesed", time.Since(start))
			start = time.Now()
			if err := vTx.Commit(); err != nil {
				return err
			}
			vTx, err = verkleDb.BeginRw(ctx)
			if err != nil {
				return err
			}

			tx, err = stateDb.BeginRo(ctx)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			break Loop
		}
	}

	return vTx.Commit()
}

func waitUntilValue(v *atomic.Value, c int) {
	for v.Load().(int) > c {
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	ctx := context.Background()
	start := time.Now()
	// Log handling
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(3), log.StderrHandler))
	// Flags settings
	daemonUrl := flag.String("daemon-url", "localhost:5001", "IPFS Daemon url")
	stateDb := flag.String("state-chaindata", "", "Main erigon chain data")
	verkleDb := flag.String("verkle-chaindata", "verle-db", "output verkle")
	tmpdir := flag.String("tmpdir", "/tmp/", "temorary files dir")
	flag.Parse()
	// Load Ipfs
	log.Info("Disclaimer: You need good internet connection + 250 GB of free space")
	sh := shell.NewShell(*daemonUrl)
	var counter atomic.Value
	counter.Store(0)
	snapshots := ipfssnapshot.MainnetVerkleIpfsSnapshot()
	// Download snapshots concurrently
	for i, snapshot := range snapshots {
		counter.Store(counter.Load().(int) + 1)
		go func(snapshot *ipfssnapshot.IPFSSnapshotSztd, i int) {
			log.Info("Started Processing snapshot", "id", i)
			if err := snapshot.LoadFile(sh, *tmpdir, uint64(i), ctx); err != nil {
				panic(err)
			}
			counter.Store(counter.Load().(int) - 1)
			log.Info("Finished Processing snapshot", "id", i)
		}(snapshot, i)
		if counter.Load().(int) > workers-1 {
			waitUntilValue(&counter, workers-1)
		}
	}
	// Wait for snapshots to end
	waitUntilValue(&counter, 0)
	// Create new database
	db, err := mdbx.Open(*verkleDb, log.Root(), false)
	if err != nil {
		panic(err)
	}

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		panic(err)
	}
	log.Info("Now starting dumping process to database")
	// Convert snapshot format to erigon format
	ticker := time.NewTicker(30 * time.Second)
	for _, snapshot := range snapshots {
		for k, v, err := snapshot.Next(); k != nil; k, v, err = snapshot.Next() {
			if err != nil {
				panic(err)
			}
			if len(v) == 0 {
				continue
			}
			if err := tx.Append(kv.VerkleTrie, k, v); err != nil {
				log.Error("Error", "err", err)
				return
			}
			select {
			case <-ticker.C:
				log.Info("Progress", "key", hex.EncodeToString(k))
			default:
			}
		}
		snapshot.Close()
	}

	// Write snapshot metadata
	if err := rawdb.WriteVerkleRoot(tx, ipfssnapshot.MainnetVerkleUpdateBlock,
		ipfssnapshot.MainnetVerkleUpdateRoot); err != nil {
		log.Error("Err", "err", err)
		return
	}

	if err := stages.SaveStageProgress(tx, stages.VerkleTrie, ipfssnapshot.MainnetVerkleUpdateBlock); err != nil {
		log.Error("Err", "err", err)
		return
	}
	tx.Commit()
	log.Info("Processed snapshot to block", "number", ipfssnapshot.MainnetVerkleUpdateBlock, "elapsed", time.Since(start))
	// Catch up with chain tip.
	if *stateDb != "" {
		if err := IncrementVerkleTree(db, *stateDb); err != nil {
			log.Error("Err", "err", err)
			return
		}
	} else {
		log.Warn("No erigon db specified, stopping...")
	}

}
