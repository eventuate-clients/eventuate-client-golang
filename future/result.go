package future

import (
	"errors"
	"fmt"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
	"sync"
	"time"
)

type Settler interface {
	IsSettled() bool
	Settle(interface{}, error)
	GetValue() (interface{}, error)
}
type ThenCallback func(interface{}, error) (interface{}, error)

type Thenable interface {
	Then(ThenCallback) Settler
}

var idCounter int

type Result struct {
	sync.WaitGroup
	sync.RWMutex
	settleFlag   bool
	lastValue    interface{}
	lastError    error
	id           int
	pendingCount int
	ll           loglib.LogLevelEnum
	lg           loglib.Logger
	tl           bool
}

func (fr Result) String() string {
	return fmt.Sprintf("Result(#%v)", fr.id)
}

func NewResult() *Result {
	idCounter++

	fr := Result{
		settleFlag: false,
		lastValue:  nil,
		lastError:  nil,
		id:         idCounter,
		lg:         loglib.NewNilLogger()}

	return &fr
}

func NewTimedResult(wait time.Duration, val interface{}, err error) *Result {
	return NewResult().Timed(wait, val, err)
}

func (fr *Result) Timed(wait time.Duration, val interface{}, err error) *Result {
	if fr.tl {
		return fr
	}

	fr.log("NewTimedResult created")
	go func() {
		fr.log("NewTimedResult (go) Sleeping for ", time.Duration(wait)*time.Millisecond)
		time.Sleep(wait) // time.Duration(wait) * time.Millisecond)
		fr.log("NewTimedResult (go) Un-sleeping from ", wait)
		fr.log("NewTimedResult (go) Settling.. ", fr, val, err)
		fr.Settle(val, err)
		fr.log("NewTimedResult (go) Settled.. ", fr, val, err)
	}()
	return fr
}

func (fr *Result) SetLogLevel(level loglib.LogLevelEnum) {
	fr.lg = loglib.NewLogger(level)
	fr.ll = level
}

func (fr *Result) log(args ...interface{}) {
	fr.lg.Println(args...)
}

func NewFailure(err error) *Result {
	return &Result{
		settleFlag: true,
		lastValue:  nil,
		lastError:  err,
	}
}

func NewSuccess(val interface{}) *Result {
	return &Result{
		settleFlag: true,
		lastValue:  val,
		lastError:  nil,
	}
}

func (fr *Result) IsSettled() bool {
	fr.RLock()
	defer fr.RUnlock()
	return fr.settleFlag
}

func (fr *Result) Settle(val interface{}, err error) {
	if fr.IsSettled() {
		return
	}
	fr.lg.Println("Result.Settle() Resolving for.. ", *fr)

	fr.Lock()
	defer fr.Unlock()
	defer fr.lg.Println("Result.Settle() .Unlock() called")

	fr.lastError = err
	if err == nil {
		fr.lastValue = val
	}
	fr.settleFlag = true
	fr.lg.Println("Result.Settle() Resolved. wg.Count: ", fr.pendingCount)
	if fr.pendingCount > 0 {
		defer func(count int) {
			fr.Add(count)
			fr.lg.Println("Result.Settle() wg.Add()", count)
		}(-fr.pendingCount)
		fr.pendingCount = 0
		fr.lg.Println("Result.Settle() wg.Count reset: ", fr.pendingCount)
	}
}

func (fr *Result) GetValue() (val interface{}, err error) {
	fr.RLock()
	if fr.settleFlag {
		defer fr.RUnlock()
		return fr.lastValue, fr.lastError
	}
	fr.RUnlock()

	fr.lg.Println("Result.GetValue() Expecting a value for.. ", *fr)

	fr.Lock()
	fr.pendingCount++
	fr.Add(1)
	fr.lg.Println("Result.GetValue() Increasing pendings, now: ", fr.pendingCount)
	fr.Unlock()

	fr.lg.Println("Result.GetValue() Blocked at the fr.Wait() ", *fr)
	fr.Wait()
	fr.lg.Println("Result.GetValue() Received a value for.. ", *fr, val, err)
	return fr.lastValue, fr.lastError
}

func (fr *Result) Then(cb ThenCallback) Settler {
	newFr := NewResult()
	go func(fr1, fr2 *Result) {
		val, err := (*fr1).GetValue() // may block here
		valNext, errNext := cb(val, err)
		if errNext == nil {
			frMid, frMidOk := valNext.(Settler)
			if frMidOk {
				valNext, errNext = frMid.GetValue()
			}
		}
		(*fr2).Settle(valNext, errNext)
	}(fr, newFr)
	return newFr
}

func WhenAll(frs ...Settler) *Result {
	newFr := NewResult()
	var wg sync.WaitGroup
	wg.Add(len(frs))
	values := make([]interface{}, len(frs))
	errors_ := make([]interface{}, len(frs))
	for idx, fr := range frs {
		go func(i int, fr1 Settler) {
			val, err := fr1.GetValue() // may block here
			values[i] = val
			errors_[i] = err
			wg.Done()
		}(idx, fr)
	}
	go func() {
		wg.Wait()
		for _, err := range errors_ {
			if err != nil {
				newFr.Settle(values, errors.New(fmt.Sprint(errors_...)))
				return
			}
		}
		newFr.Settle(values, nil)
	}()
	return newFr
}
