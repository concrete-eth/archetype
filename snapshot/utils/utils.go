package utils

import (
	"bytes"
	"compress/gzip"

	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)
	gz, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if _, err := out.ReadFrom(gz); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func EncodeSnapshotSlot(value common.Hash) ([]byte, error) {
	// Encoding []byte cannot fail, ok to ignore the error
	return rlp.EncodeToBytes(common.TrimLeftZeroes(value.Bytes()))
}

func DecodeSnapshotSlot(enc []byte) (common.Hash, error) {
	var value common.Hash
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			return common.Hash{}, err
		}
		value.SetBytes(content)
	}
	return value, nil
}

func DecodeSnapshotStorageRoot(encRoot []byte) common.Hash {
	root := common.BytesToHash(encRoot)
	if root == (common.Hash{}) {
		return types.EmptyRootHash
	}
	return root
}

func StorageItToBlob(it snapshot_types.StorageIterator) ([]byte, error) {
	blob := make([]byte, 0)
	for it.Next() {
		// Get storage key and value
		slot := it.Hash()
		enc := it.Slot()
		value, err := DecodeSnapshotSlot(enc)
		if err != nil {
			return nil, err
		}
		// Append to blob
		blob = append(blob, slot.Bytes()...)
		blob = append(blob, value.Bytes()...)
	}
	return blob, nil
}

func BlobToStorageIt(blob []byte) snapshot_types.StorageIterator {
	return NewBlobIterator(blob)
}

func MappingToBlob(mapping map[common.Hash][]byte) ([]byte, error) {
	blob := make([]byte, 0, len(mapping)*64)
	for slot, enc := range mapping {
		// Get value
		value, err := DecodeSnapshotSlot(enc)
		if err != nil {
			return nil, err
		}
		// Append to blob
		blob = append(blob, slot.Bytes()...)
		blob = append(blob, value.Bytes()...)
	}
	return blob, nil
}

type blobIterator struct {
	blob  []byte
	index int
}

var _ snapshot_types.StorageIterator = (*blobIterator)(nil)

func NewBlobIterator(blob []byte) *blobIterator {
	return &blobIterator{
		blob:  blob,
		index: 0,
	}
}

func (it *blobIterator) Next() bool {
	if it.index >= len(it.blob) {
		return false
	}
	it.index += 64
	return true
}

func (it *blobIterator) Error() error {
	return nil
}

func (it *blobIterator) Release() {
}

func (it *blobIterator) Hash() common.Hash {
	return common.BytesToHash(it.blob[it.index-64 : it.index-32])
}

func (it *blobIterator) Slot() []byte {
	enc, _ := EncodeSnapshotSlot(common.BytesToHash(it.blob[it.index-32 : it.index]))
	return enc
}
