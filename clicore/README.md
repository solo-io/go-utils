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

## Usage in tests
- `clicore` was designed to simplify CLI specification and testing.
- To run `clicore` in test mode, call `cliOutput, err := cli.GlooshotConfig.RunForTest(args)`.
  - Output from the command (stdout, stderr, and any log files) can then be validated one by one.