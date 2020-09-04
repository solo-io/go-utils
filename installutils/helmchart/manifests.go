package helmchart

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/installutils"
	"github.com/solo-io/go-utils/installutils/helmignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/releaseutil"

	"sigs.k8s.io/yaml"

	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/solo-io/go-utils/vfsutils"
	"github.com/spf13/afero"

	"github.com/solo-io/go-utils/contextutils"

	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Manifests []releaseutil.Manifest

func (m Manifests) Find(name string) *releaseutil.Manifest {
	for _, man := range m {
		if man.Name == name {
			return &man
		}
	}
	return nil
}

func (m Manifests) Names() []string {
	var names []string
	for _, m := range sortByKind(m) {
		names = append(names, m.Name)
	}
	return names
}

func (m Manifests) CombinedString() string {
	buf := &bytes.Buffer{}

	for _, m := range sortByKind(m) {
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

func ManifestsFromResources(resources kuberesource.UnstructuredResources) (Manifests, error) {
	var resourceYamls []string
	for _, res := range resources {
		rawJson, err := runtime.Encode(unstructured.UnstructuredJSONScheme, res)
		if err != nil {
			return nil, err
		}
		yam, err := yaml.JSONToYAML(rawJson)
		if err != nil {
			return nil, err
		}
		resourceYamls = append(resourceYamls, string(yam))
	}

	return Manifests{{Head: &releaseutil.SimpleHead{}, Content: strings.Join(resourceYamls, "\n---\n")}}, nil
}

var commentRegex = regexp.MustCompile("#.*")

func IsEmptyManifest(manifest string) bool {
	removeComments := commentRegex.ReplaceAllString(manifest, "")
	removeNewlines := strings.Replace(removeComments, "\n", "", -1)
	removeDashes := strings.Replace(removeNewlines, "---", "", -1)
	removeSpaces := strings.Replace(removeDashes, " ", "", -1)
	return removeSpaces == ""
}

func RenderManifests(ctx context.Context, chartUri, values, releaseName, namespace, kubeVersion string) (Manifests, error) {

	file, err := tarutils.RetrieveArchive(afero.NewOsFs(), chartUri)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check chart requirements to make sure all dependencies are present in /charts
	chart, err := loader.LoadArchive(file)
	if err != nil {
		return nil, errors.Wrapf(err, "loading chart")
	}
	return renderManifests(ctx, chart, values, releaseName, namespace, kubeVersion)
}

func renderManifests(ctx context.Context, c *chart.Chart, values, releaseName, namespace, kubeVersion string) ([]releaseutil.Manifest, error) {
	valuesYaml, err := chartutil.ReadValues([]byte(values))
	if err != nil {
		return nil, err
	}

	chartValues, err := chartutil.ToRenderValues(c, valuesYaml, chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: namespace,
	}, nil)

	if err != nil {
		return nil, err
	}
	renderedTemplates, err := engine.Render(c, chartValues)
	if err != nil {
		return nil, err
	}

	for file, man := range renderedTemplates {
		if IsEmptyManifest(man) {
			contextutils.LoggerFrom(ctx).Debugf("is an empty manifest, removing %v", file)
			delete(renderedTemplates, file)
		}
	}
	manifests := installutils.SplitManifests(renderedTemplates)
	return sortByKind(manifests), nil
}

type GithubChartRef struct {
	Owner          string
	Repo           string
	Ref            string
	ChartDirectory string
}

func RenderChartsFromGithub(ctx context.Context, parentRef GithubChartRef) (map[string]*chart.Chart, error) {
	fs := afero.NewMemMapFs()
	codeDir, err := vfsutils.MountCode(fs, ctx, github.NewClient(nil), parentRef.Owner, parentRef.Repo, parentRef.Ref)
	if err != nil {
		return nil, err
	}
	defer fs.Remove(codeDir)
	chartParent := filepath.Join(codeDir, parentRef.ChartDirectory)
	subdirs, err := afero.ReadDir(fs, chartParent)
	if err != nil {
		return nil, err
	}
	charts := make(map[string]*chart.Chart)
	for _, subdir := range subdirs {
		chartRoot := filepath.Join(chartParent, subdir.Name())
		rules, err := getRulesFromArchive(fs, chartRoot)
		if err != nil {
			return nil, err
		}
		chart, err := loadFiles(rules, fs, chartRoot+"/")
		if err != nil {
			return nil, err
		}
		charts[subdir.Name()] = chart
	}
	return charts, nil
}

func RenderChartFromGithub(ctx context.Context, ref GithubChartRef) (*chart.Chart, error) {
	fs := afero.NewMemMapFs()
	codeDir, err := vfsutils.MountCode(fs, ctx, github.NewClient(nil), ref.Owner, ref.Repo, ref.Ref)
	if err != nil {
		return nil, err
	}
	defer fs.Remove(codeDir)
	chartRoot := filepath.Join(codeDir, ref.ChartDirectory)
	rules, err := getRulesFromArchive(fs, chartRoot)
	if err != nil {
		return nil, err
	}
	return loadFiles(rules, fs, chartRoot+"/")
}

func RenderManifestsFromGithub(ctx context.Context, ref GithubChartRef, values, releaseName, namespace, kubeVersion string) ([]releaseutil.Manifest, error) {
	chart, err := RenderChartFromGithub(ctx, ref)
	if err != nil {
		return nil, err
	}
	return renderManifests(ctx, chart, values, releaseName, namespace, kubeVersion)
}

func getRulesFromArchive(fs afero.Fs, chartRoot string) (*helmignore.Rules, error) {
	rules := helmignore.Empty()
	helmignorePath := filepath.Join(chartRoot, helmignore.HelmIgnore)
	exists, err := afero.Exists(fs, helmignorePath)
	if err != nil {
		return nil, errors.Wrapf(err, "error checking if helmignore exists")
	}
	if exists {
		contents, err := afero.ReadFile(fs, helmignorePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading helmignore")
		}
		reader := bytes.NewReader(contents)
		r, err := helmignore.Parse(reader)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse .helmignore")
		}
		rules = r
	}
	rules.AddDefaults()
	return rules, nil
}

