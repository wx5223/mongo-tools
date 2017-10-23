package mongoreplay

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	mgo "github.com/10gen/llmgo"
	"github.com/10gen/llmgo/bson"
)

type testDoc struct {
	Name           string `bson:"name"`
	DocumentNumber int    `bson:"docNum"`
	Success        bool   `bson:"success"`
}

func TestOpGetMore(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := GetMoreOp{}
	op.Collection = "mongoreplay_test.test"
	op.CursorId = 12345
	op.Limit = -1

	t.Logf("Generated Getmore: %#v\n", op.GetMoreOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.GetMoreOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}
	getMoreOp := receivedOp.(*GetMoreOp)

	t.Log("Comparing parsed Getmore to original Getmore")
	switch {
	case getMoreOp.Collection != "mongoreplay_test.test":
		t.Errorf("Collection not matched. Saw %v -- Expected %v\n", getMoreOp.Collection, "mongoreplay_test.test")
	case getMoreOp.CursorId != 12345:
		t.Errorf("CursorId not matched. Saw %v -- Expected %v\n", getMoreOp.CursorId, 12345)
	case getMoreOp.Limit != -1:
		t.Errorf("Limit not matched. Saw %v -- Expected %v\n", getMoreOp.Limit, -1)
	}
}

func TestOpDelete(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := DeleteOp{}
	op.Collection = "mongoreplay_test.test"
	op.Flags = 7
	selector := bson.D{{"test", 1}}
	op.Selector = selector

	t.Logf("Generated Delete: %#v\n", op.DeleteOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.DeleteOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}
	deleteOp := receivedOp.(*DeleteOp)

	t.Log("Comparing parsed Delete to original Delete")
	switch {
	case deleteOp.Collection != "mongoreplay_test.test":
		t.Errorf("Collection not matched. Saw %v -- Expected %v\n", deleteOp.Collection, "mongoreplay_test.test")
	case deleteOp.Flags != 7:
		t.Errorf("Flags not matched. Saw %v -- Expected %v\n", deleteOp.Flags, 7)
	case !reflect.DeepEqual(deleteOp.Selector, &selector):
		t.Errorf("Selector not matched. Saw %v -- Expected %v\n", deleteOp.Selector, &selector)
	}
}

func TestInsertOp(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := InsertOp{}
	op.Collection = "mongoreplay_test.test"
	op.Flags = 7

	documents := []interface{}(nil)
	for i := 0; i < 10; i++ {
		insertDoc := &testDoc{
			DocumentNumber: i,
			Success:        true,
		}
		documents = append(documents, insertDoc)
	}
	op.Documents = documents
	t.Logf("Generated Insert: %#v\n", op.InsertOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.InsertOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}
	insertOp := receivedOp.(*InsertOp)

	t.Log("Comparing parsed Insert to original Insert")
	switch {
	case insertOp.Collection != "mongoreplay_test.test":
		t.Errorf("Collection not matched. Saw %v -- Expected %v\n", insertOp.Collection, "mongoreplay_test.test")
	case insertOp.Flags != 7:
		t.Errorf("Flags not matched. Saw %v -- Expected %v\n", insertOp.Flags, 7)
	}

	for i, doc := range insertOp.Documents {
		marshaled, _ := bson.Marshal(documents[i])
		unmarshaled := &bson.Raw{}
		bson.Unmarshal(marshaled, unmarshaled)
		if !reflect.DeepEqual(*unmarshaled, doc) {
			t.Errorf("Document not matched. Saw %v -- Expected %v\n", *unmarshaled, doc)
		}
	}
}

func TestKillCursorsOp(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := KillCursorsOp{}
	op.CursorIds = []int64{123, 456, 789, 55}

	t.Logf("Generated KillCursors: %#v\n", op.KillCursorsOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.KillCursorsOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}
	killCursorsOp := receivedOp.(*KillCursorsOp)

	t.Log("Comparing parsed KillCursors to original KillCursors")
	if !reflect.DeepEqual(killCursorsOp.CursorIds, op.CursorIds) {
		t.Errorf("CursoriD Arrays not matched. Saw %v -- Expected %v\n", killCursorsOp.CursorIds, op.CursorIds)
	}
}

