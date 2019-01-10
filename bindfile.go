package ipcpipe

import (
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
	typ := elem.Type()

	return func(value string) error {
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
				return &UnmarshalTypeError{
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
				return &UnmarshalTypeError{
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
				return &UnmarshalTypeError{
					Value: "number " + value,
					Type:  typ,
				}
			}

			elem.SetFloat(n)

		case reflect.String:
			fmt.Println(value)
			elem.SetString(value)

		default:
			return &UnmarshalTypeError{
				Value: value,
				Type:  typ,
			}
		}

		return nil
	}
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

// UnmarshalTypeError describes a value that was not appropriate
// for a value of a specific Go type.
type UnmarshalTypeError struct {
	Value  string       // description of value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Struct string       // name of the struct type containing the field
	Field  string       // name of the field holding the Go value
}

func (e *UnmarshalTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "ipcpipe: cannot bind " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "ipcpipe: cannot bind " + e.Value + " into Go value of type " + e.Type.String()
}
