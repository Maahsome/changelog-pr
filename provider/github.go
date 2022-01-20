package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"

	// . "github.com/go-git/go-git/v5/_examples"
	"changelog-pr/common"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-resty/resty/v2"
)

// Github - Structure to hold stuff
type Github struct {
	Provider string
	Host     string
}

type PRBody struct {
	Body  string `json:"body"`
	Links struct {
		HTML struct {
			HREF string `json:"href"`
		} `json:"html"`
	} `json:"_links"`
}

func parsePRNumber(msg string) (uint, error) {
	matches := numRegex.FindAllStringSubmatch(msg, 1)
	if len(matches) == 0 || len(matches[0]) < 2 {
		return 0, fmt.Errorf("could not find PR number in commit message")
	}
	u64, err := strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PR number %q from commit message: %v", matches[0][1], err)
	}
	return uint(u64), nil
}

func getUserRepository(url string) (string, string, error) {
	// git@github.com:Maahsome/changelog-pr.git
	// https://github.com/Maahsome/changelog-pr.git

	var matches [][]string

	var gitRegexPattern string
	var httpRegexPattern string

	if strings.HasSuffix(url, ".git") {
		gitRegexPattern = `:(.+)\/(.+).git`
		httpRegexPattern = `\.com\/(.+)\/(.+).git`
	} else {
		gitRegexPattern = `:(.+)\/(.+)$`
		httpRegexPattern = `\.com\/(.+)\/(.+)$`
	}

	if strings.Contains(url, "git@") {
		var gitRegex = regexp.MustCompile(gitRegexPattern)
		matches = gitRegex.FindAllStringSubmatch(url, -1)
		if len(matches) == 0 || len(matches[0]) == 0 || len(matches[0][1]) == 0 || len(matches[0][2]) == 0 {
			return "", "", errors.New("failed to extract git user/repository")
		}
	}
	if strings.Contains(url, "https://") {
		var httpsRegex = regexp.MustCompile(httpRegexPattern)
		matches = httpsRegex.FindAllStringSubmatch(url, -1)
		if len(matches) == 0 || len(matches[0]) == 0 || len(matches[0][1]) == 0 || len(matches[0][2]) == 0 {
			return "", "", errors.New("failed to extract git user/repository")
		}
	}

	return matches[0][1], matches[0][2], nil
}

