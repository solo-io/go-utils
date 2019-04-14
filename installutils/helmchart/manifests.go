package helmchart

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"k8s.io/helm/pkg/ignore"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"

	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/tiller"
)

type Manifests []manifest.Manifest

func (m Manifests) Find(name string) *manifest.Manifest {
	for _, man := range m {
		if man.Name == name {
			return &man
		}
	}
	return nil
}

func (m Manifests) Names() []string {
	var names []string
	for _, m := range tiller.SortByKind(m) {
		names = append(names, m.Name)
	}
	return names
}

func (m Manifests) CombinedString() string {
	buf := &bytes.Buffer{}

	for _, m := range tiller.SortByKind(m) {
		data := m.Content
		b := filepath.Base(m.Name)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}
		fmt.Fprintf(buf, "---\n# Source: %s\n", m.Name)
		fmt.Fprintln(buf, data)
	}

	return buf.String()
}

var yamlSeparator = regexp.MustCompile("\n---")

func (m Manifests) ResourceList() (kuberesource.UnstructuredResources, error) {
	snippets := yamlSeparator.Split(m.CombinedString(), -1)

	var resources kuberesource.UnstructuredResources
	for _, objectYaml := range snippets {
		if IsEmptyManifest(objectYaml) {
			continue
		}
		jsn, err := yaml2json.ToJSON([]byte(objectYaml))
		if err != nil {
			return nil, err
		}
		uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
		if err != nil {
			return nil, err
		}
		if resourceList, ok := uncastObj.(*unstructured.UnstructuredList); ok {
			for _, item := range resourceList.Items {
				resources = append(resources, &item)
			}
			continue
		}
		resources = append(resources, uncastObj.(*unstructured.Unstructured))
	}

	return resources, nil
}

var commentRegex = regexp.MustCompile("#.*")

func IsEmptyManifest(manifest string) bool {
	removeComments := commentRegex.ReplaceAllString(manifest, "")
	removeNewlines := strings.Replace(removeComments, "\n", "", -1)
	removeDashes := strings.Replace(removeNewlines, "---", "", -1)
	removeSpaces := strings.Replace(removeDashes, " ", "", -1)
	return removeSpaces == ""
}

var defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)

func RenderManifests(ctx context.Context, chartUri, values, releaseName, namespace, kubeVersion string) (Manifests, error) {
	var file io.Reader
	if strings.HasPrefix(chartUri, "http://") || strings.HasPrefix(chartUri, "https://") {
		resp, err := http.Get(chartUri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("http GET returned status %d", resp.StatusCode)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(chartUri)
		if err != nil {
			return nil, errors.Wrapf(err, "getting absolute path for %v", chartUri)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "opening file %v", path)
		}
		file = f
	}

	if kubeVersion == "" {
		kubeVersion = defaultKubeVersion
	}
	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseName,
			IsInstall: true,
			Time:      timeconv.Now(),
			Namespace: namespace,
		},
		KubeVersion: kubeVersion,
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.LoadArchive(file)
	if err != nil {
		return nil, errors.Wrapf(err, "loading chart")
	}

	config := &chart.Config{Raw: values, Values: map[string]*chart.Value{}}
	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return nil, err
	}

	for file, man := range renderedTemplates {
		if IsEmptyManifest(man) {
			contextutils.LoggerFrom(ctx).Warnf("is an empty manifest, removing %v", file)
			delete(renderedTemplates, file)
		}
	}
	manifests := manifest.SplitManifests(renderedTemplates)
	return tiller.SortByKind(manifests), nil
}

type contentMap map[string]*github.RepositoryContent

func RenderChartFromGithub(ctx context.Context, org, repo, ref, chartDirectory string) (*chart.Chart, error) {
	toplevelContents, err := getGitDirectory(ctx, org, repo, ref, chartDirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading git directory")
	}
	rules, err := getRules(toplevelContents)
	if err != nil {
		return nil, err
	}
	prefix := chartDirectory + "/"
	files, err := getFiles(ctx, rules, prefix, org, repo, ref, toplevelContents)
	if err != nil {
		return nil, err
	}
	return chartutil.LoadFiles(files)
}

func getFiles(ctx context.Context, rules *ignore.Rules, prefix, org, repo, ref string, contents contentMap) ([]*chartutil.BufferedFile, error) {
	var files []*chartutil.BufferedFile
	for _, content := range contents {
		relativePath := strings.TrimPrefix(content.GetPath(), prefix)
		if content.GetType() == "dir" {
			if rules.Ignore(relativePath, GetFileInfo(content)) {
				continue
			} else {
				subdirContents, err := getGitDirectory(ctx, org, repo, ref, content.GetPath())
				if err != nil {
					return nil, errors.Wrapf(err, "error loading subdirectory")
				}
				subdirFiles, err := getFiles(ctx, rules, prefix, org, repo, ref, subdirContents)
				if err != nil {
					return nil, err
				}
				files = append(files, subdirFiles...)
				continue
			}
		}
		if rules.Ignore(relativePath, GetFileInfo(content)) {
			continue
		}
		loaded, err := loadFileContent(ctx, org, repo, ref, content.GetPath())
		if err != nil {
			return nil, err
		}
		file := chartutil.BufferedFile{
			Name: relativePath,
			Data: []byte(loaded),
		}
		files = append(files, &file)
	}
	return files, nil
}

func loadFileContent(ctx context.Context, org, repo, ref, path string) (string, error) {
	client := github.NewClient(nil)
	getOpts := github.RepositoryContentGetOptions{
		Ref: ref,
	}
	file, dir, _, err := client.Repositories.GetContents(ctx, org, repo, path, &getOpts)
	if err != nil {
		return "", errors.Wrapf(err, "error loading file contents")
	}
	if dir != nil {
		return "", errors.Errorf("expected file not directory")
	}
	return file.GetContent()
}

func GetFileInfo(content *github.RepositoryContent) os.FileInfo {
	return &githubFileInfo{content: content}
}

type githubFileInfo struct {
	content *github.RepositoryContent
}

func (g *githubFileInfo) Name() string {
	return g.content.GetName()
}

func (g *githubFileInfo) Size() int64 {
	return int64(g.content.GetSize())
}

func (g *githubFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (g *githubFileInfo) ModTime() time.Time {
	return time.Now()
}

func (g *githubFileInfo) IsDir() bool {
	return g.content.GetType() == "dir"
}

func (g *githubFileInfo) Sys() interface{} {
	return nil
}

func getRules(toplevelContents contentMap) (*ignore.Rules, error) {
	rules := ignore.Empty()
	if content, ok := toplevelContents[ignore.HelmIgnore]; ok {
		contentString, err := content.GetContent()
		if err != nil {
			return nil, errors.Wrapf(err, "unable to read .helmignore")
		}
		reader := strings.NewReader(contentString)
		r, err := ignore.Parse(reader)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse .helmignore")
		}
		rules = r
	}
	rules.AddDefaults()
	return rules, nil
}

func getGitDirectory(ctx context.Context, org, repo, ref, chartDirectory string) (contentMap, error) {
	client := github.NewClient(nil)
	getOpts := github.RepositoryContentGetOptions{
		Ref: ref,
	}
	file, dir, _, err := client.Repositories.GetContents(ctx, org, repo, chartDirectory, &getOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting repo contents")
	}
	if file != nil {
		return nil, errors.Errorf("Expected directory, found file")
	}
	return getContentMap(dir), nil
}

func getContentMap(contents []*github.RepositoryContent) contentMap {
	nameMap := make(map[string]*github.RepositoryContent)
	for _, content := range contents {
		nameMap[content.GetName()] = content
	}
	return nameMap
}
