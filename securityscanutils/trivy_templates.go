package securityscanutils

import (
	"os"

	"github.com/rotisserie/eris"
)

// Template for markdown docs
const MarkdownTrivyTemplate = `{{- if . }}
{{- range . }}
{{- if (eq (len .Vulnerabilities) 0) }}

No Vulnerabilities Found for {{.Target}}
{{- else }}

Vulnerabilities Listed for {{.Target}}

Vulnerability ID|Package|Severity|Installed Version|Fixed Version|Reference
---|---|---|---|---|---
{{- range .Vulnerabilities }}
{{ .VulnerabilityID }}|{{ .PkgName }}|{{ .Vulnerability.Severity }}|{{ .InstalledVersion }}|{{ .FixedVersion }}|{{ .PrimaryURL }}
{{- end }}
{{- end }}
{{- end }}
{{- else }}
Trivy Returned Empty Report
{{- end }}`

// Create tempoarary file that contains the trivy template
// Trivy CLI only accepts files as input for a template, so this is a workaround
func GetTemplateFile(trivyTemplate string) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", eris.Wrap(err, "Unable to create temporary file to write template to")
	}
	_, err = f.Write([]byte(trivyTemplate))
	if err != nil {
		return "", eris.Wrapf(err, "Unable to write template to file %s", f.Name())
	}
	f.Close()
	return f.Name(), nil
}
