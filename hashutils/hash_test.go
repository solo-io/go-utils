package hashutils

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_hashutils "github.com/solo-io/go-utils/hashutils/mocks"
)

var _ = Describe("hash", func() {
	var (
		ctrl        *gomock.Controller
		safeHasher1 *mock_hashutils.MockSafeHasher
		safeHasher2 *mock_hashutils.MockSafeHasher
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		safeHasher1 = mock_hashutils.NewMockSafeHasher(ctrl)
		safeHasher2 = mock_hashutils.NewMockSafeHasher(ctrl)
	})
	Context("hashable equal", func() {
		It("will return not ok if passed in an object which is not a SafeHasher", func() {
			equal, ok := HashableEqual(safeHasher1, &notSafeHasher{})
			Expect(ok).To(BeFalse())
			Expect(equal).To(BeFalse())
		})
		It("will return false, ok if both fulfill interface, but are not equal", func() {
			safeHasher1.EXPECT().Hash(nil).Return(uint64(10), nil)
			safeHasher2.EXPECT().Hash(nil).Return(uint64(12), nil)
			equal, ok := HashableEqual(safeHasher1, safeHasher2)
			Expect(ok).To(BeTrue())
			Expect(equal).To(BeFalse())
		})
		It("will return true, ok if both fulfill interface, and are equal", func() {
			safeHasher1.EXPECT().Hash(nil).Return(uint64(10), nil)
			safeHasher2.EXPECT().Hash(nil).Return(uint64(10), nil)
			equal, ok := HashableEqual(safeHasher1, safeHasher2)
			Expect(ok).To(BeTrue())
			Expect(equal).To(BeTrue())
		})
	})

})

type notSafeHasher struct{}
