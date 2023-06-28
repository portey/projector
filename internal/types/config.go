package types

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var ProjectNotFound = errors.New("project not found")

type Config struct {
	HomeDir     string       `yaml:"-"`
	DebugMode   bool         `yaml:"debug-mode"`
	RawProjects []RawProject `yaml:"projects"`
	SSHAuth     SSHAuth      `yaml:"ssh-auth"`
}

type SSHAuth struct {
	User     string `yaml:"user"`
	PemFile  string `yaml:"pem-file"`
	Password string `yaml:"password"`
}

func (c Config) Auth() (transport.AuthMethod, error) {
	a, err := ssh.NewPublicKeysFromFile(c.SSHAuth.User, c.absolutePath(c.SSHAuth.PemFile), c.SSHAuth.Password)
	if err != nil {
		return nil, err
	}

	return a, nil
}

type RawProject struct {
	Name  string `yaml:"name"`
	Short string `yaml:"short"`
	Path  string `yaml:"path"`

	// Deploy parameters, not mandatory, defaults are in deployer
	RunFileLocation  string `yaml:"run-file-location"`
	QADeployRunFile  string `yaml:"qa-deploy-run-file"`
	UATDeployRunFile string `yaml:"uat-deploy-run-file"`
}

type Project RawProject

func (c Config) Project(name string) (Project, error) {
	for _, project := range c.Projects() {
		if project.Name == name || project.Short == name {
			return project, nil
		}
	}

	return Project{}, ProjectNotFound
}

func (c Config) Projects() []Project {
	res := make([]Project, 0, len(c.RawProjects))
	for _, rp := range c.RawProjects {
		r := Project(rp)
		r.Path = c.absolutePath(r.Path)

		res = append(res, r)
	}

	return res
}

func (c Config) absolutePath(p string) string {
	if p == "~" {
		return c.HomeDir
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(c.HomeDir, p[2:])
	}

	return p
}