// GetChangeLogSincePR - Get the changelog details from the PR/MR description
func (p *Github) GetChangeLogFromPRMR(src string, sincePR string, release string, auth AuthToken, fileName string) (string, error) {

	var (
		resp    *resty.Response
		resperr error
	)

	r, err := git.PlainOpen(src)
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	head, rerr := r.Head()
	if rerr != nil {
		return "", errors.New("failed generation of changelog")
	}
	common.Logger.Info(fmt.Sprintf("HEAD: %s", head.Name().Short()))

	c, cerr := r.Config()
	if cerr != nil {
		return "", errors.New("failed generation of changelog")
	}
	common.Logger.Debug(fmt.Sprintf("Remote URL: %s", c.Remotes["origin"].URLs[0]))
	user, repo, rerr := getUserRepository(c.Remotes["origin"].URLs[0])
	if rerr != nil {
		return "", errors.New("failed generation of changelog")
	}
	common.Logger.Info(fmt.Sprintf("User/Org: %s, Repo: %s\n", user, repo))

	tagrefs, err := r.Tags()
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		common.Logger.Debug(fmt.Sprintf("Tag Name: %s", t.Name().String()))
		var (
			nv semver.Version
			lv semver.Version
		)

		if len(sincePR) > 0 {
			if strings.HasSuffix(t.Name().String(), sincePR) {
				common.Logger.Debug(t.Hash())
				lastTag = t
			}
		} else {
			//refs/tags/v0.1.0
			var refRegex = regexp.MustCompile(`\/tags\/v(.+)`)
			newMatches := refRegex.FindAllStringSubmatch(t.Name().String(), -1)
			if len(newMatches) > 0 && len(newMatches[0]) > 0 && len(newMatches[0][1]) > 0 {
				nv, err = semver.Parse(newMatches[0][1])
				if err != nil {
					common.Logger.Error(fmt.Sprintf("Error parsing SemVer for %s", newMatches[0][1]))
					return nil
				}
			}
			if lastTag != nil {
				lastMatches := refRegex.FindAllStringSubmatch(lastTag.Name().String(), -1)
				if len(lastMatches) > 0 && len(lastMatches[0]) > 0 && len(lastMatches[0][1]) > 0 {
					lv, err = semver.Parse(lastMatches[0][1])
					if err != nil {
						common.Logger.Error(fmt.Sprintf("Error parsing SemVer for %s", lastMatches[0][1]))
						return nil
					}
					if nv.GTE(lv) {
						lastTag = t
					}
				}
			} else {
				lastTag = t
			}
		}
		return nil
	})
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}
	common.Logger.Info(fmt.Sprintf("Last Tag/Hash: %s (%s)", lastTag.Name().String(), lastTag.Hash()))

	// Gets the HEAD history from HEAD, just like this command:
	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	hasHash := false
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Hash == lastTag.Hash() {
			hasHash = true
		}
		return nil
	})
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}
	PRs := []string{}
	if hasHash {
		findingHash := true
		cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
		if err != nil {
			return "", errors.New("failed generation of changelog")
		}
		err = cIter.ForEach(func(c *object.Commit) error {
			if c.Hash == lastTag.Hash() {
				findingHash = false
				return nil
			}
			if findingHash {
				if strings.HasPrefix(c.Message, "Merge pull request #") {
					pr, err := parsePRNumber(strings.Split(c.Message, "\n")[0])
					if err != nil {
						common.Logger.WithError(err).Error("Bad PR Parse")
					}
					PRs = append(PRs, fmt.Sprintf("%d", pr))
					common.Logger.Info(fmt.Sprintf("%s %s\n", c.ID(), strings.Split(c.Message, "\n")[0]))
				}
				return nil
			}
			return nil
		})
		if err != nil {
			return "", errors.New("failed generation of changelog")
		}
	} else {
		common.Logger.Fatal(fmt.Sprintf("The TAG you specified was NOT in the currently selected BRANCH: %s", head.Name().Short()))
	}
	// curl -sH "Accept: application/vnd.github.v3+json" https://api.github.com/repos/splicemachine/splicectl/pulls/5 | jq -r '.body'
	restClient := resty.New()

	changeLog := common.Changelog{}
	changeLog.Version = release

	for _, p := range PRs {
		uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s", user, repo, p)
		common.Logger.Debug(fmt.Sprintf("PR URI: %s", uri))
		if len(auth.AccessToken) > 0 {
			resp, resperr = restClient.R().
				SetHeader("Accept", "application/vnd.github.v3+json").
				SetHeader("Authorization", fmt.Sprintf("token %s", auth.AccessToken)).
				Get(uri)
		} else {
			resp, resperr = restClient.R().
				SetHeader("Accept", "application/vnd.github.v3+json").
				Get(uri)
		}
		if resperr != nil {
			common.Logger.WithError(resperr).Error("Error getting PR")
		}

		var body PRBody

		marshErr := json.Unmarshal(resp.Body(), &body)
		if marshErr != nil {
			common.Logger.Error("Could not unmarshall data", marshErr)
		}

		err := common.ParseMarkdown(body.Body, p, &changeLog, "Pull Request", body.Links.HTML.HREF)
		if err != nil {
			common.Logger.Error("Could not parse the markdown")
		}
	}

	markdown, err := changeLog.Template()
	if err != nil {
		return "", errors.New("failed generation of changelog")
	}

	if len(fileName) > 0 {
		err := changeLog.WriteFile(fileName)
		if err != nil {
			return "", errors.New("failed to write to the output file")
		}
		return "Changelog data has been saved.", nil
	}

	return string(markdown[:]), nil
}
