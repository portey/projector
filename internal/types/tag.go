package types

import (
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
)

type Tags []Tag

var NilTag = Tag{
	Hash: plumbing.Hash{},
	Name: "Not set yet",
}

func (t Tags) LatestInEnv(env Env) Tag {
	semVersions := make([]*version.Version, 0, len(t))
	m := make(map[string]Tag, len(t))
	for _, tt := range t {
		if tt.Version.IsEnv(env) {
			semVersions = append(semVersions, tt.Version.SemVer())
			m[tt.Version.SemVer().String()] = tt
		}
	}

	if len(semVersions) == 0 {
		return NilTag
	}

	sort.Sort(sort.Reverse(version.Collection(semVersions)))

	return m[semVersions[0].String()]
}

type Tag struct {
	Hash    plumbing.Hash
	Name    string
	Version Version
}

func (t Tag) IsNil() bool {
	return t.Name == NilTag.Name
}

type DeployState struct {
	Project Project
	Tags    map[Env]Tag
}
