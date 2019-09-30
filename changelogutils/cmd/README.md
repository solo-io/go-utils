This is a sample of the output from the changelog generation script.

To update this file, run: `cd changelogutils/cmd && make README -B && cd -`

---


# v0.10.16

**New Features**

- Utility for generating changelog summary documentation from a changelog directory. (https://github.com/solo-io/go-utils/issues/309)


# v0.10.15

This release contained no user-facing changes.


# v0.10.14

This release contained no user-facing changes.


# v0.10.13

**Fixes**

- Use correct repo when getting sha for a tag in repo client. (https://github.com/solo-io/go-utils/issues/304)


# v0.10.12

**Fixes**

- Fix log messages to avoid nil value issues. (https://github.com/solo-io/go-utils/issues/302)


# v0.10.11

**New Features**

- Add utilites to help move and inspect files in a repo. (https://github.com/solo-io/go-utils/issues/298)

**Fixes**

- Invalid data was being logged, causing the bot handler to crash. (https://github.com/solo-io/go-utils/issues/301)


# v0.10.10

**New Features**

- Add builder tag and other utilities to simplify writing build bots. (https://github.com/solo-io/go-utils/issues/297)


# v0.10.9

**New Features**

- Return more info about a dependency version retrieved from a .toml file. (https://github.com/solo-io/go-utils/issues/294)


# v0.10.8

This release contained no user-facing changes.


# v0.10.7

**New Features**

- Consider all possible version fields (version, branch, revision) when retrieving a dependency version from a .toml file. (https://github.com/solo-io/go-utils/issues/291)


# v0.10.6

**New Features**

- Support PR number tag on builds. (https://github.com/solo-io/go-utils/issues/289)


# v0.10.5

**New Features**

- add util to generate certs (https://github.com/solo-io/go-utils/pull/287)


# v0.10.4

This release contained no user-facing changes.


# v0.10.3

This release contained no user-facing changes.


# v0.10.2

This release contained no user-facing changes.


# v0.10.1

**New Features**

- Add manifest test function for getting permissions granted to k8s service accounts. (https://github.com/solo-io/go-utils/issues/277)


# v0.10.0

**Breaking Changes**

- Add option to respect manifest namespace on installer. Breaks the ReconcileResources API of the KubeInstaller (https://github.com/solo-io/go-utils/pull/266)

**New Features**

- Add grpc health check lib. (https://github.com/solo-io/go-utils/issues/275)


# v0.9.21

**New Features**

- Add contextual info to manifestutils error messages. (https://github.com/solo-io/go-utils/issues/272)


# v0.9.20

This release contained no user-facing changes.


# v0.9.19

**New Features**

- Extend cluster locker to support Consul clusters (https://github.com/solo-io/go-utils/issues/261)
- make kubeinstaller more robust for creating resources (https://github.com/solo-io/go-utils/pull/263)


# v0.9.18

**Fixes**

- Wait for default service account before deploying testrunner. (https://github.com/solo-io/go-utils/issues/259)


# v0.9.17

This release contained no user-facing changes.


# v0.9.16

This release contained no user-facing changes.


# v0.9.15

**Fixes**

- Wait for jobs to complete in the kube installer. (https://github.com/solo-io/go-utils/issues/254)


# v0.9.14

**New Features**

- Add utilities for builder bots. (https://github.com/solo-io/go-utils/issues/252)


# v0.9.13

This release contained no user-facing changes.


# v0.9.12

**New Features**

- Allow for passing of optional description to ExpectEqualProtoMessages. (https://github.com/solo-io/go-utils/issues/249)


# v0.9.11

This release contained no user-facing changes.


# v0.9.10

This release contained no user-facing changes.


# v0.9.9

**Fixes**

- Fix set ref to use the correct logic. (https://github.com/solo-io/go-utils/issues/241)


# v0.9.8

**Fixes**

- Register Issues in the list of webhook events that are handled. (https://github.com/solo-io/go-utils/issues/239)


# v0.9.7

**New Features**

- Adds utilities for writing gcloud subscribers. (https://github.com/solo-io/go-utils/issues/205)
- Support issues webhook events for git bots. (https://github.com/solo-io/go-utils/issues/237)


# v0.9.6

**New Features**

- Helper for cleaning up CRDs after tests. (https://github.com/solo-io/go-utils/issues/234)


# v0.9.5

**Fixes**

- Remove quotes from strings for compatibility with Kubernetes&#39; naming convention. (https://github.com/solo-io/go-utils/issues/232)


# v0.9.4

**New Features**

- Add and implement ApiVersion interface for reasoning about kubernetes api versions. (https://github.com/solo-io/go-utils/issues/230)


# v0.9.3

This release contained no user-facing changes.


# v0.9.2

**New Features**

- tco-echo pod for testing tcp connections. (https://github.com/solo-io/go-utils/issues/227)


# v0.9.1

**New Features**

- Provide a utility function (for tests) to get the absolute path of the directory containing the caller&#39;s file. (https://github.com/solo-io/go-utils/issues/223)


# v0.9.0

**Breaking Changes**

- Removes ApplicationConfig (containing slack notifications and installation id) from the bot config struct. This should only affect consumers of the new bot utilities (changelog-bot, approval-bot). (https://github.com/solo-io/go-utils/issues/223)

**New Features**

- Support customizing bot config with WEBHOOK_SECRET, INTEGRATION_ID, and PRIVATE_KEY_FILE environment variables, to better support use cases around deploying to kubernetes from a git repo without checking in unencrypted secrets. (https://github.com/solo-io/go-utils/issues/223)


# v0.8.24

**New Features**

- Adds some basic utilities for writing bots. (https://github.com/solo-io/go-utils/issues/215)
- Provides a better abstraction for reading changelogs. (https://github.com/solo-io/go-utils/issues/220)
- Added changelog validator interface, deprecating the old changelog validation utilities. (https://github.com/solo-io/go-utils/issues/221)
- Add utilities for working with lazily mounted github repos. (https://github.com/solo-io/go-utils/issues/213)


# v0.8.23

This release contained no user-facing changes.


# v0.8.22

**New Features**

- Add resource/log gathering capability. (https://github.com/solo-io/go-utils/issues/202)
- Add github client getter that just logs a warning when no token is found. (https://github.com/solo-io/go-utils/issues/168)


# v0.8.21

**New Features**

- Port useful env utils for writing grpc api servers. (https://github.com/solo-io/go-utils/issues/206)


# v0.8.20

**Fixes**

- fix manifest sorting (https://github.com/solo-io/go-utils/issues/198)


# v0.8.19

**New Features**

- Add util function to convert resources to manifest (https://github.com/solo-io/go-utils/196)


# v0.8.18

**Fixes**

- Updated commit string format (https://github.com/solo-io/gloo/issues/719)


# v0.8.17

**New Features**

- Add a simple go function that functions like curl. (https://github.com/solo-io/glooshot/issues/16)


# v0.8.16

This release contained no user-facing changes.


# v0.8.15

**New Features**

- Added utils for interacting with config maps and structs, pulled upstream from the service mesh hub. (https://github.com/solo-io/go-utils/issues/192)


# v0.8.14

**Fixes**

- At the start of reconcile, reset Kube client so it&#39;s discovery cache is updated to reflect any newly-created CRDs (i.e. from a background Job) since it was last refreshed. (https://github.com/solo-io/sm-marketplace/issues/452)


# v0.8.13

**New Features**

- Add proto unmarshal function that allows unknown fields rather than fail on them. (https://github.com/solo-io/go-utils/issues/188)


# v0.8.12

**New Features**

- Add termination error function for kube containers. (https://github.com/solo-io/go-utils/issues/185)

**Fixes**

- Shorten error message around changelog files. (https://github.com/solo-io/solobot/issues/100)


# v0.8.11

**New Features**

- Add xerrors support to go-utils for error comparison. (https://github.com/solo-io/go-utils/issues/184)


# v0.8.10

**New Features**

- Added new functions around Toml management to support checking overrides or constraints. (https://github.com/solo-io/go-utils/issues/180)
- You can now set the log level dynamically on loggers constructed via contextutils. (https://github.com/solo-io/go-utils/issues/164)


# v0.8.9

**New Features**

- Add support for testing crds in manifest tester. (https://github.com/solo-io/go-utils/issues/178)


# v0.8.8

**New Features**

- Add support for overrides in toml parser. (https://github.com/solo-io/go-utils/issues/175)
- Add support for secrets to manifesttestutils. (https://github.com/solo-io/go-utils/issues/174)

**Fixes**

- Reduce logging in get_cluster_resources to DEBUG level. (https://github.com/solo-io/go-utils/issues/165)


# v0.8.7

**New Features**

- Add ExpectConfigMapWithYamlData to handle assertions on configmaps with yaml Data values. (https://github.com/solo-io/go-utils/issues/171)


# v0.8.6

**New Features**

- Add retrieve archive utility. (https://github.com/solo-io/go-utils/issues/169)


# v0.8.5

**Fixes**

- fix error in tarutils Tar function. (https://github.com/solo-io/go-utils/issues/167)


# v0.8.4

**Fixes**

- Add interface to testrunner to avoid accessing underlying functionality. (https://github.com/solo-io/go-utils/issues/161)


# v0.8.3

**Fixes**

- Fix testrunner pod regression from last release. (https://github.com/solo-io/go-utils/issues/161)


# v0.8.2

**New Features**

- Added utility for mounting generic compressed tar via URL. (https://github.com/solo-io/go-utils/issues/159)


# v0.8.1

**New Features**

- Add test pod which can output all request metadata. (https://github.com/solo-io/go-utils/issues/157)


# v0.8.0

**Breaking Changes**

- Finishes moving utils into go-utils. Renames logger to log to keep parity with solo-kit. (https://github.com/solo-io/go-utils/issues/155)

**New Features**

- Simplify proto equivalence checks and provide enhanced error message on failure. (https://github.com/solo-io/go-utils/issues/148)


# v0.7.21

**Fixes**

- Install cache now tracks resources even if they fail to become ready for created and deleted resources. (https://github.com/solo-io/go-utils/issues/152)


# v0.7.20

**New Features**

- Allow customizing pull policy when testing containers. (https://github.com/solo-io/go-utils/issues/151)

**Fixes**

- Install cache now tracks resources even if they fail to become ready. (https://github.com/solo-io/go-utils/issues/152)


# v0.7.19

This release contained no user-facing changes.


# v0.7.18

**New Features**

- Distribute a library for testing installation manifests generated from Helm charts. (https://github.com/solo-io/go-utils/issues/146)


# v0.7.17

**Fixes**

- Changelog validation now allows v0.0.1 releases. (https://github.com/solo-io/go-utils/issues/142)


# v0.7.16

**Fixes**

- Changelog validation now allows v0.0.1 releases. (https://github.com/solo-io/go-utils/issues/142)


# v0.7.15

This release contained no user-facing changes.


# v0.7.14

**New Features**

- Introduce a utility for managing zap logs from a CLI tool. The new `clicore` library provides a means of sending human-friendly log messages to the console while capturing full json-formatted logs to a file. Library includes a pair of `Run` methods for siplified execution of the &#34;main&#34; file and simplified output validation during integration tests. (https://github.com/solo-io/go-utils/issues/135)
- Simplified way to call cobra commands from test environments. (https://github.com/solo-io/go-utils/issues/124)
- The `PrintTrimmedStack` fail handler simplifies error tracking in ginkgo tests by printing a condensed stack trace upon failure. Printout excludes well-known overhead files so you can more easily sight the failing line. This eliminates the need to count stack offset via `ExpectWithOffset`. You can just use `Expect`. (https://github.com/solo-io/go-utils/issues/131)


# v0.7.13

**Fixes**

- Only cache flattened, installed resources. (https://github.com/solo-io/go-utils/issues/138)


# v0.7.12

This release contained no user-facing changes.


# v0.7.11

**Fixes**

- Added parameter to help more finely narrow down to just the sha256 files in the build path that pkgmgmtutils.UpdateFormulas really need. (https://github.com/solo-io/go-utils/issues/132)


# v0.7.10

**New Features**

- Added some utility functions to cliutils. (https://github.com/solo-io/go-utils/issues/130)


# v0.7.9

**Fixes**

- Added a FormulaOptions.FormulaName field to allow the FormulaOptions.Name field to be a descriptive name, and FormulaOptions.FormulaName to be used to reference the proper formula name as needed in branch and commit name conventions for external PRs. (https://github.com/solo-io/go-utils/issues/128)


# v0.7.8

**New Features**

- The utilities in `pkgmgmtutils/` now support updating homebrew and gofish style formulas as part of a release process. (https://github.com/solo-io/gloo/issues/520)


# v0.7.7

**New Features**

- The utilities in `installutils/helmchart` now support loading a helm chart and rendering manifests from a directory in a github repository. There is also a utility for batch loading charts from a github directory, for example to batch load several hundred charts from `https://github.com/helm/charts/tree/master/stable`. (https://github.com/solo-io/go-utils/issues/121)
- There is now a new utility package called `fsutils` to support common file-system operations. (https://github.com/solo-io/go-utils/issues/122)


# v0.7.6

**Fixes**

- List resources should access the cached resources via the cache API, which properly locks. (https://github.com/solo-io/go-utils/issues/119)


# v0.7.5

**New Features**

- Expose configurable retry options on kube installer (https://github.com/solo-io/go-utils/issues/117)


# v0.7.4

**Fixes**

- Ensure that subresources are deleted by the kube installer on delete (https://github.com/solo-io/go-utils/issues/113)


# v0.7.3

**New Features**

- Allow listing from the cache all the unique values for a particular label key. This enables discovering label key/value pairs that should be purged because the CRD that led to the creation of that resource was deleted. (https://github.com/solo-io/go-utils/issues/110)


# v0.7.2

**New Features**

- Added a `SanitizeName` function to `kubeutils`. (https://github.com/solo-io/go-utils/issues/108)

**Fixes**

- Remove reference to docs in shared `versionutils` function to get `TAGGED_VERSION` from the environment. (https://github.com/solo-io/go-utils/issues/99)


# v0.7.1

**New Features**

- Enable adding arbitrary logging meta in context logger (https://github.com/solo-io/go-utils/issues/105)


# v0.7.0

**Breaking Changes**

- Moves testutils from gloo into go-utils. (https://github.com/solo-io/go-utils/issues/95)

**New Features**

- Finished porting over installer from Supergloo. (https://github.com/solo-io/go-utils/issues/103)
- Start bringing in installutils from supergloo. (https://github.com/solo-io/go-utils/issues/100)

**Fixes**

- The test Gloo wrapper now calls `glooctl uninstall` with the `--delete-namespace` flag. (https://github.com/solo-io/go-utils/issues/90)


# v0.6.3

**New Features**

- Add options to curl wrapper for toggling -v and -s flags for verbosity specification. (https://github.com/solo-io/go-utils/issues/91)


# v0.6.2

**New Features**

- Changelog entries now support an optional `resolvesIssue` boolean field, defaulting to `true`, to specify if the issue referenced in `issueLink` should be resolved when the PR merges. (https://github.com/solo-io/go-utils/issues/88)

**Fixes**

- Fixes a bug where locks will fail if the lock present in the cluster is deleted while trying to acquire. (https://github.com/solo-io/go-utils/issues/83)


# v0.6.1

**Fixes**

- Fixes a regression introduced in v0.6.0 where uploading github release assets would always fail. (https://github.com/solo-io/go-utils/issues/81)


# v0.6.0

**Breaking Changes**

- create heartbeat for cluster lock to allow for quicker recovery (https://github.com/solo-io/go-utils/issues/77)

**Fixes**

- Retry uploading assets to github if there&#39;s an error. (https://github.com/solo-io/go-utils/issues/76)


# v0.5.1

**New Features**

- Add a tiemout to the cluster lock. (https://github.com/solo-io/go-utils/issues/72)


# v0.5.0

**Breaking Changes**

- Remove remnants of changelog refactor that wasn&#39;t done on master. (https://github.com/solo-io/go-utils/issues/55)

**New Features**

- Create a basic lock for test-clusters. (https://github.com/solo-io/supergloo/issues/236)

**Fixes**

- Use more specific project name in docs PRs, if repo is configured with changelogs. (https://github.com/solo-io/go-utils/issues/39)


# v0.4.1

This release contained no user-facing changes.


# v0.4.0

**Breaking Changes**

- Move install utils into this repo and start cleanup. (https://github.com/solo-io/go-utils/issues/61)


# v0.3.1

**New Features**

- Add util method to wait for crds to become active. (https://github.com/solo-io/go-utils/issues/59)


# v0.3.0

**Breaking Changes**

- Upgrade to go modules. (https://github.com/solo-io/go-utils/issues/56)


# v0.2.16

**New Features**

- Begin work on dependency changelogs. (https://github.com/solo-io/go-utils/issues/48)


# v0.2.15

**Fixes**

- Fixed bug in cli docs push logic. (https://github.com/solo-io/go-utils/issues/50)


# v0.2.14

This release contained no user-facing changes.


# v0.2.13

**Fixes**

- Fixes the behavior where changelog verification would fail if the latest release was a prerelease. (https://github.com/solo-io/go-utils/issues/45)


# v0.2.12

**Fixes**

- No longer create unwanted nested directory when pushing solo-kit docs for the first time. (https://github.com/solo-io/go-utils/issues/43)


# v0.2.11

**Fixes**

- Fixes the sha upload in `UploadReleaseAssetsCli` to upload a checksum for `foo` that matches the output of `shasum -a 256 foo &gt; foo.sha256`. (https://github.com/solo-io/go-utils/issues/41)


# v0.2.10

**New Features**

- A utility CLI has been added for uploading release artifacts to github, to replace the old shell script. See the [readme](https://github.com/solo-io/go-utils/tree/master/githubutils) for more information. (https://github.com/solo-io/go-utils/issues/38)

**Fixes**

- PushDocsCli no longer errors on the initial push when the destination directory doesn&#39;t exist. (https://github.com/solo-io/go-utils/issues/40)


# v0.2.9

**New Features**

- The docs push utility now supports automated CLI docs. (https://github.com/solo-io/go-utils/issues/33)
- The docs can support API or CLI docs that are not in the root of the repo. (https://github.com/solo-io/go-utils/issues/33)
- The docs CLI library now includes the full CLI, so projects can execute docs push in 1 line. (https://github.com/solo-io/go-utils/issues/33)
- Moves common documentation generation to a shared lib (https://github.com/solo-io/go-utils/issues/35)


# v0.2.8

**New Features**

- Changelog now enabled for this repo. (https://github.com/solo-io/go-utils/issues/31)

**Fixes**

- Markdown generation now always ends in two new lines. (https://github.com/solo-io/go-utils/issues/30)

