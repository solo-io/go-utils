package testutils

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kubernetes", func() {
	Context("WaitPodsRunning", func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		BeforeEach(func() {
			kubectlExecutable = `echo`
			ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		})
		AfterEach(func() {
			cancel()
		})
		It("Reports running pods as running", func() {
			kubectlArgs = []string{`"\
NAMESPACE            NAME                                              READY   STATUS        RESTARTS   AGE \
kube-system          coredns-5dd5756b68-62pwz                          1/1     Running       0          3d18h" #`}

			// The parameters here don't really matter since we're not actually calling kubectl.
			// We do need a label however to enter the loop that calls the command.
			Expect(WaitPodsRunning(ctx, time.Millisecond, "", "foo=bar")).ToNot(HaveOccurred())
		})
		It("Doesn't report terminating pods as running", func() {
			kubectlArgs = []string{`"\
NAMESPACE            NAME                                              READY   STATUS        RESTARTS   AGE \
kube-system          coredns-5dd5756b68-jlm7r                          1/1     Terminating   0          3d18h \
kube-system          coredns-5dd5756b68-62pwz                          1/1     Running       0          3d18h" #`}

			// The parameters here don't really matter since we're not actually calling kubectl.
			// We do need a label however to enter the loop that calls the command.
			Expect(WaitPodsRunning(ctx, time.Millisecond, "", "foo=bar")).To(MatchError(ContainSubstring("timed out waiting")))
		})
	})
})
