package ipcpipe_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/guilherme-santos/ipcpipe"

	"github.com/stretchr/testify/assert"
)

func testBindField_SimpleTypes(t *testing.T, psrv *ipcpipe.Server, path string) {
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

	for _, tc := range testCases {
		t.Run(tc.Field, func(t *testing.T) {
			// bind the field
			psrv.BindField(tc.Field, tc.Pointer)

			// set the field
			go sendToPipe(t, path, fmt.Sprintf("%s=%s", tc.Field, tc.Result))

			done := make(chan struct{})
			// check if field was setted
			go func(p interface{}, res string) {
				for {
					s := fmt.Sprintf("%v", reflect.ValueOf(p).Elem())
					if strings.EqualFold(res, s) {
						close(done)
						return
					}
					time.Sleep(time.Millisecond)
				}
			}(tc.Pointer, tc.Result)

			select {
			case <-time.After(200 * time.Millisecond):
				t.Errorf("BindField didn't update the field, it was expected: %v", tc.Result)
			case <-done:
			}
		})
	}
}

func testBindField_SliceInteger(t *testing.T, psrv *ipcpipe.Server, path string) {
	var arrInt []int
	expectedArr := []int{1, 2, 3, 4}

	// bind the field
	psrv.BindField("test.slice.integer", &arrInt)

	// set the field
	j, _ := json.Marshal(expectedArr)
	go sendToPipe(t, path, fmt.Sprintf("test.slice.integer=%s", j))

	done := make(chan struct{})
	// check if field was setted
	go func() {
		for {
			if len(arrInt) == 4 {
				sameValues := true
				for i := range arrInt {
					if expectedArr[i] != arrInt[i] {
						t.Errorf("Expected %v but got %v", expectedArr, arrInt)
					}
				}

				if sameValues {
					close(done)
					return
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Errorf("BindField didn't update the field")
	case <-done:
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

	t.Run("SimpleTypes", func(t *testing.T) {
		testBindField_SimpleTypes(t, psrv, path)
	})
	t.Run("SliceInteger", func(t *testing.T) {
		testBindField_SliceInteger(t, psrv, path)
	})
}

func TestBindField_IgnoringLeadingWhitespaces(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	var str string
	v := "text with some  spaces  "

	psrv.BindField("string.ignoring.whitespace", &str)

	go sendToPipe(t, path, fmt.Sprintf("  %s  =  %s", "string.ignoring.whitespace", v))

	done := make(chan struct{})
	// check if field was setted
	go func() {
		for {
			if strings.EqualFold(v, str) {
				close(done)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Errorf("BindField didn't update the field, it was expected: %q", v)
	case <-done:
	}
}

func TestBindField_DoNotIgnoreLeadingWhitespaces(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	var str string
	v := "  space in the end  "

	psrv.BindField("string.do-not-ignore.whitespace", &str)

	go sendToPipe(t, path, fmt.Sprintf("  %s  = \"%s\" ", "string.do-not-ignore.whitespace", v))

	done := make(chan struct{})
	// check if field was setted
	go func() {
		for {
			if strings.EqualFold(v, str) {
				close(done)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Errorf("BindField didn't update the field, it was expected: %q", v)
	case <-done:
	}
}