func TestQueryOp(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := QueryOp{}
	op.Collection = "mongoreplay_test.test"
	op.Flags = 0
	op.HasOptions = true
	op.Limit = -1
	op.Skip = 0

	selector := map[string]int{
		"test": 1,
	}
	doc, err := bson.Marshal(selector)
	if err != nil {
		t.Error(err)
	}
	op.Selector = &bson.Raw{}
	bson.Unmarshal(doc, op.Selector)
	options := mgo.QueryWrapper{}
	options.Explain = false
	options.OrderBy = &bson.D{{"_id", 1}}
	op.Options = options

	t.Logf("Generated Query: %#v\n", op.QueryOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.QueryOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}
	queryOp := receivedOp.(*QueryOp)

	t.Log("Comparing parsed Query to original Query")
	switch {
	case queryOp.Collection != op.Collection:
		t.Errorf("Collections not equal. Saw %v -- Expected %v\n", queryOp.Collection, op.Collection)
	case !reflect.DeepEqual(queryOp.Selector, op.Selector):
		t.Errorf("Selectors not equal. Saw %v -- Expected %v\n", queryOp.Selector, op.Selector)
	case queryOp.Flags != op.Flags:
		t.Errorf("Flags not equal. Saw %d -- Expected %d\n", queryOp.Flags, op.Flags)
	case queryOp.Skip != op.Skip:
		t.Errorf("Skips not equal. Saw %d -- Expected %d\n", queryOp.Skip, op.Skip)
	case queryOp.Limit != op.Limit:
		t.Errorf("Limits not equal. Saw %d -- Expected %d\n", queryOp.Limit, op.Limit)
	}
	//currently we do not test the Options functionality of mgo
}

func TestOpUpdate(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := UpdateOp{}
	// Selector
	selector := map[string]int{
		"test": 1,
	}
	doc, err := bson.Marshal(selector)
	if err != nil {
		t.Error(err)
	}

	op.Selector = &bson.Raw{}
	bson.Unmarshal(doc, op.Selector)

	// Update
	change := map[string]bool{
		"updated": true,
	}
	update := bson.D{{"$set", change}}
	doc, err = bson.Marshal(update)
	if err != nil {
		t.Error(err)
	}
	op.Update = &bson.Raw{}
	bson.Unmarshal(doc, op.Update)

	op.Collection = "mongoreplay_test.test"
	op.Flags = 12345

	t.Logf("Generated Update: %#v\n", op.UpdateOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.UpdateOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}

	updateOp := receivedOp.(*UpdateOp)
	t.Log("Comparing parsed Update to original Update")
	switch {
	case updateOp.Collection != op.Collection:
		t.Errorf("Collections not equal. Saw %v -- Expected %v\n", updateOp.Collection, op.Collection)
	case !reflect.DeepEqual(updateOp.Selector, op.Selector):
		t.Errorf("Selectors not equal. Saw %v -- Expected %v\n", updateOp.Selector, op.Selector)
	case !reflect.DeepEqual(updateOp.Update, op.Update):
		t.Errorf("Updates not equal. Saw %v -- Expected %v\n", updateOp.Update, op.Update)
	case updateOp.Flags != op.Flags:
		t.Errorf("Flags not equal. Saw %d -- Expected %d\n", updateOp.Flags, op.Flags)
	}
}

