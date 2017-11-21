package eventuate

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Int128 [2]uint64

var Int128Nil Int128

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func Int128FromString(input string) Int128 {
	var tmp [2]uint64
	parts := strings.Split(input, "-")
	if len(parts) == 2 {
		var (
			err        error
			tmp0, tmp1 uint64
		)
		tmp0, err = strconv.ParseUint(parts[0], 16, 64)
		if err != nil {
			return tmp
		}
		tmp1, err = strconv.ParseUint(parts[1], 16, 64)
		if err != nil {
			return tmp
		}
		tmp[0] = tmp0
		tmp[1] = tmp1
	}
	return tmp
}

func Int128Random() Int128 {
	var tmp [2]uint64
	tmp[0] = uint64(r.Int63())
	tmp[1] = uint64(time.Now().UnixNano() / int64(time.Millisecond)) // uint64(r.Int63()) //
	return tmp
}

func (id Int128) IsNil() bool {
	return id == Int128Nil
}

func (id Int128) String() string {
	tmp := [2]uint64(id)
	return fmt.Sprintf("%016x-%016x", tmp[0], tmp[1])
}

func (id Int128) FirstPart() uint64 {
	return [2]uint64(id)[0]
}
func (id Int128) LastPart() uint64 {
	return [2]uint64(id)[1]
}

// thanks to https://gist.github.com/mdwhatcott/8dd2eef0042f7f1c0cd8
func (id Int128) MarshalJSON() ([]byte, error) {
	value := id.String()
	return json.Marshal(value)
}

func (id *Int128) UnmarshalJSON(b []byte) error {
	src, err := unmarshalJSON(b)
	if err != nil {
		return err
	}
	((*[2]uint64)(id))[0] = src[0]
	((*[2]uint64)(id))[1] = src[1]
	return nil
}

func unmarshalJSON(b []byte) (id *Int128, err error) {
	defer func() {
		if r := recover(); r != nil {
			var tmp2 uint64
			err = json.Unmarshal(b, &tmp2)

			result := Int128FromString(fmt.Sprintf("0-%d", tmp2))

			id = &result
		}
	}()
	var tmp string
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		var tmp2 uint64
		err = json.Unmarshal(b, &tmp2)

		result := Int128FromString(fmt.Sprintf("0-%d", tmp2))

		id = &result
		return id, err
	}
	result := Int128FromString(tmp)
	return &result, nil
}
