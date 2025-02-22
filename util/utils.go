package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type Hash [32]byte

// Hash is just [32]byte
var mainNetGenHash = Hash{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var testNet3GenHash = Hash{
	0x43, 0x49, 0x7f, 0xd7, 0xf8, 0x26, 0x95, 0x71,
	0x08, 0xf4, 0xa3, 0x0f, 0xd9, 0xce, 0xc3, 0xae,
	0xba, 0x79, 0x97, 0x20, 0x84, 0xe9, 0x0e, 0xad,
	0x01, 0xea, 0x33, 0x09, 0x00, 0x00, 0x00, 0x00,
}

var regTestGenHash = Hash{
	0x06, 0x22, 0x6e, 0x46, 0x11, 0x1a, 0x0b, 0x59,
	0xca, 0xaf, 0x12, 0x60, 0x43, 0xeb, 0x5b, 0xbf,
	0x28, 0xc3, 0x4f, 0x3a, 0x5e, 0x33, 0x2a, 0x1f,
	0xc7, 0xb2, 0xb7, 0x3c, 0xf1, 0x88, 0x91, 0x0f,
}

var sigNetGenHash = Hash{
	0xf6, 0x1e, 0xee, 0x3b, 0x63, 0xa3, 0x80, 0xa4,
	0x77, 0xa0, 0x63, 0xaf, 0x32, 0xb2, 0xbb, 0xc9,
	0x7c, 0x9f, 0xf9, 0xf0, 0x1f, 0x2c, 0x42, 0x25,
	0xe9, 0x73, 0x98, 0x81, 0x08, 0x00, 0x00, 0x00,
}

// For a given BitcoinNet, yields the genesis hash
// If the BitcoinNet is not supported, an error is
// returned.
func GenHashForNet(p chaincfg.Params) (*Hash, error) {

	switch p.Name {
	case "testnet3":
		return &testNet3GenHash, nil
	case "mainnet":
		return &mainNetGenHash, nil
	case "regtest":
		return &regTestGenHash, nil
	case "signet":
		return &sigNetGenHash, nil
	}
	return nil, fmt.Errorf("net not supported")
}

// HashFromString hashes the given string with sha256
func HashFromString(s string) Hash {
	return sha256.Sum256([]byte(s))
}

// turns an outpoint into a 36 byte... mixed endian thing.
// (the 32 bytes txid is "reversed" and the 4 byte index is in order (big)
func OutpointToBytes(op *wire.OutPoint) (b [36]byte) {
	copy(b[0:32], op.Hash[:])
	binary.BigEndian.PutUint32(b[32:36], op.Index)
	return
}

// blockToDelOPs gives all the UTXOs in a block that need proofs in order to be
// deleted.  All txinputs except for the coinbase input and utxos created
// within the same block (on the skiplist)
func BlockToDelOPs(
	blk *btcutil.Block) []wire.OutPoint {

	transactions := blk.Transactions()
	inCount, _, inskip, _ := DedupeBlock(blk)

	delOPs := make([]wire.OutPoint, 0, inCount-len(inskip))

	var blockInIdx uint32
	for txinblock, tx := range transactions {
		if txinblock == 0 {
			blockInIdx += uint32(len(tx.MsgTx().TxIn)) // coinbase can have many inputs
			continue
		}

		// loop through inputs
		for _, txin := range tx.MsgTx().TxIn {
			// check if on skiplist.  If so, don't make leaf
			if len(inskip) > 0 && inskip[0] == blockInIdx {
				// fmt.Printf("skip %s\n", txin.PreviousOutPoint.String())
				inskip = inskip[1:]
				blockInIdx++
				continue
			}

			delOPs = append(delOPs, txin.PreviousOutPoint)
			blockInIdx++
		}
	}
	return delOPs
}

// DedupeBlock takes a bitcoin block, and returns two int slices: the indexes of
// inputs, and idexes of outputs which can be removed.  These are indexes
// within the block as a whole, even the coinbase tx.
// So the coinbase tx in & output numbers affect the skip lists even though
// the coinbase ins/outs can never be deduped.  it's simpler that way.
func DedupeBlock(blk *btcutil.Block) (inCount, outCount int, inskip []uint32, outskip []uint32) {
	var i uint32
	// wire.Outpoints are comparable with == which is nice.
	inmap := make(map[wire.OutPoint]uint32)

	// go through txs then inputs building map
	for coinbase, tx := range blk.Transactions() {
		if coinbase == 0 { // coinbase tx can't be deduped
			i += uint32(len(tx.MsgTx().TxIn)) // coinbase can have many inputs
			continue
		}
		for _, in := range tx.MsgTx().TxIn {
			inmap[in.PreviousOutPoint] = i
			i++
		}
	}
	inCount = int(i)

	i = 0
	// start over, go through outputs finding skips
	for coinbase, tx := range blk.Transactions() {
		txOut := tx.MsgTx().TxOut
		if coinbase == 0 { // coinbase tx can't be deduped
			i += uint32(len(txOut)) // coinbase can have multiple outputs
			continue
		}

		for outidx, _ := range txOut {
			op := wire.OutPoint{Hash: *tx.Hash(), Index: uint32(outidx)}
			inpos, exists := inmap[op]
			if exists {
				inskip = append(inskip, inpos)
				outskip = append(outskip, i)
			}
			i++
		}
	}
	outCount = int(i)
	// sort inskip list, as it's built in order consumed not created
	sortUint32s(inskip)
	return
}

// it'd be cool if you just had .sort() methods on slices of builtin types...
func sortUint32s(s []uint32) {
	sort.Slice(s, func(a, b int) bool { return s[a] < s[b] })
}

// PrefixLen16 puts a 2 byte length prefix in front of a byte slice
func PrefixLen16(b []byte) []byte {
	l := uint16(len(b))
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, l)
	return append(buf.Bytes(), b...)
}