func TestCommandOp(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := CommandOp{}
	op.Database = "foo"
	op.CommandName = "query"
	// Metadata
	metadata := map[string]int{
		"metadata": 1,
	}
	doc, err := bson.Marshal(metadata)
	if err != nil {
		t.Error(err)
	}
	op.Metadata = &bson.Raw{}
	bson.Unmarshal(doc, op.Metadata)

	//CommandArgs
	change := map[string]bool{
		"updated": true,
	}
	commandArgs := bson.D{{"$set", change}}
	doc, err = bson.Marshal(commandArgs)
	if err != nil {
		t.Error(err)
	}
	op.CommandArgs = &bson.Raw{}
	bson.Unmarshal(doc, op.CommandArgs)

	//InputDocs
	inputDocs := []interface{}{}
	for i := 0; i < 5; i++ {
		inputDocMap := map[string]bool{
			"updated": true,
		}
		inputDoc := &bson.Raw{}
		doc, err := bson.Marshal(inputDocMap)
		if err != nil {
			t.Error(err)
		}
		bson.Unmarshal(doc, inputDoc)
		inputDocs = append(inputDocs, inputDoc)
	}

	op.InputDocs = inputDocs

	t.Logf("Generated CommandOp: %#v\n", op.CommandOp)

	result, err := generator.fetchRecordedOpsFromConn(&op.CommandOp)
	if err != nil {
		t.Error(err)
	}
	receivedOp, err := result.RawOp.Parse()
	if err != nil {
		t.Error(err)
	}

	commandOp := receivedOp.(*CommandOp)

	t.Log("Comparing parsed Command to original Command")
	switch {
	case commandOp.Database != op.Database:
		t.Errorf("Databases not equal. Saw %v -- Expected %v\n", commandOp.Database, op.Database)
	case commandOp.CommandName != op.CommandName:
		t.Errorf("CommandNames not equal. Saw %v -- Expected %v\n", commandOp.CommandName, op.CommandName)
	case !reflect.DeepEqual(commandOp.Metadata, op.Metadata):
		t.Errorf("Metadata not equal. Saw %v -- Expected %v\n", commandOp.Metadata, op.Metadata)
	case !reflect.DeepEqual(commandOp.CommandArgs, op.CommandArgs):
		t.Errorf("CommandArgs not equal. Saw %v -- Expected %v\n", commandOp.CommandArgs, op.CommandArgs)
	}
	for i, doc := range commandOp.InputDocs {
		marshaledAsBytes, _ := bson.Marshal(inputDocs[i])
		unmarshaled := &bson.Raw{}
		bson.Unmarshal(marshaledAsBytes, unmarshaled)
		if !reflect.DeepEqual(unmarshaled, doc) {
			t.Errorf("Document from InputDocs not matched. Saw %v -- Expected %v\n", unmarshaled, doc)
		}
	}
}

func TestPreciseTimeMarshal(t *testing.T) {
	t1 := time.Date(2015, 4, 8, 15, 16, 23, 651387237, time.UTC)
	preciseTime := &PreciseTime{t1}
	asBson, err := bson.Marshal(preciseTime)
	if err != nil {
		t.Error(err)
	}
	result := &PreciseTime{}
	err = bson.Unmarshal(asBson, result)
	if err != nil {
		t.Error(err)
	}

	if t1 != result.Time {
		t.Errorf("Times not equal. Input: %v -- Result: %v", t1, result.Time)
	}
}

func TestCommandOpGetMoreCursorsRewriteable(t *testing.T) {
	oldCursorID := int64(1234)
	newCursorID := int64(5678)

	commandGM := &CommandGetMore{
		CommandOp: CommandOp{},
	}

	doc := &bson.D{{"getMore", oldCursorID}}

	asByte, err := bson.Marshal(doc)
	if err != nil {
		t.Errorf("could not marshal bson: %v", err)
	}
	asRaw := &bson.Raw{}
	bson.Unmarshal(asByte, asRaw)
	commandGM.CommandOp.CommandArgs = asRaw

	t.Log("fetching getmore cursorID")
	cursorIDs, err := commandGM.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != 1 {
		t.Errorf("differing number of cursorIDs found in commandlgetmore. Expected: %v --- Found: %v", 1, len(cursorIDs))
	} else {
		if oldCursorID != cursorIDs[0] {
			t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", oldCursorID, cursorIDs[0])
		}
	}

	t.Log("setting getmore cursorID")
	err = commandGM.setCursorIDs([]int64{newCursorID})
	if err != nil {
		t.Errorf("error setting cursorIDs: %v", err)
	}

	t.Log("fetching new getmore cursorID")
	cursorIDs, err = commandGM.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != 1 {
		t.Errorf("differing number of cursorIDs found in killcursors. Expected: %v --- Found: %v", 1, len(cursorIDs))
	} else {
		if newCursorID != cursorIDs[0] {
			t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", newCursorID, cursorIDs[0])
		}
	}
	commandArgs, ok := commandGM.CommandOp.CommandArgs.(*bson.D)
	if !ok {
		t.Errorf("commandArgs not a *bson.D")
	} else {
		for _, bsonDoc := range *commandArgs {
			if bsonDoc.Name == "getMore" {
				getmoreID, ok := bsonDoc.Value.(int64)
				if !ok {
					t.Errorf("cursorID in command is not int64")
				}
				if newCursorID != getmoreID {
					t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", newCursorID, getmoreID)
				}
				break
			}
		}
	}
}

