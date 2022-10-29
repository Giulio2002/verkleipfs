# VerkleIPFS

VerkleIPFS is a proof-of-concept tool that allows to generate and retrieve Mainnet Verkle trees in significantly reduced time by using snapshot compressed and stored on the IPFS network.

## Rationale

Computation Times of Verkle Trees is one of the main bottlenecks behind why it is unfeasible at current dates to switch the Ethereum Protocol from Merkle Trees to Verkle Trees, simply the chain would have to be stopped for 4-9 days so that all nodes can convert their data over to the other format. Verkle Trees are larger and takes more time to compute by nature than Merkle Patricia Trees, as a matter of fact, on Erigon Merkle Computation takes 40 Minutes and take 80 GB of disk, while Verkle Trees takes 18 Hours and 200 GB of disk, still not optimal. so the solution is to have a powerful machine generate the Verkle Tree, upload it as snapshot on IPFS, and then have common nodes operator download those snapshot for transiton the Verkle Tree from IPFS/Bittorent/Arweave/Etc...

## Technical Implementation Information 

So Geth and Erigon for now shares the same Database format and Erigon, for testing purposes so these snapshots are theorically compatible for both. the snapshots are ZSTD compressed by using the `zstd` cli of the dump format of verkle trees. The dump format just consist of a series of bytes that denotes Verkle trees as key/value pair: `Verkle Root => Rlp Node` and are a concatenation of the following: Verkle Root(32 bytes) + 8 bytes that repressent Big endian encoded node length + node RLP. Quite simple.

## Buidling and Runling

```
go build
./verkleipfs --state-chaindata=<Erigon database location>
```

if you cannot specify `--state-chaindata` it will not catch up to chain tip and stop whenever snapshots ends.

## Disclaimer

This is just going to work as long as I keep my daemon running, when I stop it the implementation will stop working, However I yoloed the snapshots on Pinata so worst thing that happens it will be slow to retrieve them but still will not take 18 hours :). Also Disk Requirement are just 250 GB of disk Avaiable otherwise you wont have enough storage to store this marvelous tree.

## Rough Benchmarks (probably not realistic but almost there)

| Geth Default                  	| 4-7 Days                                                   	|
|-------------------------------	|------------------------------------------------------------	|
| Erigon Default (Regeneration) 	| 17:52 Hours                                                	|
| Geth (Preimages)              	| 17:52 Hours (untested but likely to be on par with Erigon) 	|
| Erigon Default (Incremental)  	| 27:32 Hours                                                	|
| Erigon (Snapshots)            	| 4:42 Hours                                                 	|

This a POC developed at an Hackathon so dont expect too much if it does not work for you or whatever, please dont be mad at me :(.