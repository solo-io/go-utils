package pkgmgmtutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("package management utils", func() {
	// Keeping this example end-2-end tests as we do not want CI tests that mess with GitHub
	/*	Context("using full GitHub APIs", func() {
		It("glooctl can get latest release version", func() {
			// Create test sha256 files
			testData := map[string]struct {
				filename   string
				sha        string
				binaryName string
			}{
				"darwin": {
					filename:   "glooctl-darwin-amd64.sha256",
					sha:        "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99",
					binaryName: "glooctl-darwin-amd64",
				},
				"linux": {
					filename:   "glooctl-linux-amd64.sha256",
					sha:        "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f",
					binaryName: "glooctl-linux-amd64",
				},
				"windows": {
					filename:   "glooctl-windows-amd64.exe.sha256",
					sha:        "031434d831a394af2b7882b6f1a220e34efc91c4e4ef807a530fc8ec7990d2ca",
					binaryName: "glooctl-windows-amd64.exe",
				},
			}

			dirTmp, err := ioutil.TempDir("", "_output")
			Expect(err).To(BeNil())
			defer os.RemoveAll(dirTmp)

			for _, v := range testData {
				data := fmt.Sprintf("%s %s", v.sha, v.binaryName)
				err = ioutil.WriteFile(filepath.Join(dirTmp, v.filename), []byte(data), 0644)
				Expect(err).To(BeNil())
			}

			// https://regex101.com/ is a helpful tool to validating golang regex

			fOpts := []FormulaOptions{
				{
					Name:           "homebrew-tap/glooctl",
					FormulaName:    "glooctl",
					Path:           "Formula/glooctl.rb",
					RepoOwner:      "solo-io",           // Make change in this repo
					RepoName:       "test-homebrew-tap", // Assumes this repo is forked from PRRepo if PRRepoOwner != RepoOwner
					PRRepoOwner:    "solo-io",           // Make PR to this repo
					PRRepoName:     "test-homebrew-tap",
					PRBranch:       "master",
					PRDescription:  "/wip",
					PRCommitName:   "Solo-io Bot",
					PRCommitEmail:  "bot@solo.io",
					VersionRegex:   `version\s*"([0-9.]+)"`,
					DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
					LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,

					dryRun: true, // do NOT create a PR
				},
				{
					Name:            "fish-food/glooctl",
					FormulaName:     "glooctl",
					Path:            "Food/glooctl.lua",
					RepoOwner:       "solo-io",
					RepoName:        "test-fish-food",
					PRRepoOwner:     "solo-io",
					PRRepoName:      "test-fish-food",
					PRBranch:        "master",
					PRDescription:   "/wip",
					PRCommitName:    "Solo-io Bot",
					PRCommitEmail:   "bot@solo.io",
					VersionRegex:    `version\s*=\s*"([0-9.]+)"`,
					DarwinShaRegex:  `os\s*=\s*"darwin",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
					LinuxShaRegex:   `os\s*=\s*"linux",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
					WindowsShaRegex: `os\s*=\s*"windows",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,

					dryRun: true, // do NOT create a PR
				},
				{
					Name:            "homebrew-core/glooctl",
					FormulaName:     "glooctl",
					Path:            "Formula/glooctl.rb",
					RepoOwner:       "solo-io",
					RepoName:        "test-homebrew-core",
					PRRepoOwner:     "solo-io",
					PRRepoName:      "test-homebrew-core",
					PRBranch:        "master",
					PRDescription:   "/wip",
					PRCommitName:    "Solo-io Bot",
					PRCommitEmail:   "bot@solo.io",
					VersionRegex:    `:tag\s*=>\s*"v([0-9.]+)",`,
					VersionShaRegex: `:revision\s*=>\s*"(.*)"`,

					dryRun: true, // do NOT create a PR
				},
			}

			status, err := UpdateFormulas("solo-io", "gloo", dirTmp, `glooctl-(darwin|linux|windows).*\.sha256`, fOpts)
			Expect(err).NotTo(HaveOccurred())
			if err != nil {
				fmt.Println(err)
			}

			for _, s := range status {
				fmt.Printf("Name: %s; Updated: %t", s.Name, s.Updated)
				if s.Err != nil {
					fmt.Println(err)
				}
				Expect(s.Updated).To(BeTrue())
				Expect(s.Err).To(SatisfyAny(
					BeNil(),
					Equal(ErrAlreadyUpdated)))
			}

			// Cleanup
			ctx := context.Background()
			client, err := githubutils.GetClient(ctx)
			if err != nil {
				return
			}

			version := versionutils.GetReleaseVersionOrExitGracefully().String()[1:]
			for _, fOpt := range fOpts {
				branchName := "refs/heads/" + fOpt.FormulaName + "-" + version
				_, err = client.Git.DeleteRef(ctx, fOpt.RepoOwner, fOpt.RepoName, branchName)
				if err != nil {
					fmt.Printf("Error: failed to delete branch %s from repo %s:%s; Error was: %s", branchName,
						fOpt.RepoOwner, fOpt.RepoName, err.Error())
				}
			}
		})

		It("squashctl can get latest release version", func() {
			// Create test sha256 files
			testData := map[string]struct {
				filename   string
				sha        string
				binaryName string
			}{
				"darwin": {
					filename:   "squashctl-darwin.sha256",
					sha:        "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99",
					binaryName: "squashctl-darwin",
				},
				"linux": {
					filename:   "squashctl-linux.sha256",
					sha:        "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f",
					binaryName: "squashctl-linux",
				},
				"windows": {
					filename:   "squashctl-windows.exe.sha256",
					sha:        "031434d831a394af2b7882b6f1a220e34efc91c4e4ef807a530fc8ec7990d2ca",
					binaryName: "squashctl-windows.exe",
				},
			}

			dirTmp, err := ioutil.TempDir("", "_output")
			Expect(err).To(BeNil())
			defer os.RemoveAll(dirTmp)

			for _, v := range testData {
				data := fmt.Sprintf("%s %s", v.sha, v.binaryName)
				err = ioutil.WriteFile(filepath.Join(dirTmp, v.filename), []byte(data), 0644)
				Expect(err).To(BeNil())
			}

			// https://regex101.com/ is a helpful tool to validating golang regex

			fOpts := []FormulaOptions{
				{
					Name:           "homebrew-tap/squashctl",
					FormulaName:    "squashctl",
					Path:           "Formula/squashctl.rb",
					RepoOwner:      "solo-io",            // Make change in this repo
					RepoName:       "test-homebrew-tap", // Assumes this repo is forked from PRRepo if PRRepoOwner != RepoOwner
					PRRepoOwner:    "solo-io",            // Make PR to this repo
					PRRepoName:     "test-homebrew-tap",
					PRBranch:       "master",
					PRDescription:  "/wip",
					PRCommitName:   "Solo-io Bot",
					PRCommitEmail:  "bot@solo.io",
					VersionRegex:   `version\s*"([0-9.]+)"`,
					DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
					LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,

					dryRun: true, // do NOT create a PR
				},
			}

			status, err := UpdateFormulas("solo-io", "gloo", dirTmp, `squashctl-(darwin|linux|windows).*\.sha256`, fOpts)
			Expect(err).NotTo(HaveOccurred())
			if err != nil {
				fmt.Println(err)
			}

			for _, s := range status {
				fmt.Printf("Name: %s; Updated: %t", s.Name, s.Updated)
				if s.Err != nil {
					fmt.Println(err)
				}
				Expect(s.Updated).To(BeTrue())
				Expect(s.Err).To(SatisfyAny(
					BeNil(),
					Equal(ErrAlreadyUpdated)))
			}

			// Cleanup
			ctx := context.Background()
			client, err := githubutils.GetClient(ctx)
			if err != nil {
				return
			}

			version := versionutils.GetReleaseVersionOrExitGracefully().String()[1:]
			for _, fOpt := range fOpts {
				branchName := "refs/heads/" + fOpt.FormulaName + "-" + version
				_, err = client.Git.DeleteRef(ctx, fOpt.RepoOwner, fOpt.RepoName, branchName)
				if err != nil {
					fmt.Printf("Error: failed to delete branch %s from repo %s:%s; Error was: %s", branchName,
						fOpt.RepoOwner, fOpt.RepoName, err.Error())
				}
			}
		})

		It("supergloo can get latest release version", func() {
			// Create test sha256 files
			testData := map[string]struct {
				filename   string
				sha        string
				binaryName string
			}{
				"darwin": {
					filename:   "supergloo-cli-darwin-amd64.sha256",
					sha:        "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99",
					binaryName: "supergloo-cli-darwin-amd64",
				},
				"linux": {
					filename:   "supergloo-cli-linux-amd64.sha256",
					sha:        "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f",
					binaryName: "supergloo-cli-linux-amd64",
				},
				"windows": {
					filename:   "supergloo-cli-windows-amd64.exe.sha256",
					sha:        "031434d831a394af2b7882b6f1a220e34efc91c4e4ef807a530fc8ec7990d2ca",
					binaryName: "supergloo-cli-windows-amd64.exe",
				},
			}

			dirTmp, err := ioutil.TempDir("", "_output")
			Expect(err).To(BeNil())
			defer os.RemoveAll(dirTmp)

			for _, v := range testData {
				data := fmt.Sprintf("%s %s", v.sha, v.binaryName)
				err = ioutil.WriteFile(filepath.Join(dirTmp, v.filename), []byte(data), 0644)
				Expect(err).To(BeNil())
			}

			// https://regex101.com/ is a helpful tool to validating golang regex

			fOpts := []FormulaOptions{
				{
					Name:           "homebrew-tap/supergloo",
					FormulaName:    "supergloo",
					Path:           "Formula/supergloo.rb",
					RepoOwner:      "solo-io",            // Make change in this repo
					RepoName:       "test-homebrew-tap", // Assumes this repo is forked from PRRepo if PRRepoOwner != RepoOwner
					PRRepoOwner:    "solo-io",            // Make PR to this repo
					PRRepoName:     "test-homebrew-tap",
					PRBranch:       "master",
					PRDescription:  "/wip",
					PRCommitName:   "Solo-io Bot",
					PRCommitEmail:  "bot@solo.io",
					VersionRegex:   `version\s*"([0-9.]+)"`,
					DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
					LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,

					dryRun: true, // do NOT create a PR
				},
			}

			status, err := UpdateFormulas("solo-io", "gloo", dirTmp, `supergloo-cli-(darwin|linux|windows).*\.sha256`, fOpts)
			Expect(err).NotTo(HaveOccurred())
			if err != nil {
				fmt.Println(err)
			}

			for _, s := range status {
				fmt.Printf("Name: %s; Updated: %t", s.Name, s.Updated)
				if s.Err != nil {
					fmt.Println(err)
				}
				Expect(s.Updated).To(BeTrue())
				Expect(s.Err).To(SatisfyAny(
					BeNil(),
					Equal(ErrAlreadyUpdated)))
			}

			// Cleanup
			ctx := context.Background()
			client, err := githubutils.GetClient(ctx)
			if err != nil {
				return
			}

			version := versionutils.GetReleaseVersionOrExitGracefully().String()[1:]
			for _, fOpt := range fOpts {
				branchName := "refs/heads/" + fOpt.FormulaName + "-" + version
				_, err = client.Git.DeleteRef(ctx, fOpt.RepoOwner, fOpt.RepoName, branchName)
				if err != nil {
					fmt.Printf("Error: failed to delete branch %s from repo %s:%s; Error was: %s", branchName,
						fOpt.RepoOwner, fOpt.RepoName, err.Error())
				}
			}
		})
	})*/

	Context("happy path updates for each formula type", func() {
		It("can update homebrew-tap style formula", func() {
			src := []byte(`class Glooctl < Formula
  desc "Envoy-Powered API Gateway"
  homepage "https://gloo.solo.io"
  version "0.13.14"

  if OS.mac?
    url "https://github.com/solo-io/gloo/releases/download/v#{version}/glooctl-darwin-amd64"
    sha256 "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99"
  elsif OS.linux?
    url "https://github.com/solo-io/gloo/releases/download/v#{version}/glooctl-linux-amd64"
    sha256 "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f"
  end

  def install
    if OS.mac?
      File.rename "#{name}-darwin-amd64", name
    elsif OS.linux?
      File.rename "#{name}-linux-amd64", name
    end

    bin.install name

    # Install bash completion
    output = Utils.popen_read("SHELL=bash #{bin}/#{name} completion bash")
    (bash_completion/name.to_s).write output

    # Install zsh completion
    output = Utils.popen_read("SHELL=zsh #{bin}/#{name} completion zsh")
    (zsh_completion/"_#{name}").write output

    prefix.install_metafiles
  end
end`)
			fopt := FormulaOptions{
				VersionRegex:   `version\s*"([0-9.]+)"`,
				DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
				LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,
			}

			shas := sha256Outputs{
				darwinSha:  []byte("1234"),
				linuxSha:   []byte("9876"),
				windowsSha: []byte("abcd"),
			}

			byt, err := updateFormula(src, "0.13.15", "12345", &shas, &fopt)
			Expect(err).To(BeNil())
			Expect(byt).To(Equal([]byte(`class Glooctl < Formula
  desc "Envoy-Powered API Gateway"
  homepage "https://gloo.solo.io"
  version "0.13.15"

  if OS.mac?
    url "https://github.com/solo-io/gloo/releases/download/v#{version}/glooctl-darwin-amd64"
    sha256 "1234"
  elsif OS.linux?
    url "https://github.com/solo-io/gloo/releases/download/v#{version}/glooctl-linux-amd64"
    sha256 "9876"
  end

  def install
    if OS.mac?
      File.rename "#{name}-darwin-amd64", name
    elsif OS.linux?
      File.rename "#{name}-linux-amd64", name
    end

    bin.install name

    # Install bash completion
    output = Utils.popen_read("SHELL=bash #{bin}/#{name} completion bash")
    (bash_completion/name.to_s).write output

    # Install zsh completion
    output = Utils.popen_read("SHELL=zsh #{bin}/#{name} completion zsh")
    (zsh_completion/"_#{name}").write output

    prefix.install_metafiles
  end
end`)))
		})

		It("can update homebrew-core style formula", func() {
			src := []byte(`class Glooctl < Formula
  desc "Envoy-Powered API Gateway"
  homepage "https://gloo.solo.io"
  url "https://github.com/solo-io/gloo.git",
      :tag      => "v0.13.14",
      :revision => "f2fdf602bd6c96b0ebbc0d75bc8ac48c66f83db9"
  head "https://github.com/solo-io/gloo.git"

  bottle do
    cellar :any_skip_relocation
    sha256 "6f80dbdac6c90608efd4fb5d71f26c237ef6c4081608020d954f7e06cdc265f3" => :mojave
    sha256 "f05f042aff4765a5396c7c013b3798e6a3ab80292db92958e9376af3d9c40212" => :high_sierra
    sha256 "6b7628ee3ce9e4cf1fb1750feba5db3082857ff74a0798fffa9255490d318ad0" => :sierra
  end

  depends_on "dep" => :build
  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    dir = buildpath/"src/github.com/solo-io/gloo"
    dir.install buildpath.children - [buildpath/".brew_home"]

    cd dir do
      system "dep", "ensure", "-vendor-only"
      system "make", "glooctl", "TAGGED_VERSION=v#{version}"
      bin.install "_output/glooctl"
    end
  end

  test do
    run_output = shell_output("#{bin}/glooctl 2>&1")
    assert_match "glooctl is the unified CLI for Gloo.", run_output

    version_output = shell_output("#{bin}/glooctl --version 2>&1")
    assert_match "glooctl community edition version #{version}", version_output

    # Should error out as it needs access to a Kubernetes cluster to operate correctly
    status_output = shell_output("#{bin}/glooctl get proxy 2>&1", 1)
    assert_match "failed to create proxy client", status_output
  end
end
`)
			fopt := FormulaOptions{
				VersionRegex:    `:tag\s*=>\s*"v([0-9.]+)",`,
				VersionShaRegex: `:revision\s*=>\s*"(.*)"`,
			}

			shas := sha256Outputs{
				darwinSha:  []byte("1234"),
				linuxSha:   []byte("9876"),
				windowsSha: []byte("abcd"),
			}

			byt, err := updateFormula(src, "0.13.15", "12345", &shas, &fopt)
			Expect(err).To(BeNil())
			Expect(byt).To(Equal([]byte(`class Glooctl < Formula
  desc "Envoy-Powered API Gateway"
  homepage "https://gloo.solo.io"
  url "https://github.com/solo-io/gloo.git",
      :tag      => "v0.13.15",
      :revision => "12345"
  head "https://github.com/solo-io/gloo.git"

  bottle do
    cellar :any_skip_relocation
    sha256 "6f80dbdac6c90608efd4fb5d71f26c237ef6c4081608020d954f7e06cdc265f3" => :mojave
    sha256 "f05f042aff4765a5396c7c013b3798e6a3ab80292db92958e9376af3d9c40212" => :high_sierra
    sha256 "6b7628ee3ce9e4cf1fb1750feba5db3082857ff74a0798fffa9255490d318ad0" => :sierra
  end

  depends_on "dep" => :build
  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    dir = buildpath/"src/github.com/solo-io/gloo"
    dir.install buildpath.children - [buildpath/".brew_home"]

    cd dir do
      system "dep", "ensure", "-vendor-only"
      system "make", "glooctl", "TAGGED_VERSION=v#{version}"
      bin.install "_output/glooctl"
    end
  end

  test do
    run_output = shell_output("#{bin}/glooctl 2>&1")
    assert_match "glooctl is the unified CLI for Gloo.", run_output

    version_output = shell_output("#{bin}/glooctl --version 2>&1")
    assert_match "glooctl community edition version #{version}", version_output

    # Should error out as it needs access to a Kubernetes cluster to operate correctly
    status_output = shell_output("#{bin}/glooctl get proxy 2>&1", 1)
    assert_match "failed to create proxy client", status_output
  end
end
`)))
		})

		It("can update fish-food style formula", func() {
			src := []byte(`local name = "glooctl"
local version = "0.12.0"

food = {
    name = name,
    description = "Envoy-Powered API Gateway",
    homepage = "https://gloo.solo.io",
    version = version,
    packages = {
        {
            os = "darwin",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-darwin-amd64",
            -- shasum of the release archive
            sha256 = "8c5492d6f5c4990a5e3ca3f8cc32302d4ff29087190dc2f8df38b4a8798c154d",
            resources = {
                {
                    path = name .. "-darwin-amd64",
                    installpath = "bin/" .. name,
                    executable = true
                }
            }
        },
        {
            os = "linux",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-linux-amd64",
            -- shasum of the release archive
            sha256 = "07529777eccc404dc16ecdba4735d7fc36acd742d6fb88787c171bfdd8bdb761",
            resources = {
                {
                    path = name .. "-linux-amd64",
                    installpath = "bin/" .. name,
                    executable = true
                }
            }
        },
        {
            os = "windows",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-windows-amd64.exe",
            -- shasum of the release archive
            sha256 = "8ba4029b9811df1e1af6f907871efd753bb425d62bece677757c406e076d34ee",
            resources = {
                {
                    path = name .. "-windows-amd64" .. ".exe",
                    installpath = "bin\\" .. name .. ".exe"
                }
            }
        }
    }
}`)
			fopt := FormulaOptions{
				VersionRegex:    `version\s*=\s*"([0-9.]+)"`,
				DarwinShaRegex:  `os\s*=\s*"darwin",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
				LinuxShaRegex:   `os\s*=\s*"linux",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
				WindowsShaRegex: `os\s*=\s*"windows",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
			}

			shas := sha256Outputs{
				darwinSha:  []byte("1234"),
				linuxSha:   []byte("9876"),
				windowsSha: []byte("abcd"),
			}

			byt, err := updateFormula(src, "0.13.15", "12345", &shas, &fopt)
			Expect(err).To(BeNil())
			Expect(byt).To(Equal([]byte(`local name = "glooctl"
local version = "0.13.15"

food = {
    name = name,
    description = "Envoy-Powered API Gateway",
    homepage = "https://gloo.solo.io",
    version = version,
    packages = {
        {
            os = "darwin",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-darwin-amd64",
            -- shasum of the release archive
            sha256 = "1234",
            resources = {
                {
                    path = name .. "-darwin-amd64",
                    installpath = "bin/" .. name,
                    executable = true
                }
            }
        },
        {
            os = "linux",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-linux-amd64",
            -- shasum of the release archive
            sha256 = "9876",
            resources = {
                {
                    path = name .. "-linux-amd64",
                    installpath = "bin/" .. name,
                    executable = true
                }
            }
        },
        {
            os = "windows",
            arch = "amd64",
            url = "https://github.com/solo-io/gloo/releases/download/v" .. version .. "/" .. name .. "-windows-amd64.exe",
            -- shasum of the release archive
            sha256 = "abcd",
            resources = {
                {
                    path = name .. "-windows-amd64" .. ".exe",
                    installpath = "bin\\" .. name .. ".exe"
                }
            }
        }
    }
}`)))
		})
	})

	When("formula already updated", func() {
		It("can detect duplicate update of homebrew-core style formula", func() {
			src := []byte(`class Glooctl < Formula
  desc "Envoy-Powered API Gateway"
  homepage "https://gloo.solo.io"
  url "https://github.com/solo-io/gloo.git",
      :tag      => "v0.13.14",
      :revision => "f2fdf602bd6c96b0ebbc0d75bc8ac48c66f83db9"
  head "https://github.com/solo-io/gloo.git"

  bottle do
    cellar :any_skip_relocation
    sha256 "6f80dbdac6c90608efd4fb5d71f26c237ef6c4081608020d954f7e06cdc265f3" => :mojave
    sha256 "f05f042aff4765a5396c7c013b3798e6a3ab80292db92958e9376af3d9c40212" => :high_sierra
    sha256 "6b7628ee3ce9e4cf1fb1750feba5db3082857ff74a0798fffa9255490d318ad0" => :sierra
  end

  depends_on "dep" => :build
  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    dir = buildpath/"src/github.com/solo-io/gloo"
    dir.install buildpath.children - [buildpath/".brew_home"]

    cd dir do
      system "dep", "ensure", "-vendor-only"
      system "make", "glooctl", "TAGGED_VERSION=v#{version}"
      bin.install "_output/glooctl"
    end
  end

  test do
    run_output = shell_output("#{bin}/glooctl 2>&1")
    assert_match "glooctl is the unified CLI for Gloo.", run_output

    version_output = shell_output("#{bin}/glooctl --version 2>&1")
    assert_match "glooctl community edition version #{version}", version_output

    # Should error out as it needs access to a Kubernetes cluster to operate correctly
    status_output = shell_output("#{bin}/glooctl get proxy 2>&1", 1)
    assert_match "failed to create proxy client", status_output
  end
end
`)
			fopt := FormulaOptions{
				VersionRegex:    `:tag\s*=>\s*"v([0-9.]+)",`,
				VersionShaRegex: `:revision\s*=>\s*"(.*)"`,
			}

			shas := sha256Outputs{
				darwinSha:  []byte("1234"),
				linuxSha:   []byte("9876"),
				windowsSha: []byte("abcd"),
			}

			byt, err := updateFormula(src, "0.13.14", "f2fdf602bd6c96b0ebbc0d75bc8ac48c66f83db9", &shas, &fopt)
			Expect(err).To(Equal(ErrAlreadyUpdated))
			Expect(byt).To(Equal(src))
		})
	})
})
