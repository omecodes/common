package crypt

import (
	"bytes"
	"crypto/rand"
	"io"
)

type aesGCMEncryptReader struct {
	stream                io.Reader
	eof                   bool
	bufferedProcessed     *bytes.Buffer
	encryptionKey         []byte
	plainBlockSize        int32
	totalProcessedServed  int64
	plainRangeOffset      int64
	plainRangeLimit       int64
	plainDataStreamCursor int64
	reachedRangeLimit     bool

	options readOptions
}

func (reader *aesGCMEncryptReader) Read(b []byte) (int, error) {
	totalRead := 0
	l := len(b)

	for totalRead < l {
		availableData := reader.bufferedProcessed.Len()
		if availableData > 0 {
			n, _ := reader.bufferedProcessed.Read(b[totalRead:])
			totalRead += n
			reader.totalProcessedServed += int64(n)
			if totalRead == l {
				return totalRead, nil
			}

		} else if reader.eof {
			return totalRead, io.EOF
		}

		// prepare the next block
		buff := make([]byte, defaultBlockSize)
		count, err := readMax(reader.stream, buff)
		if err != nil {
			reader.eof = err == io.EOF
			if !reader.eof {
				return 0, err
			}
		}

		if count > 0 {
			nonce := make([]byte, AESGCMNonceSize)
			_, _ = rand.Read(nonce)

			data, err := AESGCMEncryptWithSalt(reader.encryptionKey, nonce, buff[:count])
			if err != nil {
				return 0, err
			}

			h := new(header)
			h.Nonce = nonce

			b := &encryptedBlock{}
			b.Header = h
			err = b.SetPayload(data)
			if err != nil {
				return 0, err
			}
			_, _ = b.Write(reader.bufferedProcessed)
		}
	}
	return totalRead, nil
}

func newAesGCMEncryptReader(key []byte, stream io.Reader, opts ...ReadOption) *aesGCMEncryptReader {
	reader := new(aesGCMEncryptReader)
	reader.stream = stream
	reader.encryptionKey = key
	reader.totalProcessedServed = 0
	reader.eof = false
	reader.bufferedProcessed = bytes.NewBuffer([]byte{})

	for _, o := range opts {
		o(&reader.options)
	}

	if reader.options.blockSize == 0 {
		reader.options.blockSize = defaultBlockSize
	}

	return reader
}
