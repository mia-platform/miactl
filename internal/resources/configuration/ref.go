package configuration

import (
	"fmt"
	"net/url"
)

type RefTypes map[string]bool

const (
	RevisionRefType = "revisions"
	VersionRefType  = "versions"
	BranchRefType   = "branches"
	TagRefType      = "tags"
)

var validRefTypes = RefTypes{RevisionRefType: true, VersionRefType: true, BranchRefType: true, TagRefType: true}

type Ref struct {
	refType string
	refName string
}

func NewRef(refType, refName string) (Ref, error) {
	if !validRefTypes[refType] {
		return Ref{}, fmt.Errorf("unknown reference type: %s", refType)
	}
	if len(refName) == 0 {
		return Ref{}, fmt.Errorf("missing reference name, please provide a reference name")
	}
	return Ref{
		refType: refType,
		refName: refName,
	}, nil
}

// EncodedLocationPath returns the encoded path to be used when fetching configuration data
//
// e.g., "<ConsoleURL>/api/projects/<ProjectID>/<EncodedLocationPath()>/configuration"
func (r Ref) EncodedLocationPath() string {
	switch r.refType {
	case RevisionRefType, VersionRefType:
		return fmt.Sprintf("%s/%s", r.refType, url.PathEscape(r.refName))
	case BranchRefType, TagRefType:
		// Legacy projects use /branches endpoint only
		return fmt.Sprintf("branches/%s", url.PathEscape(r.refName))
	default:
		return ""
	}
}
