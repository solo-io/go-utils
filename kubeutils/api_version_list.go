package kubeutils

// ApiVersionList implements sort.Interface
type ApiVersionList []ApiVersion

func (list ApiVersionList) Len() int {
	return len(list)
}

func (list ApiVersionList) Less(i, j int) bool {
	return list[i].LessThan(list[j])
}

func (list ApiVersionList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
