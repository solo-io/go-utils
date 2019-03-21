package clusterlock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

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
	DefaultLockTimeout = time.Minute/2
	DefaultTimeFormat  = time.RFC3339Nano

	// heartbeat settings
	DefaultHeartbeatTime = time.Second * 30
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
		}
		return false
	}),
}

var lockInUseError = fmt.Errorf("lock is currently in use")
var IsLockIsUseError = func(e error) bool {
	return e == lockInUseError
}

type TestClusterLocker struct {
	clientset kubernetes.Interface
	namespace string
	buidldId  string
	ctx       context.Context
}

func NewTestClusterLocker(ctx context.Context, clientset kubernetes.Interface, namespace string) (*TestClusterLocker, error) {
	if namespace == "" {
		namespace = LockDefaultNamespace
	}
	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(defaultConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	buildId := uuid.New().String()
	return &TestClusterLocker{clientset: clientset, namespace: namespace, buidldId: buildId, ctx: ctx}, nil
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
		cfgMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(LockResourceName, v1.GetOptions{})
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

		if _, err = t.clientset.CoreV1().ConfigMaps(t.namespace).Update(cfgMap); err != nil {
			return err
		}
		return nil
	}
	return callback
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
)

func (t *TestClusterLocker) ReleaseLock() error {
	_, cancel := context.WithCancel(t.ctx)
	cancel()
	cfgMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(LockResourceName, v1.GetOptions{})
	if err != nil {
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
