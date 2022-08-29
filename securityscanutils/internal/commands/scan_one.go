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

func ScanOneCommand(ctx context.Context, globalFlags *internal.GlobalFlags) *cobra.Command {
    opts := &scanOneOptions{
        GlobalFlags: globalFlags,
    }

    cmd := &cobra.Command{
        Use:     "scan-one",
        Aliases: []string{"scan"},
        Short:   "TODO",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doScanOne(ctx, opts)
        },
    }

    opts.addToFlags(cmd.Flags())

    return cmd
}

type scanOneOptions struct {
    *internal.GlobalFlags

    imageRepository string
    imageVersion string
    imageNames []string
}

func (o *scanOneOptions) addToFlags(flags *pflag.FlagSet) {
    flags.StringVarP(&o.imageRepository, "repository", "r", securityscanutils.QUAY_REPOSITORY, "image repository to scan")
    flags.StringVarP(&o.imageRepository, "version", "v", "", "version to scan")
    flags.StringSliceVar(&o.imageNames, "images", []string{}, "comma separated list of images to scan")

    cliutils.MustMarkFlagRequired(flags, "version")
    cliutils.MustMarkFlagRequired(flags, "images")
}

func doScanOne(ctx context.Context, opts *scanOneOptions) error {
    contextutils.LoggerFrom(ctx).Infof("Starting ScanOne with version=%s", opts.imageVersion)

    trivyScanner := securityscanutils.NewTrivyScanner(executils.CombinedOutputWithStatus)

    templateFile, err := securityscanutils.GetTemplateFile(securityscanutils.MarkdownTrivyTemplate)
    if err != nil {
        return err
    }
    defer func() {
        _ = os.Remove(templateFile)
    }()

    versionedOutputDir := path.Join(securityscanutils.OUTPUT_SCAN_DIRECTORY, opts.imageVersion)
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
            "Scaned Image: %v, ScanCompleted: %v, VulnerabilityFound: %v, Error: %v",
            image, scanCompleted, vulnerabilityFound, scanErr)

        if vulnerabilityFound {
            scanResults = multierror.Append(scanResults, eris.Errorf("vulnerabilities found for %s", image))
        }
    }
    return scanResults
}