# Changelog generation

---
These utils allow us to make our [changelogs](https://docs.solo.io/gloo-edge/master/reference/changelog/enterprise/)
rich and easy-to-use for end users. 

**Note**: `changeloggenuitls` are currently only used in the user-facing changelogs on product sites. If you are looking for 
the code that generates changelogs on github releases, that lives under the `changelogutils` package.

## The Problem

We fetch our changelog markdown from github release notes, which are in chronological ordering.
As we often release bug fixes on backported branches, chronological ordering isn't the most intuitive for end users and
often requires a non-trivial amount of scrolling to get to the desired changelog.

In addition, users often have a tough time finding out changes between different versions of our products. A changelog
"diff tool" would be exteremely useful here, but can not be done without additional processing on the 
markdown release notes from github.

## Changelog Gen Utils

These utils parse release notes from github and output a JSON which can be input to a 
javascript front-end, which will be part of the Solo.io Hugo theme used for the docs site.
There are two "levels" of changelogs that are generated. These are described below.



### `minor_release.go`
The minor release changelog generator groups changelogs by minor release. Here is an example
of the structure of changelogs.

```yaml
Opts:
  # These are options passed into the generator which are also in the output JSON
  # as metadata information for the front-end
  ...
ReleaseData:
  - v1.8.0:
    - v1.8.0-beta3:
        ...
    - v1.8.0-beta2:
        ...
  - v1.7.0:
      - v1.7.0-beta9:
       ...
```

The front-end is able to take this data and display it grouped by minor version with very little
processing required.

### `merged_release.go`
Many of our products have an open-source and enterprise component to them, split across multiple repos. 
Enterprise versions rely on open-source versions, and so trying to understand open-source changes between enterprise versions 
can get tricky. This is why we merge in open source changelog notes into the enterprise version notes. 

### Full structure of JSON:

```JSON
{
  "Opts": {
    "NumVersions": 200,
    // "MaxVersion": semver,
    "MinVersion": "v1.0.0",
    "RepoOwner": "solo-io",
    "EnterpriseRepo": "gloo-edge-enterprise",
    "OpenSourceRepo": "gloo-edge"
  },
  "ReleaseData": [
    {
      // Top level release data has major version and minor version for grouping
      "v1.8.0": [
        {
          // Version data includes full version infformation
          "v1.8.0-beta3": {
            "Categories": {
              "Dependency Bumps": [
                {
                  "Note": "bumped k8s dep to v1.20"
                },
                {
                  // This note is from the open source repo, and thus has a FromDependentVersion field
                  "Note": "bumped go-utils dep to v1.6.4",
                  "FromDependentVersion": "v1.8.0-beta2"
                }
              ]
            }
          }
        }
      ]
    }
  ]
}
```
