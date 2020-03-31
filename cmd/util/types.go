package util

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type Hash [32]byte

//simutil.Hash is just [32]byte
var MainnetGenHash = Hash{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var TestNet3GenHash = Hash{
	0x43, 0x49, 0x7f, 0xd7, 0xf8, 0x26, 0x95, 0x71,
	0x08, 0xf4, 0xa3, 0x0f, 0xd9, 0xce, 0xc3, 0xae,
	0xba, 0x79, 0x97, 0x20, 0x84, 0xe9, 0x0e, 0xad,
	0x01, 0xea, 0x33, 0x09, 0x00, 0x00, 0x00, 0x00,
}

//HashFromString hahes the given string with sha256
func HashFromString(s string) Hash {
	return sha256.Sum256([]byte(s))
}

//Struct for a tx to be converted to LeafTXOs
type Txotx struct {
	//txid of the tx
	Outputtxid string

	//Whether the output is an OP_RETURN or not
	Unspendable []bool

	//When the output is spent
	DeathHeights []uint32
}

// Tx defines a bitcoin transaction that provides easier and more efficient
// manipulation of raw transactions.  It also memoizes the hash for the
// transaction on its first access so subsequent accesses don't have to repeat
// the relatively expensive hashing operations.
type ProofTx struct {
	msgTx         *wire.MsgTx // Underlying MsgTx
	txHash        *Hash       // Cached transaction hash
	txHashWitness *Hash       // Cached transaction witness hash
	txHasWitness  *bool       // If the transaction has witness data
	txIndex       int         // Position within a block or TxIndexUnknown
}

// RawHeaderData is used for blk*.dat offsetfile building
// Used for ordering blocks as they aren't stored in order in the blk files.
// Includes 32 bytes of sha256 hash along with other variables
// needed for offsetfile building.
type RawHeaderData struct {
	// CurrentHeaderHash is the double hashed 32 byte header
	CurrentHeaderHash [32]byte
	// Prevhash is the 32 byte previous header included in the 80byte header.
	// Needed for ordering
	Prevhash [32]byte
	// FileNum is the blk*.dat file number
	FileNum [4]byte
	// Offset is where it is in the .dat file.
	Offset [4]byte
}

type BlockToWrite struct {
	Txs       []*btcutil.Tx
	Height    int32
	Blockhash [32]byte
}

// LeafData is all the data that goes into a leaf in the utreexo accumulator
type LeafData struct {
	BlockHash    [32]byte
	Outpoint     wire.OutPoint
	CbHeight     int32
	Amt          int64
	ScriptPubkey []byte
}

func LeafDataFromBytes(b []byte) (LeafData, error) {
	var l LeafData
	if len(b) < 80 {
		return l, fmt.Errorf("Not long enough for leafdata, need 80 bytes")
	}
	copy(l.BlockHash[:], b[0:32])
	copy(l.Outpoint.Hash[:], b[32:64])
	l.Outpoint.Index = BtU32(b[64:68])
	l.CbHeight = BtI32(b[68:72])
	l.Amt = BtI64(b[72:80])
	l.ScriptPubkey = b[80:]

	return l, nil
}

// turn a LeafData into bytes
func (l *LeafData) ToBytes() (b []byte) {
	b = append(l.BlockHash[:], l.Outpoint.Hash[:]...)
	b = append(b, U32tB(l.Outpoint.Index)...)
	b = append(b, I32tB(l.CbHeight)...)
	b = append(b, I64tB(l.Amt)...)
	b = append(b, l.ScriptPubkey...)
	return
}

// turn a LeafData into a LeafHash
func (l *LeafData) LeafHash() Hash {
	return sha256.Sum256(l.ToBytes())
}
