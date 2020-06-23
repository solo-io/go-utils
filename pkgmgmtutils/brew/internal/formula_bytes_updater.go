package internal

import (
	"bytes"
	"regexp"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/pkgmgmtutils/brew/formula_updater_types"
)

var (
	ErrAlreadyUpdated               = eris.New("pkgmgmtutils: formula already updated")
	ErrMissingRequiredVersion       = eris.New("pkgmgmtutils: missing required version")
	ErrMissingRequiredVersionTagSha = eris.New("pkgmgmtutils: missing required version tag sha")
	ErrMissingRequiredDarwinSha     = eris.New("pkgmgmtutils: missing required sha for darwin binary")
	ErrMissingRequiredLinuxSha      = eris.New("pkgmgmtutils: missing required sha for linux binary")
	ErrMissingRequiredWindowsSha    = eris.New("pkgmgmtutils: missing required sha for windows binary")
)

func UpdateFormulaBytes(byt []byte, version string, versionSha string, shas *formula_updater_types.PerPlatformSha256, fOpt *formula_updater_types.FormulaOptions) ([]byte, error) {
	// Update Version
	if fOpt.VersionRegex != "" {
		if version == "" {
			return nil, ErrMissingRequiredVersion
		}

		re := regexp.MustCompile(fOpt.VersionRegex)

		// Check if formula has already been updated
		if matches := re.FindSubmatch(byt); len(matches) > 1 && bytes.Compare(matches[1], []byte(version)) == 0 {
			return byt, ErrAlreadyUpdated
		}

		byt = replaceSubmatch(byt, []byte(version), re)
	}

	// Update Version SHA (git tag sha)
	if fOpt.VersionShaRegex != "" {
		if versionSha == "" {
			return nil, ErrMissingRequiredVersionTagSha
		}

		byt = replaceSubmatch(byt, []byte(versionSha), regexp.MustCompile(fOpt.VersionShaRegex))
	}

	// Update Mac SHA256
	if fOpt.DarwinShaRegex != "" {
		if len(shas.DarwinSha) == 0 {
			return nil, ErrMissingRequiredDarwinSha
		}

		byt = replaceSubmatch(byt, []byte(shas.DarwinSha), regexp.MustCompile(fOpt.DarwinShaRegex))
	}

	// Update Linux SHA256
	if fOpt.LinuxShaRegex != "" {
		if len(shas.LinuxSha) == 0 {
			return nil, ErrMissingRequiredLinuxSha
		}

		byt = replaceSubmatch(byt, []byte(shas.LinuxSha), regexp.MustCompile(fOpt.LinuxShaRegex))
	}

	// Update Windows SHA256
	if fOpt.WindowsShaRegex != "" {
		if len(shas.WindowsSha) == 0 {
			return nil, ErrMissingRequiredWindowsSha
		}

		byt = replaceSubmatch(byt, []byte(shas.WindowsSha), regexp.MustCompile(fOpt.WindowsShaRegex))
	}

	return byt, nil
}

// replaceSubmatch will replace the submatch group with repl within the matching regular expression of the src input.
// It expects there to be only one submatch group within the re regular expression.
// Since there's no golang replaceAllSubmatch function, need to do this double regex hack
func replaceSubmatch(src []byte, repl []byte, re *regexp.Regexp) []byte {
	return re.ReplaceAllFunc(src, func(m []byte) []byte {
		ind := re.FindSubmatchIndex(m)
		if ind == nil {
			return src
		}

		if len(ind) != 4 || ind[0] != 0 || ind[1] != len(m) {
			panic("Regular expression does not meet preconditions")
		}

		var b bytes.Buffer
		b.Grow(ind[2] + len(repl) + (len(m) - ind[3]))
		b.Write(m[:ind[2]])
		b.Write(repl)
		b.Write(m[ind[3]:])
		return b.Bytes()
	})
}
