package dep

type VersionType int

const (
	Revision VersionType = iota
	Version
	Branch
)

type VersionInfo struct {
	Version string
	Type    VersionType
}
