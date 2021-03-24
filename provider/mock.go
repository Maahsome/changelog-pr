package provider

import (
	"errors"
	"fmt"

	"changelog-pr/common"
)

// Mock - Structure to hold stuff
type Mock struct {
	Provider string
}

// GetChangeLogSincePR - Get the changelog details from the PR description
func (p *Mock) GetChangeLogFromPR(src string, sincePR string, release string, auth AuthToken, fileName string) (string, error) {

	var (
		PRData []string
	)

	changeLog := common.Changelog{}
	changeLog.Version = release

	switch sincePR {
	case "v0.0.1":
		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Additions

- Addition 1

### Changes

- Change 1

### Fixes

- Fix 1

### Deprecated

- Deprecate 1

### Removed

- Removed 1

### Breaking Changes

- Breaking 1

## Checklist
`)

		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Additions

- Addition 2

### Changes

- Change 2

### Fixes

- Fix 2

### Deprecated

- Deprecate 2

### Removed

- Removed 2

### Breaking Changes

- Breaking 2

## Checklist
`)
	case "v0.0.2":
		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Additions

- Addition 1
- Addition 2

## Checklist
`)

		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Additions

- Addition 3
- Addition 4

## Checklist
`)
	case "v0.0.3":
		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Changes

- Change 1
- Change 2
`)

		PRData = append(PRData, `## Description

This is the description

## Changelog Inclusions

### Changes

- Change 3
- Change 4
`)
	}

	for k, v := range PRData {
		err := common.ParseMarkdown(v, fmt.Sprintf("%d", k), &changeLog)
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
