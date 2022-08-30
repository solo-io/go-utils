package commands

import (
    "context"
    "fmt"
    "github.com/hashicorp/go-multierror"
    "github.com/rotisserie/eris"
    "github.com/solo-io/go-utils/cliutils"
    "github.com/solo-io/go-utils/contextutils"
    "github.com/solo-io/go-utils/osutils/executils"
    "github.com/solo-io/go-utils/securityscanutils"
    "github.com/solo-io/go-utils/securityscanutils/internal"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
    "os"
    "path"
)

func ScanVersionCommand(ctx context.Context, globalFlags *internal.GlobalFlags) *cobra.Command {
    opts := &scanVersionOptions{
        GlobalFlags: globalFlags,
    }

    cmd := &cobra.Command{
        Use:     "scan-version",
        Short:   "Run a Trivy scan (only reports HIGH and CRITICAL-level vulnerabilities) against a set of images for a single version",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doScanVersion(ctx, opts)
        },
    }

    opts.addToFlags(cmd.Flags())

    return cmd
}

type scanVersionOptions struct {
    *internal.GlobalFlags

    imageRepository string
    imageVersion string
    imageNames []string
}

func (o *scanVersionOptions) addToFlags(flags *pflag.FlagSet) {
    flags.StringVarP(&o.imageRepository, "image-repo", "r", securityscanutils.QuayRepository, "image repository to scan")
    flags.StringVarP(&o.imageVersion, "version", "t", "", "version to scan")
    flags.StringSliceVar(&o.imageNames, "images", []string{}, "comma separated list of images to scan")

    cliutils.MustMarkFlagRequired(flags, "version")
    cliutils.MustMarkFlagRequired(flags, "images")
}

func doScanVersion(ctx context.Context, opts *scanVersionOptions) error {
    contextutils.LoggerFrom(ctx).Infof("Starting ScanVersion with version=%s", opts.imageVersion)

    trivyScanner := securityscanutils.NewTrivyScanner(executils.CombinedOutputWithStatus)

    templateFile, err := securityscanutils.GetTemplateFile(securityscanutils.MarkdownTrivyTemplate)
    if err != nil {
        return err
    }
    defer func() {
        _ = os.Remove(templateFile)
    }()

    versionedOutputDir := path.Join(securityscanutils.OutputScanDirectory, opts.imageVersion)
    err = os.MkdirAll(versionedOutputDir, os.ModePerm)
    if err != nil {
        return err
    }

    var scanResults error
    for _, imageName := range opts.imageNames {
        image := fmt.Sprintf("%s/%s:%s", opts.imageRepository, imageName, opts.imageVersion)
        outputFile := path.Join(versionedOutputDir, fmt.Sprintf("%s.txt", imageName))

        scanCompleted, vulnerabilityFound, scanErr := trivyScanner.ScanImage(ctx, image, templateFile, outputFile)
        contextutils.LoggerFrom(ctx).Infof(
            "Scanned Image: %v, ScanCompleted: %v, VulnerabilityFound: %v, Error: %v",
            image, scanCompleted, vulnerabilityFound, scanErr)

        if vulnerabilityFound {
            scanResults = multierror.Append(scanResults, eris.Errorf("vulnerabilities found for %s", image))
        }
    }
    return scanResults
}