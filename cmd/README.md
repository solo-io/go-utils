

Any binaries that serve as general cross-repo utilities could be stored in this directory.

Expectations:
- Util is useful across multiple projects
- detailed `make` targets for scripts live in their own directories
  - only a single make target per script should be called from this directory
- Each script declares its own go.mod file
