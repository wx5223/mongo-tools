package mongoreplay

import (
	"fmt"
)

// RecordedOp stores an op in addition to record/playback -related metadata
type RecordedOp struct {
	SeenConnectionNum   int64
	PlayedConnectionNum int64
	Order               int64
	Generation          int
	EOF                 bool `bson:",omitempty"`

	Seen     *PreciseTime
	PlayAt   *PreciseTime `bson:",omitempty"`
	PlayedAt *PreciseTime `bson:",omitempty"`

	RawOp
}

type orderedOps []RecordedOp

func (o orderedOps) Len() int {
	return len(o)
}

func (o orderedOps) Less(i, j int) bool {
	return o[i].Seen.Before(o[j].Seen.Time)
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
