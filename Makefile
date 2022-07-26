
test:
	go test -v ./...


# .PHONY: tag
# tag:
# 	tag=$(git describe --tags `git rev-list --tags --max-count=1`)
# 	echo $(tag)