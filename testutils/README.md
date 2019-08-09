
# Fail Handlers

## PrintTrimmedStack

The `PrintTrimmedStack` fail handler simplifies error tracking in ginkgo tests by printing a condensed stack trace upon failure. Printout excludes well-known overhead files so you can more easily sight the failing line. This eliminates the need to count stack offset via `ExpectWithOffset`. You can just use `Expect`.

### Usage

To use, register `PrintTrimmedStack` as a prefail handler with `RegisterPreFailHandler` in your `ginkgo` suite:

```go
func TestCliCore(t *testing.T) {
	testutils.RegisterPreFailHandler(testutils.PrintTrimmedStack)
	testutils.RegisterCommonFailHandlers()
	RegisterFailHandler(Fail)
	testutils.SetupLog()
	RunSpecs(t, "Clicore Suite")
}
```