


## cloudbuildpreparation
VERSION := 0.0.1
NAME := cloudbuildpreparation
IMAGE_REPO := gcr.io/solo-io
IMAGE_SPEC := $(IMAGE_REPO)/$(NAME):$(VERSION)

.PHONY: $(NAME)-docker-build
$(NAME)-docker-build:
	docker build -t $(IMAGE_SPEC) .

.PHONY: $(NAME)-docker-push
$(NAME)-docker-push: $(NAME)-docker-build
	docker push $(IMAGE_SPEC)