func loadFiles(rules *helmignore.Rules, fs afero.Fs, chartDir string) (*chart.Chart, error) {
	var files []*loader.BufferedFile
	walk := func(name string, fi os.FileInfo, err error) error {
		n := strings.TrimPrefix(name, chartDir)
		if n == "" {
			// No need to process top level. Avoid bug with helmignore .* matching
			// empty names. See issue 1779.
			return nil
		}

		// Normalize to / since it will also work on Windows
		n = filepath.ToSlash(n)

		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(n, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If a .helmignore file matches, skip this file.
		if rules.Ignore(n, fi) {
			return nil
		}

		data, err := afero.ReadFile(fs, name)
		if err != nil {
			return fmt.Errorf("error reading %s: %s", n, err)
		}

		files = append(files, &loader.BufferedFile{Name: n, Data: data})
		return nil
	}
	if err := afero.Walk(fs, chartDir, walk); err != nil {
		return nil, err
	}
	return loader.LoadFiles(files)
}

// adapted from Helm 2
// https://github.com/helm/helm/blob/release-2.16/pkg/tiller/kind_sorter.go#L113

type kindSorter struct {
	ordering  map[string]int
	manifests []releaseutil.Manifest
}

func newKindSorter(m []releaseutil.Manifest, s []string) *kindSorter {
	o := make(map[string]int, len(s))
	for v, k := range s {
		o[k] = v
	}

	return &kindSorter{
		manifests: m,
		ordering:  o,
	}
}

func (k *kindSorter) Len() int { return len(k.manifests) }

func (k *kindSorter) Swap(i, j int) { k.manifests[i], k.manifests[j] = k.manifests[j], k.manifests[i] }

func (k *kindSorter) Less(i, j int) bool {
	a := k.manifests[i]
	b := k.manifests[j]
	first, aok := k.ordering[a.Head.Kind]
	second, bok := k.ordering[b.Head.Kind]

	if !aok && !bok {
		// if both are unknown then sort alphabetically by kind and name
		if a.Head.Kind != b.Head.Kind {
			return a.Head.Kind < b.Head.Kind
		}
		return a.Name < b.Name
	}

	// unknown kind is last
	if !aok {
		return false
	}
	if !bok {
		return true
	}

	// if same kind sub sort alphanumeric
	if first == second {
		return a.Name < b.Name
	}
	// sort different kinds
	return first < second
}

// SortByKind sorts manifests in InstallOrder
func sortByKind(manifests []releaseutil.Manifest) []releaseutil.Manifest {
	ordering := kuberesource.InstallOrder
	ks := newKindSorter(manifests, ordering)
	sort.Sort(ks)
	return ks.manifests
}
