package cmd

import (
	"errors"
	"fmt"
	"strings"

	"changelog-pr/common"
	"changelog-pr/provider"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a Changelog from PR descriptions since a specified TAG",
	Long: `EXAMPLE:
	In this example we will look at all PRs that have been merged since the repository
	was tagged with 'v0.1.0'

	  %> changelog-pr generate --path <path/to/git/src> --since-tag v0.1.0 --release-tag v0.2.0

EXAMPLE:
	In this example the very last semver based TAG will be used as the '--since-tag'

	  %> changelog-pr generate --path . --release-tag v0.2.1

EXAMPLE:
	In this example we will create a <SEMVER>.md changelog file.  If the release file already exists,
	data will be appended to the already existing file, allowing one to update a master changelog.md
	file

	  %> RELEASE_VERSION="v0.2.1"
	  %> changelog-pr generate --path . --release-tag ${RELEASE_VERSION} --file changelog/${RELEASE_VERSION.md}

EXAMPLE:
	In this example, the repository is private and requires a Personal Access Token to access the PR information.

	  %> # fetch your personal access token from whereever you store your secrets
	  %> GIT_PAT=$(security find-generic-password -l "git_pat" -w scripting.keychain-db)
	  %> changelog-pr generate --path . --release-tag "v0.2.3" --gh-token ${GIT_PAT}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		srcPath, _ := cmd.Flags().GetString("path")
		sinceTag, _ := cmd.Flags().GetString("since-tag")
		releaseTag, _ := cmd.Flags().GetString("release-tag")
		changelogFile, _ := cmd.Flags().GetString("file")

		if len(sinceTag) > 0 {
			_, sterr := semver.Parse(strings.Replace(sinceTag, "v", "", 1))
			if sterr != nil {
				common.Logger.Fatal(fmt.Sprintf("Error parsing SemVer for %s", sinceTag))
			}
		}
		_, rterr := semver.Parse(strings.Replace(releaseTag, "v", "", 1))
		if rterr != nil {
			common.Logger.Fatal(fmt.Sprintf("Error parsing SemVer for %s", releaseTag))
		}

		glog, err := generateLog(srcPath, sinceTag, releaseTag, changelogFile)
		if err != nil {
			common.Logger.WithError(err).Error("Error generating the changelog")
		}

		fmt.Println(glog)

	},
}

func generateLog(src string, sTag string, rTag string, logFile string) (string, error) {

	var (
		err   error
		chlog string
		gp    provider.Provider
	)

	switch strings.ToLower(gitProvider) {
	case "github":
		gp, err = provider.GetProvider(provider.GITHUB)
		if err != nil {
			return "", errors.New("failed to provision git provider")
		}
	default:
		return "", errors.New("unsupported provider")
	}

	auth := provider.AuthToken{
		GithubToken: ghToken,
	}
	chlog, err = gp.GetChangeLogFromPR(src, sTag, rTag, auth, logFile)
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	return chlog, nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("path", "p", "", "Specify the path to the git source directory")
	generateCmd.Flags().StringP("since-tag", "t", "", "Specify the git TAG to go back to and process PR descriptions")
	generateCmd.Flags().StringP("release-tag", "r", "", "Specify the new release TAG")
	generateCmd.Flags().StringP("file", "f", "", "Specify an output file to save the changelog to")
	generateCmd.MarkFlagRequired("path")
	// generateCmd.MarkFlagRequired("since-tag")
	generateCmd.MarkFlagRequired("release-tag")
}
