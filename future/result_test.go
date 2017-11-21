package future_test

import (
	"errors"
	"github.com/shopcookeat/eventuate-client-golang/future"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewSuccess(t *testing.T) {
	resultMsg := "resulting message"
	passed := future.NewSuccess(resultMsg)

	startTime := time.Now()

	if !passed.IsSettled() {
		t.Fail()
		return
	}

	expVal, expErr := passed.GetValue()

	if expErr != nil {
		t.Fail()
		return
	}

	if !reflect.DeepEqual(expVal, resultMsg) {
		t.Fail()
		return
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .01 {
		t.Fail()
	}
}

func TestNewFailure(t *testing.T) {
	errorMsg := errors.New("error message")
	failed := future.NewFailure(errorMsg)

	startTime := time.Now()

	if !failed.IsSettled() {
		t.Fail()
		return
	}

	expVal, expErr := failed.GetValue()

	if expVal != nil {
		t.Fail()
		return
	}

	if expErr != errorMsg {
		t.Fail()
		return
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .01 {
		t.Fail()
	}
}

func TestNewResult(t *testing.T) {
	fr := future.NewResult()
	resultMsg := "resulting message"

	startTime := time.Now()

	if fr.IsSettled() {
		t.Fail()
		return
	}

	var wg sync.WaitGroup

	go func() {
		expVal, expErr := fr.GetValue()

		if expErr != nil {
			t.Fail()
			return
		}

		if !reflect.DeepEqual(expVal, resultMsg) {
			t.Fail()
			return
		}

		wg.Done()

	}()

	wg.Add(1)

	time.Sleep(100 * time.Millisecond)

	fr.Settle(resultMsg, nil)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestNewResult2(t *testing.T) {
	fr := future.NewResult()
	errorMsg := errors.New("error message")

	startTime := time.Now()

	if fr.IsSettled() {
		t.Fail()
		return
	}

	var wg sync.WaitGroup

	go func() {
		expVal, expErr := fr.GetValue()

		if expVal != nil {
			t.Fail()
			return
		}

		if expErr != errorMsg {
			t.Fail()
			return
		}

		wg.Done()

	}()

	wg.Add(1)

	time.Sleep(100 * time.Millisecond)

	fr.Settle(nil, errorMsg)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestNewResult3(t *testing.T) {
	fr := future.NewResult()
	errorMsg := errors.New("error message")

	startTime := time.Now()

	if fr.IsSettled() {
		t.Fail()
		return
	}

	var wg sync.WaitGroup

	go func() {
		expVal, expErr := fr.GetValue()

		if expVal != nil {
			t.Fail()
			return
		}

		if expErr != errorMsg {
			t.Fail()
			return
		}

		wg.Done()

	}()

	wg.Add(1)

	time.Sleep(100 * time.Millisecond)

	fr.Settle("Some message", errorMsg)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestResult_Timed(t *testing.T) {
	errorMsg := errors.New("error message")
	dur := time.Duration(1000) * time.Millisecond
	durPlus := time.Duration(1100) * time.Millisecond

	fr := future.NewResult()
	startTime := time.Now()
	fr = fr.Timed(dur, nil, errorMsg)

	if fr.IsSettled() {
		t.Fail()
		return
	}

	time.Sleep(dur)

	_, err := fr.GetValue()
	elapsed := time.Since(startTime)

	if !reflect.DeepEqual(errorMsg, err) {
		t.Fail()
		return
	}

	if elapsed.Seconds() > durPlus.Seconds() {
		t.Fail()
	}
	if elapsed.Seconds() < dur.Seconds() {
		t.Fail()
	}
}

func TestNewTimedResult(t *testing.T) {
	errorMsg := errors.New("error message")
	dur := time.Duration(1000) * time.Millisecond
	durPlus := time.Duration(1100) * time.Millisecond

	fr := future.NewTimedResult(dur, nil, errorMsg)
	startTime := time.Now()

	if fr.IsSettled() {
		t.Fail()
		return
	}

	time.Sleep(dur)

	_, err := fr.GetValue()
	elapsed := time.Since(startTime)

	if !reflect.DeepEqual(errorMsg, err) {
		t.Fail()
		return
	}

	if elapsed.Seconds() > durPlus.Seconds() {
		t.Fail()
	}
	if elapsed.Seconds() < dur.Seconds() {
		t.Fail()
	}

}

func TestResult_IsSettled(t *testing.T) {
	fr := future.NewResult()
	resultMsg := "resulting message"

	startTime := time.Now()

	if fr.IsSettled() {
		t.Fail()
		return
	}

	var wg sync.WaitGroup

	go func() {
		fr.GetValue()

		if !fr.IsSettled() {
			t.Fail()
			return
		}

		wg.Done()
	}()

	wg.Add(1)

	time.Sleep(100 * time.Millisecond)

	fr.Settle(resultMsg, nil)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestResult_GetValue(t *testing.T) {
	resultMsg := "resulting message"
	passed := future.NewSuccess(resultMsg)

	startTime := time.Now()

	if passed.IsSettled() {
		passed.GetValue()
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .01 {
		t.Fail()
	}

}
func TestResult_GetValue2(t *testing.T) {

	errorMsg := errors.New("error message")
	failed := future.NewFailure(errorMsg)

	startTime := time.Now()

	if failed.IsSettled() {
		failed.GetValue()
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .01 {
		t.Fail()
	}
}
func TestResult_GetValue3(t *testing.T) {
	fr := future.NewResult()
	resultMsg := "resulting message"

	startTime := time.Now()
	var wg sync.WaitGroup

	go func() {
		expVal, _ := fr.GetValue()

		if !reflect.DeepEqual(expVal, resultMsg) {
			t.Fail()
			return
		}

		wg.Done()
	}()

	go func() {
		expVal, _ := fr.GetValue()

		if !reflect.DeepEqual(expVal, resultMsg) {
			t.Fail()
			return
		}

		wg.Done()
	}()

	wg.Add(2)

	time.Sleep(100 * time.Millisecond)

	fr.Settle(resultMsg, nil)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestResult_Settle(t *testing.T) {
	fr := future.NewResult()
	resultMsg := "resulting message"

	startTime := time.Now()
	var wg sync.WaitGroup

	go func() {
		defer wg.Done()

		if fr.IsSettled() {
			t.Fail()
			return
		}

		time.Sleep(100 * time.Millisecond)
		if fr.IsSettled() {
			t.Fail()
			return
		}

		time.Sleep(100 * time.Millisecond)
		if !fr.IsSettled() {
			t.Fail()
		}

	}()

	wg.Add(1)

	time.Sleep(150 * time.Millisecond)

	fr.Settle(resultMsg, nil)

	wg.Wait()
	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .3 {
		t.Fail()
	}
}

func TestResult_Then(t *testing.T) {
	resultMsg := "resulting message"
	resultMsg2 := "resulting message 2"
	passed := future.NewSuccess(resultMsg)

	startTime := time.Now()

	expVal, expErr := passed.Then(func(val1 interface{}, err1 error) (interface{}, error) {
		if err1 != nil {
			t.Fail()
		}

		if !reflect.DeepEqual(val1, resultMsg) {
			t.Fail()
		}

		return resultMsg2, nil
	}).GetValue()

	if expErr != nil {
		t.Fail()
		return
	}

	if !reflect.DeepEqual(expVal, resultMsg2) {
		t.Fail()
		return
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .01 {
		t.Fail()
	}
}

func TestResult_Then2(t *testing.T) {
	resultMsg := "resulting message"
	resultMsg2 := "resulting message 2"
	passed := future.NewSuccess(resultMsg)

	startTime := time.Now()

	thenFr := passed.Then(func(val1 interface{}, err1 error) (interface{}, error) {

		inside := future.NewResult()
		func() {
			time.Sleep(100 * time.Millisecond)
			inside.Settle(resultMsg2, nil)
		}()

		return inside, nil
	})

	if thenFr.IsSettled() {
		t.Fail()
		return
	}

	expVal, expErr := thenFr.GetValue()

	if expErr != nil {
		t.Fail()
		return
	}

	if !reflect.DeepEqual(expVal, resultMsg2) {
		t.Fail()
		return
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() > .2 {
		t.Fail()
	}
}

func TestWhenAll(t *testing.T) {
	resultMsg := "resulting message"
	resultMsg2 := "resulting message 2"

	startTime := time.Now()

	fr1 := future.NewResult()
	go func() {
		time.Sleep(100 * time.Millisecond)
		fr1.Settle(resultMsg, nil)
	}()

	fr2 := future.NewResult()
	go func() {
		time.Sleep(200 * time.Millisecond)
		fr2.Settle(resultMsg2, nil)
	}()

	whenAll := future.WhenAll(fr1, fr2)

	vals, _ := whenAll.GetValue()

	arrVals, arrValsOk := vals.([]interface{})

	if arrValsOk {
		if len(arrVals) != 2 {
			t.Fail()
			return
		}

		if !reflect.DeepEqual(arrVals[0], resultMsg) {
			t.Fail()
			return
		}
		if !reflect.DeepEqual(arrVals[1], resultMsg2) {
			t.Fail()
			return
		}

	} else {
		t.Fail()
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() < .19 {
		t.Fail()
	}

	if elapsed.Seconds() > .22 {
		t.Fail()
	}
}
