package versionutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

const (
	gopkgToml    = "Gopkg.toml"
	constraint   = "constraint"
	override     = "override"
	nameConst    = "name"
	versionConst = "version"
	imageConst   = "image"

	GlooPkg      = "github.com/solo-io/gloo"
	SoloKitPkg   = "github.com/solo-io/solo-kit"
	SuperglooPkg = "github.com/solo-io/supergloo"
)

var (
	UnableToFindVersionInTomlError = func(pkgName string) error {
		return fmt.Errorf("unable to find version for %s in toml", pkgName)
	}
)

func PinGitVersion(relativeRepoDir string, version string) error {
	tag := GetTag(version)
	cmd := exec.Command("git", "checkout", tag)
	cmd.Dir = relativeRepoDir
	buf := &bytes.Buffer{}
	out := io.MultiWriter(buf, os.Stdout)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, buf.String())
	}
	return nil
}

func GetGitVersion(relativeRepoDir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--dirty")
	cmd.Dir = relativeRepoDir
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "%v failed: %s", cmd.Args, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetTag(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func GetVersionFromTag(shouldBeAVersion string) (string, error) {
	definiteTag := GetTag(shouldBeAVersion)
	version, err := ParseVersion(definiteTag)
	if err != nil {
		return "", err
	}
	versionString := version.String()
	return versionString[1:], nil
}

// Deprecated: Use GetTomlVersion instead
func GetVersion(pkgName string, tomlTree []*toml.Tree) (string, error) {
	for _, v := range tomlTree {
		if v.Get(nameConst) == pkgName && v.Get(versionConst) != "" {
			return v.Get(versionConst).(string), nil
		}
	}
	return "", UnableToFindVersionInTomlError(pkgName)
}

func GetTomlVersion(pkgName string, toml *TomlWrapper) (string, error) {
	for _, v := range toml.Overrides {
		if v.Get(nameConst) == pkgName && v.Get(versionConst) != "" {
			return v.Get(versionConst).(string), nil
		}
	}
	for _, v := range toml.Constraints {
		if v.Get(nameConst) == pkgName && v.Get(versionConst) != "" {
			return v.Get(versionConst).(string), nil
		}
	}
	return "", UnableToFindVersionInTomlError(pkgName)
}

// Deprecated: Use ParseFullToml instead
func ParseToml() ([]*toml.Tree, error) {
	return ParseTomlFromDir("")
}

// Deprecated: Use ParseFullTomlFromDir instead
func ParseTomlFromDir(relativeDir string) ([]*toml.Tree, error) {
	return parseTomlFromDir(relativeDir, constraint)
}

// Deprecated: Use ParseFullToml instead
func ParseTomlOverrides() ([]*toml.Tree, error) {
	return ParseTomlOverridesFromDir("")
}

// Deprecated: Use ParseFullTomlFromDir instead
func ParseTomlOverridesFromDir(relativeDir string) ([]*toml.Tree, error) {
	return parseTomlFromDir(relativeDir, override)
}

func parseTomlFromDir(relativeDir, configType string) ([]*toml.Tree, error) {
	config, err := toml.LoadFile(filepath.Join(relativeDir, gopkgToml))
	if err != nil {
		return nil, err
	}

	tomlTree := config.Get(configType)

	switch typedTree := tomlTree.(type) {
	case []*toml.Tree:
		return typedTree, nil
	default:
		return nil, fmt.Errorf("unable to parse toml tree")
	}
}

type TomlWrapper struct {
	Overrides   []*toml.Tree
	Constraints []*toml.Tree
	SoloIo      []*toml.Tree
}

func ParseFullTomlFromDir(relativeDir string) (*TomlWrapper, error) {
	overrides, err := ParseTomlOverridesFromDir(relativeDir)
	if err != nil {
		return nil, err
	}
	constraints, err := ParseTomlFromDir(relativeDir)
	if err != nil {
		return nil, err
	}
	return &TomlWrapper{
		Constraints: constraints,
		Overrides:   overrides,
	}, nil
}

func ParseFullToml() (*TomlWrapper, error) {
	return ParseFullTomlFromDir("")
}

// For a toml file such as this:
//
// [[rootTableName]]
// name = "name-1"
// other = "some other value"
// [[rootTableName]]
// name = "name-2"
// other = "some other value 2"
//
// GetTomlValues("my-file.toml", "rootTableName", "name", "name-1")
// would return map[string]string{"name":"name-1","other":"some other value"}
func GetTomlValues(filename, rootTableName, identiferKey, identifierValue string) (map[string]string, error) {
	config, err := toml.LoadFile(filename)
	if err != nil {
		return nil, err
	}

	rawTomlParse := config.Get(rootTableName)
	var rootedTrees []*toml.Tree

	switch typedTree := rawTomlParse.(type) {
	case []*toml.Tree:
		rootedTrees = typedTree
	default:
		return nil, fmt.Errorf("unable to parse root toml tree")
	}
	output := make(map[string]string)
	for _, rootedTree := range rootedTrees {
		if rootedTree.Get(identiferKey) == identifierValue {
			for key, val := range rootedTree.ToMap() {
				switch typedVal := val.(type) {
				case string:
					output[key] = typedVal
				default:
					return nil, fmt.Errorf("nested or non-string element in toml: %v: %v where %v = %v, key = %v",
						filename, rootTableName, identiferKey, identifierValue, key)
				}
			}
		}
	}
	return output, nil
}

const (
	imageRepoConst    = "repo"
	imageTagConst     = "tag"
	imageAltNameConst = "altname"
)

// GetImageVersionFromToml extracts an image spec from a toml file
// the keys "name", "repo", and "tag" are required
// the key "altname" is optional, and will replace "name" if provided
func GetImageVersionFromToml(filename, imageName string) (string, error) {
	values, err := GetTomlValues(filename, imageConst, nameConst, imageName)
	if err != nil {
		return "", err
	}
	var repo, tag, altName string
	var ok bool
	if repo, ok = values[imageRepoConst]; !ok {
		return "", fmt.Errorf("no repo specified for image")
	}
	if tag, ok = values[imageTagConst]; !ok {
		return "", fmt.Errorf("no tag specified for image")
	}
	// Normally, we will just use the "name" key for the image name
	// However, that value should be unique across the toml file. If a toml file wants to specify more than one version
	// of the same image, the "altname" should be used to specify the image name while the "name" field provides a
	// unique identifier.
	if altName, ok = values[imageAltNameConst]; ok {
		imageName = altName
	}
	return fmt.Sprintf("%v/%v:%v", repo, imageName, tag), nil
}
