package fixity

import (
	"io"

	"github.com/leeola/errors"
	"github.com/leeola/fixity/q"
)

// Fixity implements user focused writing and reading of data.
//
// This interface will be implemented for multiple stores, such as a local on
// disk store and a remote over network store.
type Fixity interface {
	// Blob returns a raw blob of the given hash.
	//
	// Mainly useful for inspecting the underlying data structure.
	//
	// TODO(leeola): change the name of this to something that does not conflict
	// with the Blob type. Since the Blob type is used to fetch the full rolled
	// contents of all the chunks, this method's name has an implication, which
	// is incorrect.
	Blob(hash string) (io.ReadCloser, error)

	// Blockchain allows one to manage and inspect the Fixity Blockchain.
	//
	// The blockchain is low level and should be used with care. See Blockchain
	// docstring for further details.
	Blockchain() Blockchain

	// Close shuts down any connections that may need to be closed.
	Close() error

	// Delete marks the given id's content to be garbage collected.
	//
	// Each Content, Blob and Chunk will be deleted if no other block in the
	// blockchain depends on it. Verifying this is done by the garbage
	// collector and is a slow process.
	Delete(id string) error

	// // ChecksumExists checks if the given checksum is in the Fixity store.
	// //
	// // This method exists to allow callers to avoid uploading data if it exists
	// // in the store.
	// //
	// // Note that this is the *checksum*, not the hash. Fixity hashes are hashes
	// // of the json data structure which contains the uploaded data. The checksum
	// // is the hash of *just* the uploaded data.
	// //
	// // The checksum is used so the caller does not have to split the data and
	// // construct chunks/blobs in the same fashion that Fixity did.
	// ChecksumExists(checksum string) (bool, error)

	// Read the latest Content with the given id.
	Read(id string) (Content, error)

	// Read the Content with the given hash.
	ReadHash(hash string) (Content, error)

	// Search for documents matching the given query.
	Search(*q.Query) ([]string, error)

	// Write the given reader to the fixity store and index fields.
	//
	// This is a shorthand for manually creating a WriteRequest.
	Write(id string, r io.Reader, f ...Field) (Content, error)

	// WriteRequest writes the given blob to Fixity with the associated settings.
	//
	// The returned Content will either be the newly written Content, or if the
	// write was ignored due to WriteRequest settings, Content will be whatever
	// Content matches the settings provided.
	//
	// For example, if IgnoreDuplicateBlob is provided, the caller must return
	// the Content that matches the Blob that would have been written.
	//
	// See WriteRequest docstring for documentation on expected behavior.
	WriteRequest(*WriteRequest) (Content, error)
}

// Content stores blob, index and history information for Fixity content.
type Content struct {
	// Id provides a user friendly way to reference a chain of Contents.
	//
	// History of Content is tracked through the PreviousContentHash chain,
	// however that does not provide a clear single identity for users.
	// The id field allows this, can be indexed and assocoated and is
	// easy to conceptualize.
	Id string `json:"id,omitempty"`

	// PreviousContentHash stores the previous Content for this Content.
	//
	// This allows a single entity, such as a file or a database "record"
	// to be mutated through time. To reference this history of contents,
	// the Id is used.
	PreviousContentHash string `json:"previousContentHash,omitempty"`

	// BlobHash is the hash of the  Blob containing this content's data.
	BlobHash string `json:"blobHash"`

	// IndexedFields contains the indexed metadata for this content.
	//
	// This allows the content to be searched for and can be used to
	// store basic metadata about the content.
	IndexedFields Fields `json:"indexedFields,omitempty"`

	// Hash is the hash of the Content itself, provided by Fixity.
	//
	// This value is not stored.
	//
	// TODO(leeola): enable json marshalling of this value, but ensure that
	// the writer zero values it before writing. This allows us to send the
	// Content over http.
	Hash string `json:"-"`

	// Index of this Content relative to Head Content for this Id.
	//
	// Eg, of an id with 5 Contents, the latest has an index of 1. The oldest,
	// and most outdated, has an index of 5. Calling Previous() on the latest
	// Content will result in a Content with an index of 2 and so on.
	//
	// Note that the Index starts at 1, not 0. This is because there are times
	// when an index cannot be known. A common example of this is when content
	// is loaded from a block and not in order from the Head Content. The zero
	// value conveys that the index of this content for it's Id is not known.
	//
	// This value is not stored.
	//
	// TODO(leeola): enable json marshalling of this value, but ensure that
	// the writer zero values it before writing. This allows us to send the
	// Content over http.
	Index int `json:"-"`

	// Store allows block method(s) to load previous content.
	//
	// This value is not stored.
	Store Store `json:"-"`
}

