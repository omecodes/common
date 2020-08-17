package crypt

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	//"github.com/pydio/cells/common/proto/encryption"
)

const (
	aesGCMAuthTagSize = 16
	AESGCMNonceSize   = 12

	dataLengthMask uint32 = 0x7FFFFFFF

	// defaultBlockSize minimal size of block files are chunk into
	defaultBlockSize = int64(10485760) // 10Mb
)

type header struct {
	Nonce      []byte
	dataLength uint32
}

func (h *header) Write(writer io.Writer) (int, error) {
	totalWritten := 0

	n, err := writer.Write(h.Nonce)
	if err != nil {
		return 0, nil
	}
	totalWritten += n

	lenBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(lenBytes, h.dataLength)
	n, err = writer.Write(lenBytes)
	if err != nil {
		return 0, nil
	}
	totalWritten += n

	return totalWritten, nil
}

func (h *header) Read(reader io.Reader) (int, error) {
	totalRead := 0

	h.Nonce = make([]byte, 12)
	n, err := reader.Read(h.Nonce)
	if err != nil {
		return 0, err
	}
	totalRead += n

	buff := make([]byte, 4)
	n, err = reader.Read(buff)
	if err != nil {
		return 0, err
	}
	totalRead += n
	h.dataLength = binary.BigEndian.Uint32(buff)

	h.dataLength = h.dataLength & dataLengthMask
	return totalRead, nil
}

func (h *header) GetDataLength() uint32 {
	return h.dataLength & dataLengthMask
}

func (h *header) String() string {
	sb := strings.Builder{}
	sb.Write([]byte("\n[Header:\n"))
	sb.Write([]byte(fmt.Sprintf("\tNonce : %s\n", base64.StdEncoding.EncodeToString(h.Nonce))))
	sb.Write([]byte(fmt.Sprintf("\tLength: %d bytes\n", h.dataLength&dataLengthMask)))
	sb.Write([]byte("]"))
	return sb.String()
}

type encryptedBlock struct {
	Header     *header
	HeaderSize uint32
	Payload    []byte
}

func (b *encryptedBlock) SetPayload(payload []byte) error {
	l := len(payload)

	if uint32(l) > dataLengthMask {
		return errors.New("payload to big")
	}

	b.Payload = payload
	if b.Header == nil {
		b.Header = new(header)
	}

	b.Header.dataLength = (dataLengthMask & uint32(l)) | b.Header.dataLength
	return nil
}

func (b *encryptedBlock) GetPayloadLength() uint32 {
	return b.Header.GetDataLength()
}

func (b *encryptedBlock) Write(writer io.Writer) (int, error) {
	totalWritten := 0

	n, err := b.Header.Write(writer)
	if err != nil {
		return 0, err
	}

	totalWritten += n
	b.HeaderSize = uint32(totalWritten)

	n, err = writer.Write(b.Payload)
	if err != nil {
		return 0, err
	}
	totalWritten += n
	return totalWritten, err
}

func (b *encryptedBlock) Read(reader io.Reader) (int, error) {
	totalRead := 0

	b.Header = new(header)
	n, err := b.Header.Read(reader)
	if err != nil {
		return n, err
	}
	totalRead += n

	b.Payload = make([]byte, b.Header.GetDataLength())
	n, err = readMax(reader, b.Payload)
	if err != nil {
		return totalRead, err
	}

	totalRead += n
	return totalRead, nil
}

func readMax(reader io.Reader, buff []byte) (int, error) {
	totalRead := 0
	max := len(buff)
	for totalRead < max {
		n, err := reader.Read(buff[totalRead:])
		if err != nil {
			return totalRead + n, err
		}
		totalRead += n
	}
	return totalRead, nil
}