func TestOpGetMoreCursorsRewriteable(t *testing.T) {
	oldCursorID := int64(1234)
	newCursorID := int64(5678)

	gm := &GetMoreOp{}
	gm.CursorId = oldCursorID

	t.Log("fetching getmore cursorID")
	cursorIDs, err := gm.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != 1 {
		t.Errorf("differing number of cursorIDs found in getmore. Expected: %v --- Found: %v", 1, len(cursorIDs))
	} else {
		if oldCursorID != cursorIDs[0] {
			t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", oldCursorID, cursorIDs[0])
		}
	}

	t.Log("setting getmore cursorID")
	err = gm.setCursorIDs([]int64{newCursorID})
	if err != nil {
		t.Errorf("error setting cursorIDs: %v", err)
	}

	t.Log("fetching new getmore cursorID")
	cursorIDs, err = gm.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != 1 {
		t.Errorf("differing number of cursorIDs found in killcursors. Expected: %v --- Found: %v", 1, len(cursorIDs))
	} else {
		if newCursorID != cursorIDs[0] {
			t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", newCursorID, cursorIDs[0])
		}
	}

}

func TestKillCursorsRewriteable(t *testing.T) {
	oldCursorIDs := []int64{11, 12, 13}
	newCursorIDs := []int64{21, 22}

	kc := &KillCursorsOp{}
	kc.CursorIds = oldCursorIDs

	cursorIDs, err := kc.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != len(oldCursorIDs) {
		t.Errorf("differing number of cursorIDs found in killcursors. Expected: %v --- Found: %v", len(oldCursorIDs), len(cursorIDs))
	} else {
		for i := range cursorIDs {
			if oldCursorIDs[i] != cursorIDs[i] {
				t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", oldCursorIDs[i], cursorIDs[i])
			}
		}
	}

	err = kc.setCursorIDs(newCursorIDs)
	if err != nil {
		t.Errorf("error setting cursorIDs: %v", err)
	}

	cursorIDs, err = kc.getCursorIDs()
	if err != nil {
		t.Errorf("error fetching cursorIDs: %v", err)
	}
	if len(cursorIDs) != len(newCursorIDs) {
		t.Errorf("differing number of cursorIDs found in killcursors. Expected: %v --- Found: %v", len(oldCursorIDs), len(cursorIDs))
	} else {
		for i := range cursorIDs {
			if newCursorIDs[i] != cursorIDs[i] {
				t.Errorf("cursorIDs not matched when retrieved. Expected: %v --- Found: %v", newCursorIDs[i], cursorIDs[i])
			}
		}
	}
}

