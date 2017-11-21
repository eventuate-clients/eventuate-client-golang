package eventuate_test

import (
	"encoding/json"
	"github.com/shopcookeat/eventuate-client-golang"
	"testing"
)

func TestInt128_String(t *testing.T) {
	actual := eventuate.Int128([2]uint64{1, 1}).String()
	expected := "0000000000000001-0000000000000001"
	if expected != actual {
		t.Fail()
	}
	actual = eventuate.Int128([2]uint64{0x15CC85BFDADFF, 0x15CC85BFDADFF}).String()
	expected = "00015cc85bfdadff-00015cc85bfdadff"
	if expected != actual {
		t.Fail()
	}
}

func TestInt128FromString(t *testing.T) {
	actual := eventuate.Int128FromString("0000000000000001-0000000000000001")
	expected := eventuate.Int128([2]uint64{1, 1})
	if expected != actual {
		t.Fail()
	}

	actual = eventuate.Int128FromString("00015cc85bfdadff-00015cc85bfdadff")
	expected = eventuate.Int128([2]uint64{0x15CC85BFDADFF, 0x15CC85BFDADFF})
	if expected != actual {
		t.Fail()
	}
}

func TestInt128_FirstPart(t *testing.T) {
	actual := eventuate.Int128FromString("0000000000000001-0000000000000002").FirstPart()
	var expected uint64 = 1
	if expected != actual {
		t.Fail()
	}
	actual = eventuate.Int128FromString("00015cc85bfdadff-00015cc85bfdadfe").FirstPart()
	expected = 0x15CC85BFDADFF
	if expected != actual {
		t.Fail()
	}
}

func TestInt128_LastPart(t *testing.T) {
	actual := eventuate.Int128FromString("0000000000000001-0000000000000002").LastPart()
	var expected uint64 = 2
	if expected != actual {
		t.Fail()
	}
	actual = eventuate.Int128FromString("00015cc85bfdadff-00015cc85bfdadfe").LastPart()
	expected = 0x15CC85BFDADFE
	if expected != actual {
		t.Fail()
	}
}

func TestInt128_MarshalJSON(t *testing.T) {

	type EventTypeAndData struct {
		EventType string `json:"eventType"`
		EventData string `json:"eventData"`
	}

	type EventIdTypeAndData struct {
		EventId eventuate.Int128 `json:"id"`
		EventTypeAndData
	}

	expected := `{"id":"0000015cc85bfdad-0242ac1101190003","eventType":"goodType","eventData":"payload"}`
	source := EventIdTypeAndData{
		EventId: eventuate.Int128FromString("0000015cc85bfdad-0242ac1101190003"),
		EventTypeAndData: EventTypeAndData{
			EventType: "goodType",
			EventData: "payload"}}

	result, err := json.Marshal(source)
	actual := string(result)

	if err != nil {
		t.Fail()
	}
	if expected != actual {
		t.Fail()
	}
}

func TestInt128_UnmarshalJSON(t *testing.T) {

	type EventTypeAndData struct {
		EventType string `json:"eventType"`
		EventData string `json:"eventData"`
	}

	type EventIdTypeAndData struct {
		EventId eventuate.Int128 `json:"id"`
		EventTypeAndData
	}

	source := `{"id":"0000015cc85bfdad-0242ac1101190003", "eventType":"goodType", "eventData":"payload"}`
	dst := &EventIdTypeAndData{}
	err := json.Unmarshal([]byte(source), dst)

	expected := EventIdTypeAndData{
		EventId: eventuate.Int128FromString("0000015cc85bfdad-0242ac1101190003"),
		EventTypeAndData: EventTypeAndData{
			EventType: "goodType",
			EventData: "payload"}}

	actual := *dst

	if err != nil {
		t.Fail()
	}
	if expected != actual {
		t.Fail()
	}

}
