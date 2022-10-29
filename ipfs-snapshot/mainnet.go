package ipfssnapshot

import "github.com/ledgerwatch/erigon/common"

const DefaultSize uint64 = 10523560000
const HalfSize uint64 = 7507572000

const MainnetVerkleUpdateBlock = uint64(15799899)

var MainnetVerkleUpdateRoot = common.BytesToHash(common.Hex2Bytes("02ad5cafd08690529b6c01575a94464bb45e87ffeb2bc2c29c267a1fc572aa63"))

func MainnetVerkleIpfsSnapshot() []*IPFSSnapshotSztd {
	return []*IPFSSnapshotSztd{
		NewIpfsSnapshotSztd("QmeMTPuSShowbxj1SHDBX87PnzB7atSBfsMD5rpqKRx7tR", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmXeyn3fNCuThBWht5FycvUTum46r53eNNFbqFvRAgttim", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmWwEq2TzyMWkRd3Ak75CuW4jBLAaP53fJ4kfbMpW4eA53", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmYuB3GTsH6h5UoSpaxyyt5MQF6Wby6gktL17CHFEMKyLc", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmcdpguV8PobUmTxbzUMZYXL2YLMD84zN1cNox8ESw93nV", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmeoujC3vC5cwgQJ49rJs1ZPDUTt4aGWuWCayo7EDZV1ek", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmNQmEPR9opvNhSMe8SRhiUhoeRsDaHPKNaXp6ddRoRVqK", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmRxLoArS5E77aTcq7LEHSBFHwa7foSLreLHvKYjsHEQyt", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmPwXcthLa2viz4jiQ21caZtNhLytjyAFM81eayUvEnKQM", DefaultSize, VerkleSnapshotIterator{}),
		NewIpfsSnapshotSztd("QmTzqnC2bDkJUCF2twUnG22K6SxrxQyUoqMXwJBS3Ho3md", HalfSize, VerkleSnapshotIterator{}),
	}
}
