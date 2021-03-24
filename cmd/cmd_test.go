package cmd_test

import (
	"errors"
	"io/ioutil"
	"os"

	"changelog-pr/common"
	clprovider "changelog-pr/provider"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: Shall we create some error conditions in the Mock and detect them in a test?
var _ = Describe("Generate", func() {

	var (
		mockGitRepo       clprovider.Provider
		err               error
		markdown          string
		markdownAdditions string
		markdownNoClosure string
	)

	AfterSuite(func() {
		if _, err := os.Stat("/tmp/changelog-pr.md"); err == nil {
			rerr := os.Remove("/tmp/changelog-pr.md")
			if rerr != nil {
				Expect(rerr).To(Equal(nil))
			}
		}
	})
	BeforeEach(func() {
		mockGitRepo, err = clprovider.GetProvider(clprovider.MOCK)
		if err != nil {
			Expect(err).To(Equal(nil))
		}
		common.NewLogger("Warn", "")
		if _, err := os.Stat("/tmp/changelog-pr.md"); err == nil {
			rerr := os.Remove("/tmp/changelog-pr.md")
			if rerr != nil {
				Expect(rerr).To(Equal(nil))
			}
		}

		markdown = `## v0.2.0

### Additions

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Addition 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Addition 2


### Changes

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Change 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Change 2


### Removals

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Removed 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Removed 2


### Deprecations

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Deprecate 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Deprecate 2


### Bug Fixes

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Fix 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Fix 2


### Breaking Changes

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Breaking 1

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Breaking 2

`

		markdownAdditions = `## v0.2.0

### Additions

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Addition 1
- Addition 2

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Addition 3
- Addition 4

`

		markdownNoClosure = `## v0.2.0

### Changes

#### [Pull Request #0](https://github.com/splicemachine/splicectl/pull/0)

- Change 1
- Change 2

#### [Pull Request #1](https://github.com/splicemachine/splicectl/pull/1)

- Change 3
- Change 4

`

	})

	Describe("New Changelog", func() {

		It("create a new changelog, all sections", func() {
			auth := clprovider.AuthToken{
				GithubToken: "abcdefghijklmnop",
			}
			out, err := mockGitRepo.GetChangeLogFromPR("", "v0.0.1", "v0.2.0", auth, "")
			if err != nil {
				Expect(err).To(Equal(nil))
			}
			Expect(out).To(Equal(markdown))
		})
		It("create a new changelog, additions only", func() {
			auth := clprovider.AuthToken{
				GithubToken: "abcdefghijklmnop",
			}
			out, err := mockGitRepo.GetChangeLogFromPR("", "v0.0.2", "v0.2.0", auth, "")
			if err != nil {
				Expect(err).To(Equal(nil))
			}
			Expect(out).To(Equal(markdownAdditions))
		})

		It("create a new changelog, all sections, to file", func() {
			auth := clprovider.AuthToken{
				GithubToken: "abcdefghijklmnop",
			}
			_, err := mockGitRepo.GetChangeLogFromPR("", "v0.0.1", "v0.2.0", auth, "/tmp/changelog-pr.md")
			if err != nil {
				Expect(err).To(Equal(nil))
			}
			fileData, err := ioutil.ReadFile("/tmp/changelog-pr.md")
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				Expect(err).To(Equal(nil))
			}
			Expect(string(fileData[:])).To(Equal(markdown))
		})

		It("create a new changelog, change section, no closure", func() {
			auth := clprovider.AuthToken{
				GithubToken: "abcdefghijklmnop",
			}
			out, err := mockGitRepo.GetChangeLogFromPR("", "v0.0.3", "v0.2.0", auth, "")
			if err != nil {
				Expect(err).To(Equal(nil))
			}
			Expect(out).To(Equal(markdownNoClosure))
		})

	})

})
