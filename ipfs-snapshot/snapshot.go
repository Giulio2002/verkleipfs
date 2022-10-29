package ipfssnapshot

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/c2h5oh/datasize"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ledgerwatch/log/v3"
)

type IPFSSnapshotSztd struct {
	cid   string
	size  uint64
	r     io.Reader
	fname string
	it    SnapshotIterator
}

func NewIpfsSnapshotSztd(cid string, size uint64, it SnapshotIterator) *IPFSSnapshotSztd {
	return &IPFSSnapshotSztd{
		cid:  cid,
		size: size,
		it:   it,
	}
}

func (i *IPFSSnapshotSztd) LoadFile(sh *shell.Shell, path string, id uint64, ctx context.Context) error {
	out := fmt.Sprintf("%s%d.txt", path, id)
	doneCh := make(chan struct{})

	go func() {
		logInterval := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-doneCh:
				return
			case <-ctx.Done():
				return
			case <-logInterval.C:
				file, err := os.Open(out)
				if err != nil {
					panic(err)
				}
				fi, err := file.Stat()
				if err != nil {
					panic(err)
				}
				log.Info("Downloading Snapshot Progress", "id", id, "cid", i.cid, "progress",
					datasize.ByteSize(fi.Size()).HumanReadable()+"/"+datasize.ByteSize(i.size).HumanReadable())
			}
		}
	}()
	err := sh.Get(i.cid, out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	doneCh <- struct{}{}
	log.Info("Now start decompression of snapshot", "id", id)
	i.fname = out + ".dec"
	// Generation of decompressed object (I am sorry, this is a sin)
	if err := exec.Command("unzstd", out, "-o", i.fname).Run(); err != nil {
		return err
	}

	i.r, err = os.Open(i.fname)
	if err != nil {
		return err
	}
	// Get rid of compressed snapshot
	return os.Remove(out)
}

func (i *IPFSSnapshotSztd) Next() ([]byte, []byte, error) {
	return i.it.Next(i.r)
}

func (i *IPFSSnapshotSztd) Close() {
	os.Remove(i.fname)
}