func TestOpCommandReplyGetCursorID(t *testing.T) {
	testCursorID := int64(123)
	doc := &struct {
		Cursor struct {
			ID int64 `bson:"id"`
		} `bson:"cursor"`
	}{}
	doc.Cursor.ID = testCursorID
	asByte, err := bson.Marshal(doc)
	if err != nil {
		t.Errorf("could not marshal bson: %v", err)
	}
	asRaw := &bson.Raw{}
	bson.Unmarshal(asByte, asRaw)

	commandReplyOp := &CommandReplyOp{}
	commandReplyOp.CommandReply = asRaw
	cursorID, err := commandReplyOp.getCursorID()
	if err != nil {
		t.Errorf("error fetching cursor %v", err)
	}
	if cursorID != testCursorID {
		t.Errorf("cursorID did not match expected. Found: %v --- Expected: %v", cursorID, testCursorID)
	}

	t.Log("Ensuring cursorID consistent between multiple calls")
	cursorID, err = commandReplyOp.getCursorID()
	if err != nil {
		t.Errorf("error fetching cursor %v", err)
	}
	if cursorID != testCursorID {
		t.Errorf("cursorID did not match expected. Found: %v --- Expected: %v", cursorID, testCursorID)
	}
}

func TestShortenLegacyReply(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := ReplyOp{}
	op.ReplyDocs = 2

	result, err := generator.fetchRecordedOpsFromConn(&op.ReplyOp)

	doc1 := &testDoc{
		Name:           "Op Raw Short Reply Test 1",
		DocumentNumber: 1,
		Success:        true,
	}
	doc2 := &testDoc{
		Name:           "Op Raw Short Reply Test 2",
		DocumentNumber: 2,
		Success:        true,
	}

	asByte1, err := bson.Marshal(doc1)
	if err != nil {
		t.Errorf("could not marshal bson: %v", err)
	}

	asByte2, err := bson.Marshal(doc2)
	if err != nil {
		t.Errorf("could not marshal bson: %v", err)
	}

	opBody := result.RawOp.Body
	// add the two docs as the docs from the reply
	opBody = append(opBody, asByte1...)
	opBody = append(opBody, asByte2...)
	// set the size in the header
	setInt32(opBody, 0, int32(len(opBody)))

	recordedOp := &RecordedOp{}

	recordedOp.BodyFromSlice(opBody)
	if err != nil {
		t.Error(err)
	}

	parsed, err := recordedOp.RawOp.Parse()
	if err != nil {
		t.Errorf("error parsing op: %v", err)
	}

	fullReply, ok := parsed.(*ReplyOp)
	if !ok {
		t.Errorf("parsed op was wrong type")
	}

	if !(len(fullReply.Docs) == 2) {
		t.Errorf("parsed reply has wrong number of docs: %d", len(fullReply.Docs))
	}

	// shorten the reply
	recordedOp.ShortenReply()

	bodySlice := make([]byte, 0)
	// Write it out to a slice
	buf := bytes.NewBuffer(bodySlice)

	err = recordedOp.ToWriter(buf)
	if err != nil {
		t.Errorf("unable to write out to writer")
	}

	recordedOpResult := &RecordedOp{}

	err = recordedOpResult.FromReader(buf)
	if err != nil {
		t.Errorf("unable to read op from buffer %v", err)
	}

	parsed, err = recordedOpResult.RawOp.Parse()
	if err != nil {
		t.Errorf("error parsing op: %v", err)
	}

	fullReply, ok = parsed.(*ReplyOp)
	if !ok {
		t.Errorf("parsed op was wrong type")
	}

	// ensure that the reply now has only 1 document
	if !(len(fullReply.Docs) == 1) {
		t.Errorf("parsed reply has wrong number of docs: %d", len(fullReply.Docs))
	}
}

type cursorDoc struct {
	Batch []interface{} `bson:"firstBatch"`
	Id    int64         `bson:"id"`
	Ns    string        `bson:"ns"`
}
type findReply struct {
	Cursor cursorDoc `bson:"cursor"`
	Ok     int       `bson:"ok"`
}

