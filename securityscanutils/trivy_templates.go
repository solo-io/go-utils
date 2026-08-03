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

// Create temporary file that contains the trivy template.
// Trivy requires custom template files to use the .tpl extension.
func GetTemplateFile(trivyTemplate string) (string, error) {
	f, err := os.CreateTemp("", "trivy-*.tpl")
	if err != nil {
		return "", eris.Wrap(err, "Unable to create temporary file to write template to")
	}
	templateFile := f.Name()
	if _, err = f.WriteString(trivyTemplate); err != nil {
		_ = f.Close()
		_ = os.Remove(templateFile)
		return "", eris.Wrapf(err, "Unable to write template to file %s", f.Name())
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(templateFile)
		return "", eris.Wrapf(err, "Unable to close template file %s", templateFile)
	}
	return templateFile, nil
}
