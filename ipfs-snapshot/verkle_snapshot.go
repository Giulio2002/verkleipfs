package ipfssnapshot

import (
	"encoding/binary"
	"fmt"
	"io"
)

type VerkleSnapshotIterator struct{}

// Read Erigon/Geth Verkle snapshot format
func (VerkleSnapshotIterator) Next(r io.Reader) ([]byte, []byte, error) {
	key := make([]byte, 32) // Key is committment hash
	_, err := r.Read(key)
	if err == io.EOF {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	nodeLengthBytes := make([]byte, 8)
	if _, err := r.Read(nodeLengthBytes); err != nil {
		return nil, nil, err
	}
	nodeLength := binary.BigEndian.Uint64(nodeLengthBytes)
	// Read Verkle node
	value := make([]byte, nodeLength)
	var bytesRead int
	if bytesRead, err = r.Read(value); err != nil {
		return nil, nil, err
	}
	if bytesRead != int(nodeLength) {
		return nil, nil, fmt.Errorf("EOF Truncated")
	}
	return key, value, nil
}
