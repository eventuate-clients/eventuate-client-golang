package eventuate_test

import (
	"fmt"
	"github.com/eventuate-clients/eventuate-client-golang"
	"github.com/eventuate-clients/eventuate-client-golang/future"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
	"math"
	"testing"
	"time"
)

const AGG_TYPE = "aggregateType"
const EVT_TYPE_A = "eventTypeA"
const EVT_TYPE_B = "eventTypeB"
const SLEEP_A_MS = 500
const SLEEP_B_MS = 700

var logLevel = loglib.Silent
var lg = loglib.NewLogger(logLevel)

type testCase int

const (
	FirstEventSetup testCase = iota
	SecondEventSetup
	ThirdEventSetup
	FourthEventSetup
)

type eventDataAndMeta struct {
	data interface{}
	meta *eventuate.EventMetadata
}

func MakeTestEventsSlice(setup testCase) []eventDataAndMeta {
	switch setup {
	case FirstEventSetup:
		{
			return []eventDataAndMeta{
				*newEvent(AGG_TYPE, EVT_TYPE_A, "1", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_A, "2", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_B, "3", 1)}
		}
	case SecondEventSetup:
		{
			return []eventDataAndMeta{
				*newEvent(AGG_TYPE, EVT_TYPE_A, "1", 1),

				*newEvent(AGG_TYPE, EVT_TYPE_A, "4", 2)}
		}
	case ThirdEventSetup:
		{
			return []eventDataAndMeta{
				*newEvent(AGG_TYPE, EVT_TYPE_A, "1", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_A, "2", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_A, "3", 1),

				*newEvent(AGG_TYPE, EVT_TYPE_B, "4", 2),
				*newEvent(AGG_TYPE, EVT_TYPE_B, "5", 2)}
		}
	case FourthEventSetup:
		{
			return []eventDataAndMeta{
				*newEvent(AGG_TYPE, EVT_TYPE_A, "1", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_B, "2", 1),
				*newEvent(AGG_TYPE, EVT_TYPE_A, "3", 1),

				*newEvent(AGG_TYPE, EVT_TYPE_B, "4", 2),
				*newEvent(AGG_TYPE, EVT_TYPE_A, "5", 2),
				*newEvent(AGG_TYPE, EVT_TYPE_B, "6", 2)}
		}
	}
	return []eventDataAndMeta{}
}

func TestEventTypeSwimlaneDispatcher_Dispatch_1(t *testing.T) {

	bunch := MakeTestEventsSlice(FirstEventSetup) // A-A-B
	dispatcher := getCommonDispatcher(len(bunch), SLEEP_A_MS, SLEEP_B_MS)
	started := time.Now()
	results := make([]future.Settler, len(bunch))
	for idx, evt := range bunch {
		lg.Println("Dispatching..", evt)
		results[idx] = dispatcher.Dispatch(evt.data, evt.meta)
	}

	lg.Println("Dispatched all bunch, waiting for all to finish")

	future.WhenAll(results...).Then(func(val interface{}, err error) (interface{}, error) {
		elapsedMs := getElapsedMsRound100(started)
		if elapsedMs != SLEEP_A_MS+SLEEP_A_MS+SLEEP_B_MS {
			t.Fail()
		}
		lg.Println("Done  in.. ", elapsedMs, val, err)
		return nil, nil
	}).GetValue()

}

func TestEventTypeSwimlaneDispatcher_Dispatch_2(t *testing.T) {

	events := MakeTestEventsSlice(SecondEventSetup) // A|A => Max(A, A)
	dispatcher := getCommonDispatcher(len(events), SLEEP_A_MS, SLEEP_B_MS)
	started := time.Now()
	results := make([]future.Settler, len(events))
	for idx, evt := range events {
		lg.Println("Dispatching..", evt)
		results[idx] = dispatcher.Dispatch(evt.data, evt.meta)
	}

	lg.Println("Dispatched all events, waiting for all to finish")

	future.WhenAll(results...).Then(func(interface{}, error) (interface{}, error) {
		elapsedMs := getElapsedMsRound100(started)
		if elapsedMs != int(math.Max(SLEEP_A_MS, SLEEP_A_MS)) {
			t.Fail()
		}
		lg.Println("Done  in.. ", elapsedMs)
		return nil, nil
	}).GetValue()

}

