package gcloudutils_test

import (
	"encoding/json"

	"github.com/solo-io/go-utils/gcloudutils"

	"google.golang.org/api/cloudbuild/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	bfs  cloudbuild.Build
	bfr  cloudbuild.Build
	bfrt cloudbuild.Build
)

const (
	repoName = "repo-name"
	tagName  = "v0.1.0"
)

var _ = BeforeSuite(func() {
	var err error
	err = json.Unmarshal([]byte(buildFromSource), &bfs)
	Expect(err).NotTo(HaveOccurred())
	err = json.Unmarshal([]byte(buildFromRepo), &bfr)
	Expect(err).NotTo(HaveOccurred())
	err = json.Unmarshal([]byte(buildFromRepoTag), &bfrt)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("github test", func() {

	Context("repo name", func() {
		It("can find the repo name from repo source", func() {
			Expect(gcloudutils.GetRepoName(&bfr)).To(Equal(repoName))
		})
	})

	Context("tag name", func() {

		It("can find tag name from source", func() {
			tag := gcloudutils.GetReleaseVersionTag(&bfrt)
			Expect(tag).To(Equal(tagName))
		})

		It("returns empty string if no tag exists", func() {
			tag := gcloudutils.GetReleaseVersionTag(&bfr)
			Expect(tag).To(Equal(""))
		})
	})

	Context("tags", func() {
		It("can find release tag from tags", func() {
			tag := gcloudutils.GetReleaseVersionTag(&bfs)
			Expect(tag).To(Equal(tagName))
		})

		It("can find sha from tags", func() {
			tag := gcloudutils.GetTargetCommitSh(&bfs)
			Expect(tag).To(Equal("world"))
		})

		It("can find repo name from tags", func() {
			tag := gcloudutils.GetRepoName(&bfs)
			Expect(tag).To(Equal(repoName))

		})
	})

})

const buildFromSource = `
    {
      "id": "6c114a53-872a-4b3b-aabb-ea547c9e9406",
      "status": "SUCCESS",
      "source": {
        "storageSource": {
          "bucket": "project-name_cloudbuild",
          "object": "foo.tgz",
          "generation": "1550549390047482"
        }
      },
      "tags": [
      	"ref_hello",
      	"tag_v0.1.0",
      	"sha_world",
      	"repo_repo-name"
      ],
      "createTime": "2019-02-19T04:09:50.309127374Z",
      "startTime": "2019-02-19T04:09:51.269815120Z",
      "finishTime": "2019-02-19T04:13:10.228509Z",
      "results": {},
      "steps": [],
      "timeout": "600s",
      "projectId": "project-name",
      "logsBucket": "",
      "sourceProvenance": {
        "resolvedStorageSource": {
          "bucket": "project-name_cloudbuild",
          "object": "",
          "generation": "1550549390047482"
        },
        "fileHashes": {
          "": {}
        }
      },
      "options": {
        "logging": "LEGACY"
      },
      "substitutions": {
        "TAG_NAME": "v0.1.0"
      },
      "logUrl": ""
    }
`

const buildFromRepo = `
    {
      "id": "27ca4d2d-312e-4794-a6c2-5f7aeb2166fc",
      "status": "SUCCESS",
      "source": {
        "repoSource": {
          "projectId": "project-name",
          "repoName": "github_solo-io_repo-name",
          "branchName": "master"
        }
      },
      "createTime": "2019-02-18T18:49:56.911431333Z",
      "startTime": "2019-02-18T18:49:57.730496515Z",
      "finishTime": "2019-02-18T18:55:41.086746Z",
      "results": {},
      "steps": [],
      "timeout": "600s",
      "projectId": "project-name",
      "logsBucket": "",
      "sourceProvenance": {
        "resolvedRepoSource": {
          "projectId": "project-name",
          "repoName": "github_solo-io_repo-name",
          "commitSha": "c345645ddbee9c6f930b7d930631956a3553df09"
        }
      },
      "buildTriggerId": "9e1ab0e5-1cc4-4857-9635-a819565c7c5b",
      "options": {
        "substitutionOption": "ALLOW_LOOSE",
        "logging": "LEGACY"
      },
      "logUrl": "",
      "tags": [
        "event-4dcba763-5140-4250-a580-4f75cc83454a",
        "trigger-9e1ab0e5-1cc4-4857-9635-a819565c7c5b"
      ]
    }
`

const buildFromRepoTag = `
{
  "id": "64aba1ff-28fb-450a-9549-a78fbbd4fe8d",
  "status": "SUCCESS",
  "source": {
    "repoSource": {
      "projectId": "project-name",
      "repoName": "github_solo-io_repo-name",
      "tagName": "v0.1.0"
    }
  },
  "createTime": "2019-03-06T03:47:39.517923039Z",
  "startTime": "2019-03-06T03:47:40.342297739Z",
  "finishTime": "2019-03-06T03:57:53.133222Z",

  "timeout": "900s",
  "images": [
    "gcr.io/project-name/repo-name"
  ],
  "projectId": "project-name",
  "buildTriggerId": "726293f4-2716-4906-91c7-d8aa48d4e6a5",
  "options": {
    "substitutionOption": "ALLOW_LOOSE",
    "logging": "LEGACY"
  },
  "logUrl": "",
  "tags": [
    "event-b9c277ee-162e-4ab7-aa45-e3ae65344609",
    "trigger-726293f4-2716-4906-91c7-d8aa48d4e6a5"
  ]
}

`
