# IPC Pipe

**I**nter **P**rocess **C**ommunication **Pipe** is a library that creates a named pipe in the location specified by the config and provides the following features:

- Execute commands sent to the named pipe
- Update a variable on the fly
- Trigger an action when a variable is updated

# Example

### Server

Here we have a server example that shows how you can use the library and some usecases.

```
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Config struct {
	Logger struct {
		Prefix string
	}
	SleepFor time.Duration
}

func main() {
	var cfg Config
	cfg.SleepFor = 1 * time.Second

	logger := log.New(os.Stdout, cfg.Logger.Prefix, 0)

	psrv := ipcpipe.NewServer("maincfg")
    defer psrv.Close()

    // Add a command called log which will log the parameters received
	psrv.Command("log", func(args ...string) {
		if len(args) == 0 {
			return
		}
		if len(args) == 1 {
			logger.Println(args[0])
			return
		}
		logger.Println(args[0], args[1]...)
	})

    // Update directly a field
	psrv.BindField("app.sleep_for", &cfg.SleepFor)
	psrv.BindField("app.logger.prefix", &cfg.Logger.Prefix)

    // Execute the following function when update app.logger.output field
	psrv.Bind("app.logger.output", func(output string) error {
		var w io.Writer

		switch output {
		case "output":
			w = os.Stdout
		case "discard", "":
			w = ioutil.Discard
		default:
			f, err := os.Open(output)
			if err != nil {
				return err
			}
			w = f
		}

        // update logger with the new output
		logger = log.New(w, cfg.Logger.Prefix, 0)
	})

    // loop 
	for {
		logger.Println("Something happened")
		time.Sleep(cfg.SleepFor)
	}
}
```

You can interact with this server typing the following commands:

* Set `sleep_for` to 1.5 seconds

```
echo app.sleep_for=1.5s > maincfg
```

* Set `logger.prefix` to `new-prefix: `

```
echo app.logger.prefix="new-prefix: " > maincfg
```

* Update `logger.output` to send the logs to `./my.log`

```
echo app.logger.output=./my.log > maincfg
```

* Execute `log` command

```
echo log "hello world" > maincfg
echo log "hello %s %d" world 2 > maincfg
```
