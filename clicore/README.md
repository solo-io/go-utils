# CLI Core
- A library for feature-rich, configurable, and cross-project consistent command line interfaces.

## Usage
- To use clicore, just define a `CommandConfig` and call `config.Run()`.
### Example
- Sample `main` package, importing your cli:
```go
package main
import (
	"myrepo/pkg/mycli"
)
func main() {
	mycli.MyCommandConfig.Run()
}
```

- Sample package, defining your CLI library and exporting the `CommandConfig`:

```go
package mycli

import "github.com/solo-io/go-utils/clicore"

var MyCommandConfig = clicore.CommandConfig{
	Command:             App,
	Version:             version.Version,
	FileLogPathElements: FileLogPathElements,
	OutputModeEnvVar:    OutputModeEnvVar,
	RootErrorMessage:    ErrorMessagePreamble,
	LoggingContext:      []interface{}{"version", version.Version},
}
```

### How to write logs to the console and log file

There are three helpers that you can use:

```go

contextutils.CliLogInfow(ctx, "this info log goes to file and console")
contextutils.CliLogWarnw(ctx, "this warn log goes to file and console")
contextutils.CliLogErrorw(ctx, "this error log goes to file and console")
```

Key-value pairs are supported too:

```go
contextutils.CliLogInfow(ctx, "this infow log should go to file and console",
    "extrakey1", "val1")
```

Which is equivalent to the longer form:

```go
contextutils.LoggerFrom(ctx).Infow("message going to file only",
	zap.String("cli", "info that will go to the console and file",
	"extrakey1", "val1")
```

## Usage in tests
- `clicore` was designed to simplify CLI specification and testing.
- To run `clicore` in test mode, call `cliOutput, err := cli.GlooshotConfig.RunForTest(args)`.
  - Output from the command (stdout, stderr, and any log files) can then be validated one by one.
- **See the [test file](cli_test.go) for an example**

