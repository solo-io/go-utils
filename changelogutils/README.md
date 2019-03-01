# Changelog

Any repository set up with the Solo bot webhook for CI and release builds can opt into 
using the automated changelog provided in this utility. The benefits of using the changelog include:

- Automatically producing the description for the github release page, and optionally the docs. 
- Ensuring that every change is described for in the release notes, with links to Github issues.  
- Ensuring that the release versions are incremented correctly according to semantic versioning. 

## Turning on Changelogs

Add a top-level directory called "changelog" in the repo. This directory will ultimately contain a structure like
the following: 

```
changelog/
  v0.1.0/
    summary.md
    foo.yaml
    bar.yaml
  v0.2.0/
    foo2.yaml
  v0.2.1/
    foo3.yaml

```

Here, `v0.1.0`, `v0.2.0`, and `v0.2.1` represent released or unreleased versions of the project. The bot will 
check to make sure there is exactly one directory corresponding to an unreleased version 
(greater than the latest release tag) and that at least one new changelog file has been added.

## Changelog files

A changelog file contains a list of changelog entries, like this:

```yaml
changelog:
  - type: NEW_FEATURE
    description: Gloo releases now automatically publish a changelog to the docs.
    issueLink: https://github.com/solo-io/gloo/issues/465
  - ...
  - ...
```
 
Type must be one of `NEW_FEATURE`, `FIX`, `BREAKING_CHANGE`, or `NON_USER_FACING`. 

Changelog entries that are not of type `NON_USER_FACING` must have a description and an issue link. 
Those fields are optional for `NON_USER_FACING` changes. 

The name of the changelog filename does not matter. It is useful to pick a unique name for the PR, 
to avoid potential merge conflicts. For instance, you may add this change in a file called 
`publish_changelogs.yaml`. As long as it is valid yaml in the correct tag directory, it will be 
considered valid. 

### Special files: summary and closing

There are two special files that can be added to assist with changelog rendering. These are:
`summary.md` and `closing.md`. These files can be added or modified at any time, and should
contain valid markdown. When the changelog is rendered, the summary will be included at the 
top, before the list of changes, and the closing notes will be included at the end. 

## Changelog validation

When changelogs are enabled, PRs must include a valid changelog file or they will fail verification 
and cannot be merged. 

The changelog file must be included in the correct directory according to semver. For example, if the 
last released version was `v0.2.5`, new PRs that don't create breaking changes should add changelog 
files into `v0.2.6` (they should create the directory if it does not yet exist). If the PR contains 
a breaking change entry, it should create or rename the existing directory to `v0.3.0`. 

For projects that have already released `v1.0.0`, breaking changes should increment the major version 
instead (`v2.0.0`). Non-breaking changes should increment the minor version (`v1.1.0`).

## Releasing a stable v1.0 version

There is one special case for incrementing versions: publishing a stable 1.0 API. This can be done 
by setting the `releaseStableApi` field to true. 

```yaml
changelog: 
  - ...
  . ...
releaseStableApi: true 
``` 

## Publishing release notes to Github

Changelogs will automatically be rendered into a markdown string, and the CI release bot will 
immediately update a release description to be the changelog when a release is published. 
No manually entered description should be used (it will get overwritten). 

## Pushing release notes and docs to Solo Docs

For some repos, it makes sense to push the change log automatically into a docs PR, along 
with other generated docs. There is a utility in `go-utils/docsutils/cli.go` to help with this. 

This utility assumes the docs repo has been set up in the following way:
* The theme was updated with a readfile shortcode override (see `solo-docs/testrepo/layouts/shortcodes`)
* A changelog directory was added in `solo-docs/<product>/docs/changelog`
* The changelog directory includes an `_index.md` file that sets the front matter and specifies one or more files to read in
* The changelog file for a particular github project is in `solo-docs/<product>/docs/changelog/<project>-changelog`

Here's an example go script that creates a PR on release builds:
```go
package main

import (
	"fmt"
	"github.com/solo-io/go-utils/docsutils"
	"os"
)

func main() {
	tag, present := os.LookupEnv("TAGGED_VERSION")
	if !present || tag == "" {
		fmt.Printf("TAGGED_VERSION not found in environment, skipping docs PR.")
		os.Exit(0)
	}
	err := docsutils.CreateDocsPRSimple("solo-io", "testrepo", tag)
	if err != nil {
		fmt.Printf("Error creating docs PR: %s\n", err.Error())
		os.Exit(1)
	}
}
```

This PR prepends the new changelog at the top of the `testrepo-changelog` file. An example PR 
created from this script is: https://github.com/solo-io/solo-docs/pull/69. 

This utility can also be used to push autogenerated API docs, and soon, CLI docs. These features
will be documented in the near future. 
