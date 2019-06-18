package gcloudutils

import (
	"fmt"
	"strings"
)

type Tags []string

const (
	seperator = "_"
	tagConst  = "tag"
	shaConst  = "sha"
	refConst  = "ref"
	repoConst = "repo"
)

func InitializeTags(input []string) Tags {
	if input == nil {
		return make(Tags, 0)
	}
	return input
}

func createTag(title, item string) string {
	return fmt.Sprintf("%s%s%s", title, seperator, item)
}

func (t Tags) AddReleaseTag(tag string) Tags {
	return append(t, createTag(tagConst, tag))
}

func (t Tags) AddShaTag(sha string) Tags {
	return append(t, createTag(shaConst, sha))
}

func (t Tags) AddRepoTag(repo string) Tags {
	return append(t, createTag(repoConst, repo))
}

func (t Tags) AddRefTag(ref string) Tags {
	return append(t, createTag(refConst, ref))
}

func (t Tags) IsReleaseBuild() bool {
	return t.GetReleaseTag() != ""
}

// returns "" for empty
func (t Tags) GetReleaseTag() string {
	return t.getByConst(tagConst)
}

// returns "" for empty
func (t Tags) GetSha() string {
	return t.getByConst(shaConst)
}

// returns "" for empty
func (t Tags) GetRef() string {
	return t.getByConst(refConst)
}

// returns "" for empty
func (t Tags) GetRepo() string {
	return t.getByConst(repoConst)
}

func (t Tags) getByConst(item string) string {
	for _, v := range t {
		splitVal := strings.SplitN(v, seperator, 2)
		if len(splitVal) == 2 && splitVal[0] == item {
			return splitVal[1]
		}
	}
	return ""
}
