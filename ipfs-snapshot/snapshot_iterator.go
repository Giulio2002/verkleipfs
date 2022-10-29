package ipfssnapshot

import "io"

type SnapshotIterator interface {
	// Parse snapshot and return preimages one by one
	Next(r io.Reader) ([]byte, []byte, error) // Return next kv pair
}
