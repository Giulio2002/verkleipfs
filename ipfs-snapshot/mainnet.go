package ipfssnapshot

import "github.com/ledgerwatch/erigon/common"

const DefaultSize uint64 = 10523560000
const HalfSize uint64 = 7507572000

const MainnetVerkleUpdateBlock = uint64(16061713)

var MainnetVerkleUpdateRoot = common.BytesToHash(common.Hex2Bytes("048ac0e036f02f4b4572502dbd506d9673bff18462ba6c587c8e9e9b9aa3304c"))

func MainnetVerkleIpfsSnapshot() []*IPFSSnapshotSztd {
	return []*IPFSSnapshotSztd{
		NewIpfsSnapshotSztd("Qmb15kNqrb9nk5BMbZDUXHzDbUde13HcsmFJeMcLFcU1aw", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmcVKx3HZpMunFSyogy4vbH1jZ91xXZfBJyKUJ9sYbMAw6", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmZuqJpsu71ZqW5VVMxBJwGZPzqqV8AcNg9JW1pa3Y1wPg", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmRvpJWC4XjP7XiLYpkk1CepCVvCgis4DQmSo43Tc3oSPi", DefaultSize, VerkleSnapshotIterator{}),

		NewIpfsSnapshotSztd("QmRvpJWC4XjP7XiLYpkk1CepCVvCgis4DQmSo43Tc3oSPi", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("Qmea37QzxcwxnM1sLkJnF5SUCC5d9njdRFmSwG8QTj7bgt", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmUdHXAkuvuLfMiTor4Ea2zNgYVxyY6FE19MAbmveHVgu4", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmXWkwNq8q649yxxinr3fk6yTYZtyf9T8GsbDk3ypjSozo", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmNUTXa4zSRoreqz1ptsV2ivFnH5NNA3fLAFT8zWXLNLt5", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmRrEY4yP3cxGdJ5aSNxNaQVyRfnqnPayxfNhbueywfeGw", HalfSize, VerkleSnapshotIterator{}),
	}
}
