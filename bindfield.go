package ipcpipe

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func bindField(field string, v interface{}) FieldFunc {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		err := &InvalidBindFieldError{reflect.TypeOf(v)}
		panic(err.Error())
	}

	elem := rv.Elem()
	return func(field, value string) error {
		return bindToElement(elem, value)
	}
}

func bindToElement(elem reflect.Value, value string) error {
	typ := elem.Type()

	switch elem.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		elem.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		if reflect.Zero(typ).OverflowInt(n) {
			return &BindTypeError{
				Value: "number " + value,
				Type:  typ,
			}
		}
		elem.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}

		if reflect.Zero(typ).OverflowUint(n) {
			return &BindTypeError{
				Value: "number " + value,
				Type:  typ,
			}
		}

		elem.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}

		if reflect.Zero(typ).OverflowFloat(n) {
			return &BindTypeError{
				Value: "number " + value,
				Type:  typ,
			}
		}

		elem.SetFloat(n)

	case reflect.String:
		elem.SetString(value)

	case reflect.Slice:
		// Create an slice with the type of the array
		newArr := reflect.MakeSlice(elem.Type(), 1, 1)
		fmt.Println(newArr.Type())

		iArr := newArr.Interface()
		fmt.Printf("iArr, %T\n", iArr)
		err := json.Unmarshal([]byte(value), &iArr)
		if err != nil {
			return err
		}

		fmt.Printf("iArr, %T\n", iArr)
		arr := iArr.([]interface{})
		fmt.Printf("iArr, %T\n", iArr)
		fmt.Printf("arr, %T\n", arr)
		for k := range arr {
			fmt.Println(reflect.ValueOf(arr[k]).Type())
			reflect.Append(newArr, reflect.ValueOf(arr[k]))
		}
		elem.Set(newArr)

	default:
		return &BindTypeError{
			Value: value,
			Type:  typ,
		}
	}

	return nil
}

// InvalidBindFieldError describes an invalid argument passed to Bind.
// (The argument to Bind must be a non-nil pointer.)
type InvalidBindFieldError struct {
	Type reflect.Type
}

func (e *InvalidBindFieldError) Error() string {
	if e.Type == nil {
		return "ipcpipe: BindField(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "ipcpipe: BindField(non-pointer " + e.Type.String() + ")"
	}
	return "ipcpipe: BindField(nil " + e.Type.String() + ")"
}

// BindTypeError describes a value that was not appropriate
// for a value of a specific Go type.
type BindTypeError struct {
	Value  string       // description of value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Struct string       // name of the struct type containing the field
	Field  string       // name of the field holding the Go value
}

func (e *BindTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "ipcpipe: cannot bind " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "ipcpipe: cannot bind " + e.Value + " into Go value of type " + e.Type.String()
}
