package mongoreplay

import (
	"fmt"
	"io"
	"time"
)

const RecordedOpMetadataSize = 8 + 1 + 12

// RecordedOp stores an op in addition to record/playback -related metadata
type RecordedOp struct {
	// Metadata
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
	//Read the length
	lenBuf := [4]byte{}
	_, err := io.ReadFull(reader, lenBuf[:])
	if err != nil {
		return err
	}

	recordLength := getInt32(lenBuf[:], 0)

	opBuffer := make([]byte, recordLength-4)

	_, err = io.ReadFull(reader, opBuffer)
	if err != nil {
		return err
	}

	r.MetadataFromSlice(opBuffer[:RecordedOpMetadataSize])

	r.BodyFromSlice(opBuffer[RecordedOpMetadataSize:])

	return nil
}

func (r *RecordedOp) MetadataFromSlice(s []byte) {
	offset := 0
	r.SeenConnectionNum = getInt64(s, offset)
	offset += 8

	r.EOF = (s[offset] == 1)
	offset += 1

	sec := getInt64(s, offset)
	offset += 8

	nsec := getInt32(s, offset)

	r.Seen = time.Unix(sec+internalToUnix, int64(nsec)).UTC()
}

func (r *RecordedOp) BodyFromSlice(s []byte) {
	if len(s) == 0 {
		r.RawOp = EmptyRawOp
		return
	}
	size := getInt32(s, 0)
	r.RawOp.Header.FromWire(s)
	r.RawOp.Body = s[:size]
}

func (r *RecordedOp) ToSlice(s []byte) {
	if cap(s) < RecordedOpMetadataSize+len(r.RawOp.Body)+4 {
		panic("slice to record to too small")
	}
	offset := 0
	setInt32(s, offset, int32(r.encodingSize()))
	offset += 4

	setInt64(s, offset, r.SeenConnectionNum)
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
	return len(r.RawOp.Body) + RecordedOpMetadataSize + 4
}
