package ipcpipe

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"text/scanner"

	"github.com/pkg/errors"
)

// ErrFileIsNotNamedPipe is returned when you pass a path for NewServer and the file
// is not a named pipe.
var ErrFileIsNotNamedPipe = errors.New("ipcpipe: file is not a named pipe")

type (
	CommandFunc func(args ...string) error
	FieldFunc   func(value string) error
)

type Server struct {
	pipe     *os.File
	stop     chan struct{}
	commands map[string]CommandFunc
	fields   map[string]FieldFunc
}

// createNamedPipe creates a named pipe.
func createNamedPipe(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := syscall.Mkfifo(path, 0600)
		if err != nil {
			return errors.Wrap(err, "ipcpipe: creating named pipe")
		}
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "ipcpipe")
	}

	if (info.Mode() & os.ModeNamedPipe) == 0 {
		return ErrFileIsNotNamedPipe
	}
	return nil
}

// NewServer returns a new server that will create a named pipe (if necessary)
// and returns a server instance to be able to register your commands and binds.
func NewServer(path string) (*Server, error) {
	err := createNamedPipe(path)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrap(err, "ipcpipe")
	}

	srv := &Server{
		pipe:     f,
		stop:     make(chan struct{}),
		commands: map[string]CommandFunc{},
		fields:   map[string]FieldFunc{},
	}
	go srv.read()

	return srv, nil
}

func (srv *Server) execute(fn CommandFunc, s scanner.Scanner) error {
	args := make([]string, 0)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		args = append(args, strings.Trim(s.TokenText(), `"`))
	}

	return fn(args...)
}

func (srv *Server) read() {
	for {
		// Check if the client is trying to stop the server
		select {
		case <-srv.stop:
			return
		default:
		}

		var s scanner.Scanner
		s.Init(srv.pipe)
		s.Error = func(s *scanner.Scanner, msg string) {
			// TODO handle error
		}

		tok := s.Scan()
		if tok == scanner.EOF {
			continue
		}

		cmd := s.TokenText()
		if fn, ok := srv.commands[cmd]; ok {
			err := srv.execute(fn, s)
			if err != nil {
				// TODO handle error
				panic(fmt.Sprint(cmd, ": ", err))
			}
			continue
		}

		field := new(strings.Builder)
		field.WriteString(s.TokenText())

		val := new(strings.Builder)
		tmp := field

		for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			fmt.Printf("%q / %q\n", s.Peek(), s.TokenText())
			if s.TokenText() == "=" {
				// When find = means that value is starting
				tmp = val
				continue
			}
			if s.Peek() == ' ' {
				tmp.WriteString(" ")
			}
			tmp.WriteString(s.TokenText())
		}

		if fn, ok := srv.fields[field.String()]; ok {
			err := fn(val.String())
			if err != nil {
				// TODO handle error
				panic(fmt.Sprint(field.String(), ": ", err))
			}
			continue
		}
	}
}

// Command adds a function to when receive the cmd be executed.
func (srv *Server) Command(cmd string, fn CommandFunc) {
	if _, ok := srv.commands[cmd]; ok {
		panic("command already registered")
	}

	srv.commands[cmd] = fn
}

// BindField updates v when field is set.
func (srv *Server) BindField(field string, v interface{}) {
	srv.Bind(field, bindField(field, v))
}

// Bind binds the field to the variable informed.
func (srv *Server) Bind(field string, fn FieldFunc) {
	if _, ok := srv.fields[field]; ok {
		panic("field already registered")
	}

	srv.fields[field] = fn
}

// Close closes the file descriptor and also remove the named pipe.
func (srv *Server) Close() error {
	srv.stop <- struct{}{}

	err := srv.pipe.Close()
	if err != nil {
		return errors.Wrap(err, "ipcpipe: closing file")
	}

	err = os.Remove(srv.pipe.Name())
	if err != nil {
		return errors.Wrap(err, "ipcpipe: removing file")
	}

	return nil
}
