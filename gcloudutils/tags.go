package gcloudutils

import (
	"fmt"
	"strconv"
	"strings"
)

type Tags []string

const (
	seperator    = "_"
	tagConst     = "tag"
	shaConst     = "sha"
	refConst     = "ref"
	repoConst    = "repo"
	instIdConst  = "instId"
	prConst      = "pr"
	builderConst = "builder"
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

func createIntegerTag(title string, item int64) string {
	return fmt.Sprintf("%s%s%d", title, seperator, item)
}

func (t Tags) AddReleaseTag(tag string) Tags {
	return append(t, createTag(tagConst, tag))
}

func (t Tags) AddBuilderTag(builder string) Tags {
	return append(t, createTag(builderConst, builder))
}

func (t Tags) AddShaTag(sha string) Tags {
	return append(t, createTag(shaConst, sha))
}

func (t Tags) AddRepoTag(repo string) Tags {
	return append(t, createTag(repoConst, repo))
}

func (t Tags) AddInstallationIdTag(instId int64) Tags {
	return append(t, createIntegerTag(instIdConst, instId))
}

func (t Tags) AddRefTag(ref string) Tags {
	return append(t, createTag(refConst, ref))
}

func (t Tags) AddPRTag(pr int) Tags {
	return append(t, createIntegerTag(prConst, int64(pr)))
}

func (t Tags) IsReleaseBuild() bool {
	return t.GetReleaseTag() != ""
}

// returns "" for empty
func (t Tags) GetReleaseTag() string {
	return t.getByConst(tagConst)
}

func (t Tags) GetBuilderTag() string {
	return t.getByConst(builderConst)
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

func (t Tags) GetPR() int {
	stringVal := t.getByConst(prConst)
	prNum, _ := strconv.Atoi(stringVal)
	return prNum
}

func (t Tags) GetInstallationId() int64 {
	stringVal := t.getByConst(instIdConst)
	instId, _ := strconv.ParseInt(stringVal, 10, 64)
	return instId
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
