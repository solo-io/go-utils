package commands

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils/executils"
	"github.com/solo-io/go-utils/securityscanutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"path"
)

func ScanVersionCommand(ctx context.Context, rootOptions *RootOptions) *cobra.Command {
	opts := &scanVersionOptions{
		RootOptions: rootOptions,
	}

	cmd := &cobra.Command{
		Use:   "scan-version",
		Short: "Run Trivy scans against images for a single version",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := doScanVersion(ctx, opts)
			if err != nil {
				return err
			}

			contextutils.LoggerFrom(ctx).Infof(result.String())
			return nil
		},
	}

	opts.addToFlags(cmd.Flags())

	return cmd
}

type scanVersionOptions struct {
	*RootOptions

	imageRepository string
	imageVersion    string
	imageNames      []string
}

func (o *scanVersionOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.imageRepository, "image-repo", "r", securityscanutils.QuayRepository, "image repository to scan")
	flags.StringVarP(&o.imageVersion, "version", "t", "", "version to scan")
	flags.StringSliceVar(&o.imageNames, "images", []string{}, "comma separated list of images to scan")

	cliutils.MustMarkFlagRequired(flags, "version")
	cliutils.MustMarkFlagRequired(flags, "images")
}

type scanVersionResult struct {
	OutputDir                 string
	ImagesWithVulnerabilities []string
}

func (s scanVersionResult) String() string {
	if len(s.ImagesWithVulnerabilities) == 0 {
		return "No vulnerabilities found"
	}

	return fmt.Sprintf("Vulernabilities found! Affected images: %v. Formatted results: %s",
		s.ImagesWithVulnerabilities, s.OutputDir)
}

func doScanVersion(ctx context.Context, opts *scanVersionOptions) (scanVersionResult, error) {
	contextutils.LoggerFrom(ctx).Infof("Starting ScanVersion with version=%s", opts.imageVersion)
	versionedOutputDir := path.Join(securityscanutils.OutputScanDirectory, opts.imageVersion)

	result := scanVersionResult{
		OutputDir: versionedOutputDir,
	}

	trivyScanner := securityscanutils.NewTrivyScanner(executils.CombinedOutputWithStatus)

	templateFile, err := securityscanutils.GetTemplateFile(securityscanutils.MarkdownTrivyTemplate)
	if err != nil {
		return result, err
	}
	defer func() {
		_ = os.Remove(templateFile)
	}()

	err = os.MkdirAll(versionedOutputDir, os.ModePerm)
	if err != nil {
		return result, err
	}

	for _, imageName := range opts.imageNames {
		image := fmt.Sprintf("%s/%s:%s", opts.imageRepository, imageName, opts.imageVersion)
		outputFile := path.Join(versionedOutputDir, fmt.Sprintf("%s.txt", imageName))

		scanCompleted, vulnerabilityFound, scanErr := trivyScanner.ScanImage(ctx, image, templateFile, outputFile)
		contextutils.LoggerFrom(ctx).Debugf(
			"Scanned Image: %v, ScanCompleted: %v, VulnerabilityFound: %v, Error: %v",
			image, scanCompleted, vulnerabilityFound, scanErr)

		if vulnerabilityFound {
			result.ImagesWithVulnerabilities = append(result.ImagesWithVulnerabilities, image)
		}
	}

	return result, nil
}
