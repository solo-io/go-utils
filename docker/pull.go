package docker

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
)

// PullIfNotPresent will pull an image if it is not present locally
// retrying up to retries times
// it returns true if it attempted to pull, and any errors from pulling
func PullIfNotPresent(ctx context.Context, image string, retries int) (pulled bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	// if this did not return an error, then the image exists locally
	cmd := Command("inspect", "--type=image", image)
	if err := cmd.Run(); err == nil {
		logger.Infof("Image: %s present locally", image)
		return false, nil
	}
	// otherwise try to pull it
	return true, Pull(ctx, image, retries)
}

// Pull pulls an image, retrying up to retries times
func Pull(ctx context.Context, image string, retries int) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Pulling image: %s ...", image)
	err := Command("pull", image).Run()
	// retry pulling up to retries times if necessary
	if err != nil {
		for i := 0; i < retries; i++ {
			time.Sleep(time.Second * time.Duration(i+1))
			logger.Warnf(errors.Wrapf(err, "Trying again to pull image: %s ...", image).Error())
			err = Command("pull", image).Run()
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		logger.Warnf(errors.Wrapf(err, "Failed to pull image: %s", image).Error())
	}
	return err
}
