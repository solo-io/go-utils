#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

.PHONY: install-go-tools
install-go-tools:
	go install golang.org/x/tools/cmd/goimports

.PHONY: go-fmt
go-fmt: install-go-tools
	gofmt -w .
	goimports -w .