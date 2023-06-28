package types

import "github.com/hashicorp/go-version"

type Version struct {
	base   string
	env    Env
	semver *version.Version
}

func VersionFromString(s string) (Version, error) {
	vv, err := version.NewSemver(s)
	if err != nil {
		return Version{}, err
	}

	ee, err := EnvFromSuffix(vv.Prerelease())
	if err != nil {
		return Version{}, err
	}

	return Version{
		base:   vv.Core().String(),
		env:    ee,
		semver: vv,
	}, nil
}

func (v Version) IsEnv(env Env) bool {
	return v.env == env
}

func (v Version) ToEnv(env Env) Version {
	c := v
	c.env = env

	return c
}

func (v Version) Tag() string {
	if v.IsEnv(EnvDEV) {
		return "v" + v.base
	}

	return "v" + v.base + "-" + string(v.env)
}

func (v Version) Equal(s Version) bool {
	return v.env == s.env && v.base == s.base
}

func (v Version) SemVer() *version.Version {
	return v.semver
}
