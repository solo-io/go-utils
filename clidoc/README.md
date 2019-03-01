# Generate standard Solo.io documentation from Cobra CLIs

- Takes the entry point to a Cobra CLI (the root `*cobra.CMD`) and produces files formatted for Solo's online documentation

## Example usage
- from Squash:
```go
package main

import (
	"log"

	"github.com/solo-io/go-utils/clidoc"

	"github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	app, err := squashctl.App(version.Version)
	if err != nil {
		log.Fatal(err)
	}
	clidoc.MustGenerateCliDocs(app)
}
```
