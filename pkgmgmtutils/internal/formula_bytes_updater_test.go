package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types"
	"github.com/solo-io/go-utils/pkgmgmtutils/internal"
)

var _ = Describe("FormulaBytesUpdater", func() {
	It("can replace submatch", func() {
		regex := `version\s*"([0-9.]+)"`
		testData := `Some other data version "1.2.3" and even more data`

		b, err := internal.UpdateFormulaBytes(
			[]byte(testData),
			"4.5.6",
			"git-commit-sha",
			&formula_updater_types.PerPlatformSha256{},
			&formula_updater_types.FormulaOptions{
				VersionRegex: regex,
			})

		Expect(err).NotTo(HaveOccurred())
		Expect(b).To(Equal([]byte(`Some other data version "4.5.6" and even more data`)))
	})

	It("can replace submatch fail", func() {
		regex := `version\s*"([a-z.]+)"`
		testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

		b, err := internal.UpdateFormulaBytes(
			[]byte(testData),
			"7.8.9",
			"git-commit-sha",
			&formula_updater_types.PerPlatformSha256{},
			&formula_updater_types.FormulaOptions{VersionRegex: regex},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(b).To(Equal([]byte(testData)))
	})

	It("can replace submatch fail", func() {
		regex := `version\s*"[0-9.]+"`
		testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

		Expect(func() {
			_, _ = internal.UpdateFormulaBytes(
				[]byte(testData),
				"7.8.9",
				"git-commit-sha",
				&formula_updater_types.PerPlatformSha256{},
				&formula_updater_types.FormulaOptions{VersionRegex: regex},
			)
		}).To(Panic())
	})
})
