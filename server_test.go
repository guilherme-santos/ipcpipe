package ipcpipe_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/guilherme-santos/ipcpipe"

	"github.com/stretchr/testify/assert"
)

func TestNewServer_DoNotReturnErrorInCaseNamePipeAlreadyExists(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	err := syscall.Mkfifo(path, 0600)
	assert.NoError(t, err)

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)
	psrv.Close()
}

func TestNewServer_ReturnErrorIfFileExistsAndItIsNotANamedPipe(t *testing.T) {
	path, cleanup := genPath(t, "regular-file")
	defer cleanup()

	_, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)

	_, err = ipcpipe.NewServer(path)
	assert.Equal(t, ipcpipe.ErrFileIsNotNamedPipe, err)
}

func TestClose_RemoveNamedPipe(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	err = psrv.Close()
	assert.NoError(t, err)

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestCommand(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	fnCalled := make(chan struct{})
	psrv.Command("test", func(cmd string, args ...string) error {
		assert.Equal(t, "test", cmd)
		if assert.Len(t, args, 2) {
			assert.Equal(t, "arg", args[0])
			assert.Equal(t, "with space", args[1])
		}
		close(fnCalled)
		return nil
	})

	go sendToPipe(t, path, `test arg "with space"`)

	select {
	case <-time.After(time.Second):
		t.Error("Command test was expected to be called but it wasn't")
	case <-fnCalled:
	}
}

func TestBind(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	fnCalled := make(chan struct{})
	psrv.Bind("test", func(field, val string) error {
		assert.Equal(t, "test", field)
		assert.Equal(t, "true", val)
		close(fnCalled)
		return nil
	})

	go sendToPipe(t, path, `test=true`)

	select {
	case <-time.After(time.Second):
		t.Error("Bind function should be called when set test")
	case <-fnCalled:
	}
}

func TestBind_WithDot(t *testing.T) {
	path, cleanup := genPath(t, "namedpipe")
	defer cleanup()

	psrv, err := ipcpipe.NewServer(path)
	assert.NoError(t, err)

	fnCalled := make(chan struct{})
	psrv.Bind("app.test", func(field, val string) error {
		assert.Equal(t, "app.test", field)
		assert.Equal(t, "true", val)
		close(fnCalled)
		return nil
	})

	go sendToPipe(t, path, `app.test=true`)

	select {
	case <-time.After(time.Second):
		t.Error("Bind function should be called when set app.test")
	case <-fnCalled:
	}
}
