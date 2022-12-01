package commands

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils/executils"
	"github.com/solo-io/go-utils/securityscanutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func ScanVersionCommand(ctx context.Context, rootOptions *RootOptions) *cobra.Command {
	opts := &scanVersionOptions{
		RootOptions: rootOptions,
	}

	cmd := &cobra.Command{
		Use:   "scan-version",
		Short: "Run Trivy scans against images for a single version",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := doScanVersion(ctx, opts)
			if err != nil {
				return err
			}
			if results != nil {
				contextutils.LoggerFrom(ctx).Infof(results.Error())
			} else {
				contextutils.LoggerFrom(ctx).Infof("No vulnerabilities found")
			}
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

type scanResults error

func doScanVersion(ctx context.Context, opts *scanVersionOptions) (scanResults, error) {
	contextutils.LoggerFrom(ctx).Infof("Starting ScanVersion with version=%s", opts.imageVersion)

	trivyScanner := securityscanutils.NewTrivyScanner(executils.CombinedOutputWithStatus)

	templateFile, err := securityscanutils.GetTemplateFile(securityscanutils.MarkdownTrivyTemplate)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(templateFile)
	}()

	versionedOutputDir := path.Join(securityscanutils.OutputScanDirectory, opts.imageVersion)
	contextutils.LoggerFrom(ctx).Infof("Results will be written to %s", versionedOutputDir)
	err = os.MkdirAll(versionedOutputDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	var results scanResults
	for _, imageName := range opts.imageNames {
		image := fmt.Sprintf("%s/%s:%s", opts.imageRepository, imageName, opts.imageVersion)
		outputFile := path.Join(versionedOutputDir, fmt.Sprintf("%s.txt", imageName))

		scanCompleted, vulnerabilityFound, scanErr := trivyScanner.ScanImage(ctx, image, templateFile, outputFile)
		contextutils.LoggerFrom(ctx).Infof(
			"Scanned Image: %v, ScanCompleted: %v, VulnerabilityFound: %v, Error: %v",
			image, scanCompleted, vulnerabilityFound, scanErr)

		if vulnerabilityFound {
			results = multierror.Append(results, eris.Errorf("vulnerabilities found for %s", image))
		}
	}

	return results, nil
}
