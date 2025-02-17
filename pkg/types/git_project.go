package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	git_url "github.com/kluctl/kluctl/v2/pkg/git/git-url"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
)

// gitDirPatternNeg defines forbiden characters on git directory path/subDir
var gitDirPatternNeg = regexp.MustCompile(`[\\\/:\*?"<>|[:cntrl:]\0^]`)

type GitProject struct {
	Url    git_url.GitUrl `yaml:"url" validate:"required"`
	Ref    string         `yaml:"ref,omitempty"`
	SubDir string         `yaml:"subDir,omitempty"`
}

func (gp *GitProject) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&gp.Url); err == nil {
		// it's a simple string
		return nil
	}
	type raw GitProject
	return unmarshal((*raw)(gp))
}

// invalidDirName evaluate directory name against forbidden characters
func invalidDirName(dirName string) bool {
	return gitDirPatternNeg.MatchString(dirName)
}

// validateGitSubDir evaluate syntax for subdirectory path
func validateGitSubDir(path string) bool {
	for _, dirName := range strings.Split(path, "/") {
		if invalidDirName(dirName) {
			return false
		}
	}
	return true
}

func ValidateGitProject(sl validator.StructLevel) {
	gp := sl.Current().Interface().(GitProject)
	if !validateGitSubDir(gp.SubDir) {
		sl.ReportError(gp.SubDir, "subDir", "SubDir", fmt.Sprintf("'%s' is not valid git subdirectory path", gp.SubDir), "")
	}
}

type ExternalProject struct {
	Project *GitProject `yaml:"project,omitempty"`
	Path    *string     `yaml:"path,omitempty"`
}

func ValidateExternalProject(sl validator.StructLevel) {
	p := sl.Current().Interface().(ExternalProject)
	if p.Project == nil && p.Path == nil {
		sl.ReportError(p, ".", ".", "either project or path must be set", "")
	} else if p.Project != nil && p.Path != nil {
		sl.ReportError(p, ".", ".", "only one of project or path can be set", "")
	}
}

func init() {
	yaml.Validator.RegisterStructValidation(ValidateGitProject, GitProject{})
	yaml.Validator.RegisterStructValidation(ValidateExternalProject, ExternalProject{})
}
