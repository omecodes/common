package crypt

import (
	"bytes"
	"io"
)

type aesGCMDecryptReader struct {
	stream                io.Reader
	eof                   bool
	bufferedProcessed     *bytes.Buffer
	encryptionKey         []byte
	totalProcessedServed  int64
	plainRangeOffset      int64
	plainRangeLimit       int64
	plainDataStreamCursor int64
	reachedRangeLimit     bool
}

func (reader *aesGCMDecryptReader) Read(b []byte) (int, error) {
	if reader.reachedRangeLimit {
		return 0, io.EOF
	}

	totalRead := 0
	l := len(b)

	leftToRead := int(reader.plainRangeLimit - reader.plainRangeOffset - reader.totalProcessedServed)
	if leftToRead <= 0 || leftToRead > l {
		leftToRead = l
	}

	for {
		expectedReadCount := leftToRead - totalRead

		n, _ := reader.bufferedProcessed.Read(b[totalRead:leftToRead])
		if n > 0 {
			totalRead += n
			reader.totalProcessedServed += int64(n)
			reader.reachedRangeLimit = reader.plainRangeLimit == reader.plainRangeOffset+int64(reader.totalProcessedServed)
		}

		if n == expectedReadCount || reader.reachedRangeLimit {
			return totalRead, nil
		}

		if reader.eof {
			reader.reachedRangeLimit = true
			return totalRead, nil
		}

		b := &encryptedBlock{}
		count, err := b.Read(reader.stream)
		if err != nil {
			reader.eof = err == io.EOF
			if !reader.eof {
				return 0, err
			}
		}

		if count > 0 {
			data, err := AESGCMDecryptWithNonce(reader.encryptionKey, b.Header.Nonce, b.Payload)
			if err != nil {
				return 0, err
			}

			// We skip out of range data
			if reader.plainDataStreamCursor < reader.plainRangeOffset {
				bytesToConsumeSize := int(reader.plainRangeOffset - reader.plainDataStreamCursor)
				if bytesToConsumeSize > len(data) {
					reader.plainDataStreamCursor = reader.plainDataStreamCursor + int64(len(data))
					data = data[0:0]
				} else {
					reader.plainDataStreamCursor = reader.plainDataStreamCursor + int64(bytesToConsumeSize)
					data = data[bytesToConsumeSize:]
				}
			}
			reader.bufferedProcessed.Write(data)
		}
	}
}

func newAesGCMDecryptReader(key []byte, stream io.Reader, opts ...ReadOption) *aesGCMDecryptReader {
	reader := new(aesGCMDecryptReader)
	reader.stream = stream
	reader.encryptionKey = key
	reader.totalProcessedServed = 0
	reader.eof = false
	reader.bufferedProcessed = bytes.NewBuffer([]byte{})

	ropts := new(readOptions)
	for _, o := range opts {
		o(ropts)
	}

	if ropts.withRange {
		reader.plainRangeOffset = ropts.offset
		reader.plainRangeLimit = ropts.limit
	}
	return reader
}
