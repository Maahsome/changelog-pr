package provider

//import errors to log errors when they occur
import (
	"errors"
	"regexp"

	// . "github.com/go-git/go-git/v5/_examples"

	"github.com/go-git/go-git/v5/plumbing"
)

// Provider = The main interface used to describe appliances
type Provider interface {
	GetChangeLogFromPRMR(sourcePath string, sinceTag string, releaseTag string, auth AuthToken, fileName string) (string, error)
}

type AuthToken struct {
	AccessToken string
}

var lastTag *plumbing.Reference
var numRegex = regexp.MustCompile(`#(\d+) from`)
var numBangRegex = regexp.MustCompile(`!(\d+)$`)

// Provider Types
const (
	GITHUB = "github"
	GITLAB = "gitlab"
	MOCK   = "mock"
)

// GetProvider - Function to create the appliances
func GetProvider(t string, h string) (Provider, error) {
	//Use a switch case to switch between types, if a type exist then error is nil (null)
	switch t {
	case GITHUB:
		return &Github{
			Provider: "github",
			Host:     h,
		}, nil
	case GITLAB:
		return &Gitlab{
			Provider: "gitlab",
			Host:     h,
		}, nil
	case MOCK:
		return new(Mock), nil
	default:
		//if type is invalid, return an error
		return nil, errors.New("unsupported provider")
	}
}