// Blob stores a series of ordered ChunkHashes
type Blob struct {
	// ChunkHashes contains a slice of chunk hashes for this blob.
	//
	// Depending on usage of NextBlobHash, this could be either all
	// chunk hashes or some chunk hashes.
	ChunkHashes []string `json:"chunkHashes"`

	// Size is the total bytes for the blob.
	Size int64 `json:"size,omitempty"`

	// Checksum of the blobs real bytes, ie not including the data structure.
	//
	// This is not a content address in Fixity! This serves to help verify
	// the written data and should be hex encoded for common CLI usage.
	//
	// The underlying hashing function is up to the store, but usually is the
	// same as what the store uses to hash the content addresses.
	Checksum string `json:"checksum,omitempty"`

	// ChunkSize is the average bytes each chunk is aimed to be.
	//
	// Chunks are separated by Cotent Defined Chunks (CDC) and this value
	// allows mutations of this blob to use the same ChunkSize with each
	// version. This ensures the chunks are chunk'd by the CDC algorithm
	// with the same spacing.
	//
	// Note that the algorithm is decided by the fixity.Store.
	AverageChunkSize uint64 `json:"averageChunkSize,omitempty"`

	// NextBlobHash is not currently supported / implemented anywhere, but
	// is required for very large storage. Eg, if there are so many chunks
	// for a given dataset that it cannot be stored in memory during writing
	// and reading, then we will need to split them up via NextBlobHash.
	//
	// // NextBlobHash stores another blob which is to be appended to this blob.
	// //
	// // This serves to allow very large blobs that cannot be loaded entirely
	// // into to memory to be split up into many parts.
	// NextBlobHash string `json:"nextBlobHash,omitempty"`

	// Hash is the hash of the Blob itself, provided by Fixity.
	//
	// This value is not stored.
	Hash string `json:"-"`

	// Store allows block method(s) to load previous content.
	//
	// This value is not stored.
	Store Store `json:"-"`
}

// Chunk represents a content defined chunk of data in fixity.
type Chunk struct {
	ChunkBytes []byte `json:"chunkBytes"`
	Size       int64  `json:"size"`

	// Start of this chunk within the bounds of the Blob.
	//
	// NOTE: This is not stored in the Fixity Store and is only a means to
	// allow the chunker to return additional data about the created chunk.
	// If this was stored in Fixity, each Chunk would have a different
	// Content Address, defeating the purpose of CDC & Content Addressed
	// storage.
	StartBoundry uint `json:"-"`

	// End of this chunk within the bounds of the Blob.
	//
	// NOTE: This is not stored in the Fixity Store and is only a means to
	// allow the chunker to return additional data about the created chunk.
	// If this was stored in Fixity, each Chunk would have a different
	// Content Address, defeating the purpose of CDC & Content Addressed
	// storage.
	EndBoundry uint `json:"-"`
}

func (b *Block) Previous() (Block, error) {
	if b.PreviousBlockHash == "" {
		return Block{}, ErrNoPrev
	}

	if b.Store == nil {
		return Block{}, errors.New("block: Store not set")
	}

	var previousBlock Block
	err := readAndUnmarshal(b.Store, b.PreviousBlockHash, &previousBlock)
	if err != nil {
		return Block{}, err
	}

	previousBlock.Hash = b.PreviousBlockHash
	previousBlock.Store = b.Store

	return previousBlock, nil
}

func (b *Block) Content() (Content, error) {
	if b.Store == nil {
		return Content{}, errors.New("block: Store not set")
	}

	if b.ContentBlock == nil {
		return Content{}, errors.New("block: not content block type")
	}

	var c Content
	err := readAndUnmarshal(b.Store, b.ContentBlock.Hash, &c)
	if err != nil {
		return Content{}, err
	}

	c.Hash = b.ContentBlock.Hash
	c.Store = b.Store

	return c, nil
}

func (c *Content) Blob() (Blob, error) {
	if c.Store == nil {
		return Blob{}, errors.New("content: Store not set")
	}

	if c.BlobHash == "" {
		return Blob{}, errors.New("content: blobHash is empty")
	}

	var b Blob
	err := readAndUnmarshal(c.Store, c.BlobHash, &b)
	if err != nil {
		return Blob{}, err
	}
	b.Hash = c.BlobHash
	b.Store = c.Store

	return b, nil
}

func (c *Content) Previous() (Content, error) {
	if c.PreviousContentHash == "" {
		return Content{}, ErrNoPrev
	}

	if c.Store == nil {
		return Content{}, errors.New("content: Store not set")
	}

	var pc Content
	err := readAndUnmarshal(c.Store, c.PreviousContentHash, &pc)
	if err != nil {
		return Content{}, err
	}
	pc.Hash = c.PreviousContentHash
	pc.Store = c.Store

	if c.Index != 0 {
		pc.Index = c.Index + 1
	}

	return pc, nil
}

func (c *Content) Read() (io.ReadCloser, error) {
	b, err := c.Blob()
	if err != nil {
		return nil, err
	}

	return b.Read()
}

func (b *Blob) Read() (io.ReadCloser, error) {
	if b.Store == nil {
		return nil, errors.New("read: Store not set")
	}

	return Reader(b.Store, b.Hash), nil
}