func TestShortenCommandReply(t *testing.T) {
	generator := newRecordedOpGenerator()

	op := CommandReplyOp{}
	op.Metadata = &testDoc{
		Name:           "Metadata",
		DocumentNumber: 100000,
		Success:        true,
	}

	doc1 := testDoc{
		Name:           "Op Raw Short Reply Test 1",
		DocumentNumber: 1,
		Success:        true,
	}
	doc2 := testDoc{
		Name:           "Op Raw Short Reply Test 2",
		DocumentNumber: 2,
		Success:        true,
	}

	batch := []interface{}{doc1, doc2}

	cursorDocIn := cursorDoc{
		batch, 12345678, "test"}

	op.CommandReply = findReply{cursorDocIn, 1}
	op.OutputDocs = []interface{}{}

	result, err := generator.fetchRecordedOpsFromConn(&op.CommandReplyOp)

	// reply should be functional and parseable
	parsed, err := result.RawOp.Parse()
	if err != nil {
		t.Errorf("error parsing op: %#v", err)
	}

	t.Logf("parsed Op: %v", parsed)

	fullReply, ok := parsed.(*CommandReplyOp)
	if !ok {
		t.Errorf("parsed op was wrong type")
	}

	commandReplyCheckRaw, ok := fullReply.CommandReply.(*bson.Raw)
	if !ok {
		t.Errorf("comamndReply not bson.Raw")
	}

	commandReplyCheck := &findReply{
		Cursor: cursorDoc{},
	}
	err = bson.Unmarshal(commandReplyCheckRaw.Data, commandReplyCheck)
	if err != nil {
		t.Errorf("error unmarshaling commandReply %v", err)
	}

	// ensure that the reply now has 2 document
	if !(len(commandReplyCheck.Cursor.Batch) == 2) {
		t.Errorf("parsed reply has wrong number of docs: %d", len(commandReplyCheck.Cursor.Batch))
	}

	// shorten the reply
	result.ShortenReply()

	parsed, err = result.RawOp.Parse()
	if err != nil {
		t.Errorf("error parsing op: %v", err)
	}

	fullReply, ok = parsed.(*CommandReplyOp)
	if !ok {
		t.Errorf("parsed op was wrong type")
	}

	commandReplyRaw, ok := fullReply.CommandReply.(*bson.Raw)
	if !ok {
		t.Errorf("comamndReply not bson.Raw")
	}

	commandReplyOut := &findReply{
		Cursor: cursorDoc{},
	}
	err = bson.Unmarshal(commandReplyRaw.Data, commandReplyOut)
	if err != nil {
		t.Errorf("error unmarshaling commandReply %v", err)
	}

	// ensure that the reply now has 0 documents
	if !(len(commandReplyOut.Cursor.Batch) == 0) {
		t.Errorf("parsed reply has wrong number of docs: %d", len(commandReplyOut.Cursor.Batch))
	}
}

func TestLegacyOpReplyGetCursorID(t *testing.T) {
	testCursorID := int64(123)
	doc := &struct {
		Cursor struct {
			ID int64 `bson:"id"`
		} `bson:"cursor"`
	}{}
	doc.Cursor.ID = testCursorID
	asByte, err := bson.Marshal(doc)
	if err != nil {
		t.Errorf("could not marshal bson: %v", err)
	}
	asRaw := bson.Raw{}
	bson.Unmarshal(asByte, &asRaw)

	reply := &ReplyOp{}
	reply.Docs = []bson.Raw{asRaw}

	t.Log("Retrieving cursorID from reply docs")
	cursorID, err := reply.getCursorID()
	if err != nil {
		t.Errorf("error fetching cursor %v", err)
	}
	if cursorID != testCursorID {
		t.Errorf("cursorID did not match expected. Found: %v --- Expected: %v", cursorID, testCursorID)
	}

	t.Log("Ensuring cursorID consistent between multiple calls")
	cursorID, err = reply.getCursorID()
	if err != nil {
		t.Errorf("error fetching cursor %v", err)
	}
	if cursorID != testCursorID {
		t.Errorf("cursorID did not match expected. Found: %v --- Expected: %v", cursorID, testCursorID)
	}

	reply2 := &ReplyOp{}
	reply2.CursorId = testCursorID

	t.Log("Retrieving cursorID from reply field")
	cursorID, err = reply.getCursorID()
	if err != nil {
		t.Errorf("error fetching cursor %v", err)
	}
	if cursorID != testCursorID {
		t.Errorf("cursorID did not match expected. Found: %v --- Expected: %v", cursorID, testCursorID)
	}

}
