package clusterlock

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/google/uuid"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/avast/retry-go"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	// default namespace to install
	LockDefaultNamespace = "default"
	// name of kubernetes service holding the lock
	LockResourceName = "test-lock"
	// name of the annotation containing the lock
	LockAnnotationKey = "test.lock"
	// name of the annotation containing the timeout
	LockTimeoutAnnotationKey = "test.lock.timeout"

	// Default timeout for lock to be held
	DefaultLockTimeout = time.Second * 30
	DefaultTimeFormat  = time.RFC3339Nano

	// heartbeat settings
	DefaultHeartbeatTime = time.Second * 10
)

var defaultClusterLock = &ClusterLock{
	Name: LockResourceName,
}

var defaultOpts = []retry.Option{
	retry.Delay(10 * time.Second),
	retry.Attempts(60),
	retry.DelayType(retry.FixedDelay),
	retry.RetryIf(func(e error) bool {
		if e != nil {
			if IsLockIsUseError(e) {
				return true
			}
			if errors.IsConflict(e) {
				return true
			}
			if errors.IsNotFound(e) {
				return true
			}
		}
		return false
	}),
}

type TestClusterLocker struct {
	client  ClusterLockClient
	ownerId string
	ctx     context.Context
}

type Options struct {
	Namespace string
	IdPrefix  string
	Context   context.Context
}

// Deprecated: use NewKubeClusterLocker
func NewTestClusterLocker(clientset kubernetes.Interface, options Options) (*TestClusterLocker, error) {
	return NewKubeClusterLocker(clientset, options)
}

func NewKubeClusterLocker(clientset kubernetes.Interface, options Options) (*TestClusterLocker, error) {
	if options.Namespace == "" {
		options.Namespace = LockDefaultNamespace
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	client := &KubeClusterLockClient{
		namespace: options.Namespace,
		clientset: clientset,
	}
	return NewClusterLocker(options.Context, options.IdPrefix, client)
}

func NewConsulClusterLocker(ctx context.Context, idPrefix string, consul *api.Client) (*TestClusterLocker, error) {
	client := &ConsulClusterLockClient{
		client: consul,
	}
	return NewClusterLocker(ctx, idPrefix, client)
}

func NewClusterLocker(ctx context.Context, idPrefix string, client ClusterLockClient) (*TestClusterLocker, error) {
	ownerId := idPrefix + uuid.New().String()

	_, err := client.Create(ctx, defaultClusterLock)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	return &TestClusterLocker{
		client:  client,
		ownerId: ownerId,
		ctx:     ctx,
	}, nil
}

func (t *TestClusterLocker) AcquireLock(opts ...retry.Option) error {
	opts = append(defaultOpts, opts...)
	lockLoop := t.lockLoop()

	err := retry.Do(
		lockLoop,
		opts...,
	)

	if err == nil {
		// if lock is acquired send heartbeat
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(DefaultHeartbeatTime):
					if err := retry.Do(t.reacquireLock,
						retry.DelayType(retry.FixedDelay),
						retry.Attempts(3),
						retry.Delay(1*time.Second),
						retry.RetryIf(func(e error) bool {
							return errors.IsNotFound(e)
						}),
					); err != nil {

						contextutils.LoggerFrom(ctx).Errorw("could not reacquire lock", zap.Error(err))
						return
					}
				}
			}
		}(t.ctx)
	}

	return err

}

func (t *TestClusterLocker) reacquireLock() error {
	lock, err := t.client.Get(t.ctx, LockResourceName)
	if err != nil {
		return err
	}

	if lock.Empty() {
		return emptyLockReleaseError
	} else if lock.OwnerID != t.ownerId {
		return notLockOwnerError
	}

	lock.OwnerID = t.ownerId
	lock.Timeout = time.Now().Format(DefaultTimeFormat)

	if _, err = t.client.Update(t.ctx, lock); err != nil {
		return err
	}
	return nil
}

func (t *TestClusterLocker) lockLoop() retry.RetryableFunc {
	var callback = func() error {
		lock, err := t.concurrentLockGet()
		if err != nil && !errors.IsTimeout(err) {
			return err
		}

		if lock.Empty() {
			// Case if lock is are empty
			lock.Set(t.ownerId, time.Now().Format(DefaultTimeFormat))
		} else {
			if lock.Timeout != "" {
				// case if timeout has expired
				savedTime, err := time.Parse(DefaultTimeFormat, lock.Timeout)
				if err != nil {
					return err
				}
				if time.Since(savedTime) > DefaultLockTimeout {
					lock.Set(t.ownerId, time.Now().Format(DefaultTimeFormat))
				}
			}

			if lock.OwnerID != t.ownerId {
				return lockInUseError
			}
		}

		if _, err := t.client.Update(t.ctx, lock); err != nil {
			return err
		}
		return nil
	}
	return callback
}

func (t *TestClusterLocker) concurrentLockGet() (*ClusterLock, error) {
	originalLock, err := t.client.Get(t.ctx, LockResourceName)
	if err == nil {
		return originalLock, nil
	}
	if errors.IsNotFound(err) {
		newLock, err := t.client.Create(t.ctx, defaultClusterLock)
		if err != nil {
			// force the loop to restart
			if errors.IsAlreadyExists(err) {
				return nil, lockInUseError
			}
			// actual error to be handled above
			return nil, err
		}
		return newLock, nil
	}
	return nil, err
}

var (
	emptyLockReleaseError   = fmt.Errorf("cannot release empty lock")
	IsEmptyLockReleaseError = func(e error) bool {
		return e == emptyLockReleaseError
	}

	notLockOwnerError   = fmt.Errorf("only the lock owner can release the lock")
	IsNotLockOwnerError = func(e error) bool {
		return e == notLockOwnerError
	}

	lockInUseError   = fmt.Errorf("lock is currently in use")
	IsLockIsUseError = func(e error) bool {
		return e == lockInUseError
	}
)

func (t *TestClusterLocker) ReleaseLock() error {
	lock, err := t.client.Get(t.ctx, LockResourceName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if lock.Empty() {
		return emptyLockReleaseError
	} else if lock.OwnerID != t.ownerId {
		return notLockOwnerError
	}

	lock.Clear()

	if _, err := t.client.Update(t.ctx, lock); err != nil {
		return err
	}
	return nil
}