func PopPrefixLen16(b []byte) ([]byte, []byte, error) {
	if len(b) < 2 {
		return nil, nil, fmt.Errorf("PrefixedLen slice only %d long", len(b))
	}
	prefix, payload := b[:2], b[2:]
	var l uint16
	buf := bytes.NewBuffer(prefix)
	binary.Read(buf, binary.BigEndian, &l)
	if int(l) > len(payload) {
		return nil, nil, fmt.Errorf("Prefixed %d but payload %d left", l, len(payload))
	}
	return payload[:l], payload[l:], nil
}

// CheckMagicByte checks for the Bitcoin magic bytes.
// returns false if it didn't read the Bitcoin magic bytes.
// Checks only for testnet3 and mainnet
func CheckMagicByte(bytesgiven []byte) bool {
	if bytes.Compare(bytesgiven, []byte{0x0b, 0x11, 0x09, 0x07}) != 0 && //testnet
		bytes.Compare(bytesgiven, []byte{0xf9, 0xbe, 0xb4, 0xd9}) != 0 && // mainnet
		bytes.Compare(bytesgiven, []byte{0xfa, 0xbf, 0xb5, 0xda}) != 0 && // regtest
		bytes.Compare(bytesgiven, []byte{0x0a, 0x03, 0xcf, 0x40}) != 0 { // signet
		fmt.Printf("got non magic bytes %x, finishing\n", bytesgiven)
		return false
	}

	return true
}

// HasAccess reports whether we have access to the named file.
// Returns true if HasAccess, false if it doesn't.
// Does NOT tell us if the file exists or not.
// File might exist but may not be available to us
func HasAccess(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

//IsUnspendable determines whether a tx is spendable or not.
//returns true if spendable, false if unspendable.
func IsUnspendable(o *wire.TxOut) bool {
	switch {
	case len(o.PkScript) > 10000: //len 0 is OK, spendable
		return true
	case len(o.PkScript) > 0 && o.PkScript[0] == 0x6a: // OP_RETURN is 0x6a
		return true
	default:
		return false
	}
}
