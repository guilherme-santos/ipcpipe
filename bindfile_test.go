package ipcpipe_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/guilherme-santos/ipcpipe"

	"github.com/stretchr/testify/assert"
)

func testBindField(t *testing.T, psrv *ipcpipe.Server, path string, wg *sync.WaitGroup) {
	var (
		b bool
		// integer
		i   int
		ni  int // negative integer
		i8  int8
		i16 int16
		i32 int32
		i64 int64
		// unsigned integer
		ui   uint
		ui8  uint8
		ui16 uint16
		ui32 uint32
		ui64 uint64
		// float
		f32 float32
		f64 float64
		// string
		str string
	)

	testCases := []struct {
		Field   string
		Pointer interface{}
		Result  string
	}{
		{
			Field:   "test.boolean",
			Pointer: &b,
			Result:  "true",
		},
		{
			Field:   "test.integer",
			Pointer: &i,
			Result:  "1",
		},
		{
			Field:   "test.negative-integer",
			Pointer: &ni,
			Result:  "-100",
		},
		{
			Field:   "test.integer8",
			Pointer: &i8,
			Result:  "10",
		},
		{
			Field:   "test.integer16",
			Pointer: &i16,
			Result:  "100",
		},
		{
			Field:   "test.integer32",
			Pointer: &i32,
			Result:  "1000",
		},
		{
			Field:   "test.integer64",
			Pointer: &i64,
			Result:  "10000",
		},
		{
			Field:   "test.unsigned-integer",
			Pointer: &ui,
			Result:  "2",
		},
		{
			Field:   "test.unsigned-integer8",
			Pointer: &ui8,
			Result:  "20",
		},
		{
			Field:   "test.unsigned-integer16",
			Pointer: &ui16,
			Result:  "200",
		},
		{
			Field:   "test.unsigned-integer32",
			Pointer: &ui32,
			Result:  "2000",
		},
		{
			Field:   "test.unsigned-integer64",
			Pointer: &ui64,
			Result:  "20000",
		},
		{
			Field:   "test.float32",
			Pointer: &f32,
			Result:  "1.234",
		},
		{
			Field:   "test.float64",
			Pointer: &f64,
			Result:  "2.3456",
		},
		{
			Field:   "test.string",
			Pointer: &str,
			Result:  "my test",
		},
	}

	wg.Add(len(testCases))

	for _, tc := range testCases {
		t.Run(tc.Field, func(t *testing.T) {
			// bind the field
			psrv.BindField(tc.Field, tc.Pointer)

			// set the field
			go sendToPipe(t, path, fmt.Sprintf("%s=%s", tc.Field, tc.Result))

			// check if field was setted
			go func(p interface{}, res string) {
				defer wg.Done()
				for {
					s := fmt.Sprintf("%v", reflect.ValueOf(p).Elem())
					if strings.EqualFold(res, s) {
						return
					}
					time.Sleep(time.Millisecond)
				}
			}(tc.Pointer, tc.Result)
		})
	}
}

func TestBindField(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	// var (
	// 	arrInt       []int
	// 	arrString    []string
	// 	mapStringInt map[string]int
	// 	myStruct     struct {
	// 		Boolean bool
	// 		Integer int
	// 		Float32 float32
	// 		String  string
	// 	}
	// )

	var wg sync.WaitGroup

	testBindField(t, psrv, path, &wg)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(time.Second):
		t.Error("BindField didn't update the field")
	case <-done:
	}
}
