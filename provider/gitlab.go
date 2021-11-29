package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-resty/resty/v2"

	// . "github.com/go-git/go-git/v5/_examples"
	"changelog-pr/common"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Github - Structure to hold stuff
type Gitlab struct {
	Provider string
	Host     string
}

type MRDescription struct {
	Description string `json:"description"`
	WebURL      string `json:"web_url"`
}

func parseMRNumber(msg string) (uint, error) {
	matches := numBangRegex.FindAllStringSubmatch(msg, 1)
	if len(matches) == 0 || len(matches[0]) < 2 {
		return 0, fmt.Errorf("could not find MR number in commit message")
	}
	u64, err := strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse MR number %q from commit message: %v", matches[0][1], err)
	}
	return uint(u64), nil
}

// func getUserRepository(url string) (string, string, error) {
// 	// git@github.com:Maahsome/changelog-pr.git
// 	// https://github.com/Maahsome/changelog-pr.git

// 	var matches [][]string

// 	if strings.Contains(url, "git@") {
// 		var gitRegex = regexp.MustCompile(`:(.+)\/(.+).git`)
// 		matches = gitRegex.FindAllStringSubmatch(url, -1)
// 		if len(matches) == 0 || len(matches[0]) == 0 || len(matches[0][1]) == 0 || len(matches[0][2]) == 0 {
// 			return "", "", errors.New("failed to extract git user/repository")
// 		}
// 	}
// 	if strings.Contains(url, "https://") {
// 		var httpsRegex = regexp.MustCompile(`\.com\/(.+)\/(.+).git`)
// 		matches = httpsRegex.FindAllStringSubmatch(url, -1)
// 		if len(matches) == 0 || len(matches[0]) == 0 || len(matches[0][1]) == 0 || len(matches[0][2]) == 0 {
// 			return "", "", errors.New("failed to extract git user/repository")
// 		}
// 	}

// 	return matches[0][1], matches[0][2], nil
// }

// GetChangeLogSincePR - Get the changelog details from the PR/MR description
func (p *Gitlab) GetChangeLogFromPRMR(src string, sincePR string, release string, auth AuthToken, fileName string) (string, error) {

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
	MRs := []string{}
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
				common.Logger.Trace(c.Message)
				if strings.Contains(c.Message, "See merge request") {
					messageLines := strings.Split(c.Message, "\n")
					lastLine := messageLines[len(messageLines)-1]
					common.Logger.Info("lastLine", lastLine)
					mr, err := parseMRNumber(lastLine)
					if err != nil {
						common.Logger.WithError(err).Error("Bad MR Parse")
					}
					MRs = append(MRs, fmt.Sprintf("%d", mr))
					common.Logger.Info(fmt.Sprintf("%s %s\n", c.ID(), lastLine))
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
	glSlug := url.PathEscape(fmt.Sprintf("%s/%s", user, repo))
	for _, m := range MRs {
		uri := fmt.Sprintf("https://%s/api/v4/projects/%s/merge_requests/%s", p.Host, glSlug, m)
		common.Logger.Debug(fmt.Sprintf("PR URI: %s", uri))
		if len(auth.AccessToken) > 0 {
			resp, resperr = restClient.R().
				SetHeader("Accept", "application/json").
				SetHeader("PRIVATE-TOKEN", auth.AccessToken).
				Get(uri)
		} else {
			resp, resperr = restClient.R().
				SetHeader("Accept", "application/json").
				Get(uri)
		}
		if resperr != nil {
			common.Logger.WithError(resperr).Error("Error getting MR")
		}

		common.Logger.Trace("MR Response", string(resp.Body()[:]))

		var description MRDescription

		marshErr := json.Unmarshal(resp.Body(), &description)
		if marshErr != nil {
			common.Logger.Error("Could not unmarshall data", marshErr)
		}

		err := common.ParseMarkdown(description.Description, m, &changeLog, "Merge Request", description.WebURL)
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
