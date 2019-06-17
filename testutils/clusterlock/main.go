package clusterlock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/avast/retry-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
	DefaultLockTimeout = time.Second*30
	DefaultTimeFormat  = time.RFC3339Nano

	// heartbeat settings
	DefaultHeartbeatTime = time.Second * 15
)

var defaultConfigMap = &coreV1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name: LockResourceName,
	},
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
	clientset kubernetes.Interface
	namespace string
	buidldId  string
	ctx       context.Context
}

type Options struct {
	Namespace string
	IdPrefix string
	Context context.Context
}

func NewTestClusterLocker(clientset kubernetes.Interface, options Options) (*TestClusterLocker, error) {


	if options.Namespace == "" {
		options.Namespace = LockDefaultNamespace
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	buildId := options.IdPrefix + uuid.New().String()
	_, err := clientset.CoreV1().ConfigMaps(options.Namespace).Create(defaultConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	return &TestClusterLocker{clientset: clientset, namespace: options.Namespace, buidldId: buildId, ctx: options.Context}, nil
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
					if err := t.reacquireLock(); err != nil {
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
	cfgMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(LockResourceName, v1.GetOptions{})
	if err != nil {
		return err
	}

	if cfgMap.Annotations == nil || len(cfgMap.Annotations) == 0 {
		return emptyLockReleaseError
	} else if val, ok := cfgMap.Annotations[LockAnnotationKey]; ok && val != t.buidldId {
		return notLockOwnerError
	}

	cfgMap.Annotations = map[string]string{
		LockAnnotationKey:        t.buidldId,
		LockTimeoutAnnotationKey: time.Now().Format(DefaultTimeFormat),
	}
	if _, err = t.clientset.CoreV1().ConfigMaps(t.namespace).Update(cfgMap); err != nil {
		return err
	}
	return nil
}

func (t *TestClusterLocker) lockLoop() retry.RetryableFunc {
	var callback = func() error {
		cfgMap, err := t.concurrentLockGet()
		if err != nil && !errors.IsTimeout(err) {
			return err
		}

		if cfgMap.Annotations == nil || len(cfgMap.Annotations) == 0 {
			// Case if annotations are empty
			cfgMap.Annotations = map[string]string{
				LockAnnotationKey:        t.buidldId,
				LockTimeoutAnnotationKey: time.Now().Format(DefaultTimeFormat),
			}

		} else {
			if timeoutStr, ok := cfgMap.Annotations[LockTimeoutAnnotationKey]; ok {
				// case if timeout has expired
				savedTime, err := time.Parse(DefaultTimeFormat, timeoutStr)
				if err != nil {
					return err
				}
				if time.Since(savedTime) > DefaultLockTimeout {
					cfgMap.Annotations = map[string]string{
						LockAnnotationKey:        t.buidldId,
						LockTimeoutAnnotationKey: time.Now().Format(DefaultTimeFormat),
					}
				}
			}

			if val, ok := cfgMap.Annotations[LockAnnotationKey]; ok && val != t.buidldId {
				return lockInUseError
			}
		}

		if _, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Update(cfgMap); err != nil {
			return err
		}
		return nil
	}
	return callback
}

func (t *TestClusterLocker) concurrentLockGet() (*coreV1.ConfigMap, error) {
	originalConfigMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(LockResourceName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newConfigMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Create(defaultConfigMap)
			if err != nil {
				// force the loop to restart
				if errors.IsAlreadyExists(err) {
					return nil, lockInUseError
				}
				// actual error to be handled above
				return nil, err
			}
			return newConfigMap, nil
		}
		return nil, err
	}
	return originalConfigMap, nil
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

	lockInUseError = fmt.Errorf("lock is currently in use")
	IsLockIsUseError = func(e error) bool {
		return e == lockInUseError
	}
)

func (t *TestClusterLocker) ReleaseLock() error {
	_, cancel := context.WithCancel(t.ctx)
	cancel()
	cfgMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(LockResourceName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if cfgMap.Annotations == nil || len(cfgMap.Annotations) == 0 {
		return emptyLockReleaseError
	} else if val, ok := cfgMap.Annotations[LockAnnotationKey]; ok && val != t.buidldId {
		return notLockOwnerError
	}

	if _, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Update(defaultConfigMap); err != nil {
		return err
	}
	return nil
}