func TestEventTypeSwimlaneDispatcher_Dispatch_3(t *testing.T) {

	events := MakeTestEventsSlice(ThirdEventSetup) // A-A-A|B-B => Max(A+A+A, B+B)
	dispatcher := getCommonDispatcher(len(events), SLEEP_A_MS, SLEEP_B_MS)
	started := time.Now()
	results := make([]future.Settler, len(events))
	for idx, evt := range events {
		lg.Println("Dispatching..", evt)
		results[idx] = dispatcher.Dispatch(evt.data, evt.meta)
	}

	lg.Println("Dispatched all events, waiting for all to finish")

	future.WhenAll(results...).Then(func(interface{}, error) (interface{}, error) {
		elapsedMs := getElapsedMsRound100(started)
		if elapsedMs != int(math.Max(SLEEP_A_MS+SLEEP_A_MS+SLEEP_A_MS, SLEEP_B_MS+SLEEP_B_MS)) {
			t.Fail()
		}
		lg.Println("Done  in.. ", elapsedMs)
		return nil, nil
	}).GetValue()

}

func TestEventTypeSwimlaneDispatcher_Dispatch_4(t *testing.T) {

	events := MakeTestEventsSlice(FourthEventSetup) // A-B-A|B-A-B => Max(A+B+A, B+A+B)
	dispatcher := getCommonDispatcher(len(events), SLEEP_A_MS, SLEEP_B_MS)
	started := time.Now()
	results := make([]future.Settler, len(events))
	for idx, evt := range events {
		lg.Println("Dispatching..", evt)
		results[idx] = dispatcher.Dispatch(evt.data, evt.meta)
	}

	lg.Println("Dispatched all events, waiting for all to finish")

	future.WhenAll(results...).Then(func(interface{}, error) (interface{}, error) {
		elapsedMs := getElapsedMsRound100(started)
		if elapsedMs != int(math.Max(SLEEP_A_MS+SLEEP_B_MS+SLEEP_A_MS, SLEEP_B_MS+SLEEP_A_MS+SLEEP_B_MS)) {
			t.Fail()
		}
		lg.Println("Done  in.. ", elapsedMs)
		return nil, nil
	}).GetValue()

}

func getCommonDispatcher(eventCount int, sleepA, sleepB int) eventuate.Dispatcher {
	var (
		wgCount int
	)
	wgCount = eventCount

	commonCb := func(val interface{}, err error) (interface{}, error) {
		lg.Println("commonCb() Decreasing lock counter", val, err)
		wgCount--
		lg.Println("commonCb() Decreased lock counter. Now ", wgCount)

		if wgCount == 0 {
			lg.Println("commonCb() Ending lock counter")
		}

		return wgCount, nil
	}

	handlersMap := eventuate.NewEventResultHandlerMap()
	handlersMap.AddHandler(AGG_TYPE, EVT_TYPE_A, getTestHandler(0, sleepA, commonCb))
	handlersMap.AddHandler(AGG_TYPE, EVT_TYPE_B, getTestHandler(1, sleepB, commonCb))
	dispatcher, _ := eventuate.NewEventTypeSwimlaneDispatcher(handlersMap)

	if dsp, isDsp := dispatcher.(*eventuate.EventTypeSwimlaneDispatcher); isDsp {
		dsp.SetLogLevel(logLevel)
	}

	return dispatcher
}

func getTestHandler(idx int, wait int, cb future.ThenCallback) func(interface{}, *eventuate.EventMetadata) future.Settler {
	return func(data interface{}, evt *eventuate.EventMetadata) future.Settler {
		str := fmt.Sprintf("Event: idx:%v - data:%v", idx, data)
		fr := future.NewTimedResult(time.Duration(wait)*time.Millisecond, str, nil)
		fr.SetLogLevel(logLevel)

		return fr.Then(cb)
	}
}

func getElapsedMsRound100(started time.Time) int {
	elapsed := time.Since(started)
	return int(int(elapsed.Seconds()*1000)/100.0) * 100
}

func newEvent(aggType, eventType string, data interface{}, swimLane int) *eventDataAndMeta {
	return &eventDataAndMeta{
		data: data,
		meta: &eventuate.EventMetadata{
			Id:         eventuate.Int128Nil,
			EntityId:   eventuate.Int128Nil,
			EntityType: aggType,
			EventType:  eventType,
			SwimLane:   swimLane}}
}
