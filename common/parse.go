package common

import (
	"fmt"
	"strings"
)

func collectSectionText(cl *Changelog, sectionName string, sectionText string, pr string, requestText string, requestURL string) {

	switch sectionName {
	case "## Changelog Inclusions.### Additions":
		cl.Additions = append(cl.Additions, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})
	case "## Changelog Inclusions.### Changes":
		cl.Changes = append(cl.Changes, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})
	case "## Changelog Inclusions.### Fixes":
		cl.Bugfixes = append(cl.Bugfixes, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})
	case "## Changelog Inclusions.### Deprecated":
		cl.Deprecations = append(cl.Deprecations, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})
	case "## Changelog Inclusions.### Removed":
		cl.Removals = append(cl.Removals, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})
	case "## Changelog Inclusions.### Breaking Changes":
		cl.Breaking = append(cl.Breaking, ChangelogEntry{
			Description: sectionText,
			Link:        fmt.Sprintf("[%s #%s](%s)", requestText, pr, requestURL),
		})

	}

}

func ParseMarkdown(body string, pr string, cl *Changelog, requestText string, requestURL string) error {

	var splits []string
	sections := map[string]string{}
	depthNames := map[int]string{
		0: "none",
		1: "none",
		2: "none",
		3: "none",
	}
	sectionName := ""
	sectionText := ""
	currentDepth := 0
	// Process to extract the data between the ### markdown sections after ## Changelog Inclusions
	if strings.Contains(body, "\r") {
		splits = strings.Split(body, "\r")
	} else {
		splits = strings.Split(body, "\n")
	}
	Logger.Info(fmt.Sprintf("Searching PR#%s for Changelog Inclusions...", pr))
	Logger.Trace(fmt.Sprintf("Body: %s", body))
	for _, v := range splits {
		Logger.Trace(fmt.Sprintf("%s\n", v))
		if strings.HasPrefix(strings.TrimSpace(v), "#") {
			// We are at a markdown section marker, if we have section text, we need to capture it
			if len(sectionName) > 0 && len(sectionText) > 0 {
				collectSectionText(cl, sectionName, sectionText, pr, requestText, requestURL)
				sections[sectionName] = sectionText
				sectionText = ""
			}
			// determine our new section as our line starts with a markdown section marker
			if strings.HasPrefix(strings.TrimSpace(v), "## ") {
				currentDepth = 2
				depthNames[currentDepth] = strings.TrimSpace(v)
				sectionName = strings.TrimSpace(v)
				Logger.Debug(sectionName)
			}
			if strings.HasPrefix(strings.TrimSpace(v), "### ") {
				currentDepth = 3
				depthNames[currentDepth] = strings.TrimSpace(v)
				sectionName = fmt.Sprintf("%s.%s", depthNames[2], strings.TrimSpace(v))
				Logger.Debug(sectionName)
			}
		} else {
			// We are not at a markdown section, collect the section text
			if len(sectionName) > 0 {
				if len(strings.Trim(strings.TrimSpace(v), "\n\r")) > 0 {
					Logger.Debug(fmt.Sprintf("~%s~", strings.Trim(v, "\n\r")))
					sectionText += fmt.Sprintf("%s\n", strings.Trim(v, "\n\r"))
				}
			}
		}
	}
	// If we exit with no enclosing section, where one of our changes sections is last in the description
	// we need to process that last block of text we collected.
	if len(sectionName) > 0 && len(sectionText) > 0 {
		Logger.Info("Collecting text from section", sectionName)
		collectSectionText(cl, sectionName, sectionText, pr, requestText, requestURL)
		sections[sectionName] = sectionText
		sectionText = ""
	}

	return nil
}
