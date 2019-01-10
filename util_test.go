package ipcpipe_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func genPath(t *testing.T, filename string) (string, func()) {
	tmpdir, err := ioutil.TempDir("", "ipcpipe")
	assert.NoError(t, err)

	return filepath.Join(tmpdir, filename), func() {
		os.RemoveAll(tmpdir)
	}
}

// sendMsgMu is a lock to guarantee that only message is send a time
// otherwise the server can mix up them.
var sendMsgMu sync.Mutex

func sendToPipe(t *testing.T, path, cmd string) {
	sendMsgMu.Lock()
	defer sendMsgMu.Unlock()

	f, err := os.OpenFile(path, os.O_WRONLY, 0600)
	assert.NoError(t, err)

	_, err = f.WriteString(cmd)
	assert.NoError(t, err)

	f.Sync()
	f.Close()
	time.Sleep(time.Millisecond)
}
