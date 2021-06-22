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

.PHONY: format-code
format-code: install-go-tools
	goimports -w .