package mongoreplay

import (
	"fmt"
	"io"
	"time"
)

const RecordedOpMetadataSize = 8 + 1 + 12

// RecordedOp stores an op in addition to record/playback -related metadata
type RecordedOp struct {
	SeenConnectionNum int64
	EOF               bool
	Seen              time.Time

	RawOp

	// THESE FIELDS DON'T NEED TO BE SERIALIZED AT ALL
	PlayAt              time.Time
	PlayedAt            time.Time
	Generation          int
	PlayedConnectionNum int64
	Order               int64
}

type orderedOps []RecordedOp

func (o orderedOps) Len() int {
	return len(o)
}

func (o orderedOps) Less(i, j int) bool {
	return o[i].Seen.Before(o[j].Seen)
}

func (o orderedOps) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o *orderedOps) Pop() interface{} {
	i := len(*o) - 1
	op := (*o)[i]
	*o = (*o)[:i]
	return op
}

func (o *orderedOps) Push(op interface{}) {
	*o = append(*o, op.(RecordedOp))
}

type opKey struct {
	connectionNum int64
	opID          int32
}

func (k *opKey) String() string {
	return fmt.Sprintf("opID:%d, connectionNum:%d", k.connectionNum, k.opID)
}

func (r *RecordedOp) FromReader(reader io.Reader) error {
	b := [RecordedOpMetadataSize + 4]byte{}
	metadataBuffer := b[:]
	_, err := io.ReadFull(reader, metadataBuffer)
	if err != nil {
		return err
	}

	r.MetadataFromSlice(metadataBuffer)
	fmt.Println(r.Seen)
	fmt.Println(metadataBuffer)

	size := getInt32(metadataBuffer, RecordedOpMetadataSize)
	fmt.Println(size)

	bodyBuffer := make([]byte, size)
	copy(bodyBuffer[:4], metadataBuffer[RecordedOpMetadataSize:])
	_, err = io.ReadFull(reader, bodyBuffer[4:])
	if err != nil {
		return err
	}
	r.BodyFromSlice(bodyBuffer)

	return nil
}

func (r *RecordedOp) MetadataFromSlice(s []byte) {
	offset := 0

	r.Order = getInt64(s, offset)
	offset += 8

	r.EOF = (s[offset] == 1)
	offset += 1

	sec := getInt64(s, offset)
	offset += 4

	nsec := getInt32(s, offset)

	r.Seen = time.Unix(sec+internalToUnix, int64(nsec))
}

func (r *RecordedOp) BodyFromSlice(s []byte) {
	size := getInt32(s, 0)
	r.RawOp.Header.FromWire(s)
	r.RawOp.Body = s[:size]
}

func (r *RecordedOp) ToSlice(s []byte) {
	if cap(s) < RecordedOpMetadataSize+len(r.RawOp.Body) {
		panic("slice to record to too small")
	}
	offset := 0

	setInt64(s, offset, r.Order)
	offset += 8

	if r.EOF {
		s[offset] = 1
	} else {
		s[offset] = 0
	}
	offset += 1

	setInt64(s, offset, r.Seen.Unix()+unixToInternal)
	offset += 8
	setInt32(s, offset, int32(r.Seen.Nanosecond()))
	offset += 4
	copy(s[offset:], r.Body)
}

func (r *RecordedOp) ToWriter(w io.Writer) error {
	buf := make([]byte, r.encodingSize())
	r.ToSlice(buf)
	_, err := w.Write(buf)
	return err
}

func (r *RecordedOp) encodingSize() int {
	return len(r.RawOp.Body) + RecordedOpMetadataSize
}

func readRecordedOpBody(r io.Reader, s []byte, size int32) error {
	if cap(s) < int(size) {
		panic("recorded op body slice too small")
	}
	_, err := io.ReadFull(r, s[:size])
	if err != nil {
		return err
	}
	return nil
}

func readRecordedOpMetadata(r io.Reader, s []byte) error {
	if cap(s) < RecordedOpMetadataSize+4 {
		panic("recorded op header slice too small")
	}
	_, err := io.ReadFull(r, s[:RecordedOpMetadataSize+4])
	if err != nil {
		return err
	}
	return nil
}
