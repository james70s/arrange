# export CGO_ENABLED=0
# export GOOS=linux
# export GOARCH=amd64

Name=kaoca	# 项目名称
BINARY=main # 编译出来的文件名
VERSION=$(shell git describe --abbrev=0 --tags 2> /dev/null | sed 's/kaoca\/v//')  # 版本号 该命令查找从提交可访问的最新标记。
BUILD=$(shell git rev-parse --short HEAD 2> /dev/null || echo "undefined")  # build=ea74c49
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"  # 编译参数
DOCKERIMAGE=hkccr.ccs.tencentyun.com/safore-com/$(Name)

# go env
GO_VERSION=$(shell go version)

banner:
	@echo Project:    $(Name)
	@echo Go Version: ${GO_VERSION}
	@echo Go Path:    ${GOPATH}
	@echo Version:    ${VERSION}
	@echo Build:      ${BUILD}
	@echo

# 编译执行文件
binary: banner
	@echo "building binary..."
	go build -o $(BINARY) $(LDFLAGS)
	@echo [*] Done building $(BINARY)



release: banner
# Extract the current version number
# Increment the version number
# $(eval NEW_VERSION=$(shell echo $(VERSION) | python -c "version = input().split('.'); version[-1] = str(int(version[-1]) + 1); print('.'.join(version))"))
	$(eval NEW_VERSION=$(shell echo $(VERSION) | awk -F. '{printf "%d.%d.%d\n", $$1, $$2, $$3+1}'))
# Replace the version number in pyproject.toml
# sed -i '' 's/version = "$(VERSION)"/version = "$(NEW_VERSION)"/' ./services/$(service)/pyproject.toml
# Tag the commit with the new version number
	$(eval TAG=$(shell echo $(Name)/v$(NEW_VERSION)))
	@echo "Bumping version to $(NEW_VERSION) and tagging as $(TAG)"
# Commit the change and push to git
# git add .
	@if ! git diff-index --quiet HEAD --; then \
        git add .; \
		git commit -m "ci: Build $(Name) v$(NEW_VERSION)"; \
	fi
	git tag $(TAG)
	git push origin HEAD $(TAG)
	git remote -v

# clean:
# 	$(GOCLEAN)
# 	$(RMTARGZ)


# release:
# 	# Clean
# 	$(GOCLEAN)
# 	$(RMTARGZ)
# 	# Build for mac
# 	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD)
# 	tar czvf ${BINARY}-mac64-${VERSION}.tar.gz ./${BINARY}
# 	# Build for arm
# 	$(GOCLEAN)
# 	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD)
# 	tar czvf ${BINARY}-arm64-${VERSION}.tar.gz ./${BINARY}
# 	# Build for linux
# 	$(GOCLEAN)
# 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD)
# 	tar czvf ${BINARY}-linux64-${VERSION}.tar.gz ./${BINARY}
# 	# Build for win
# 	$(GOCLEAN)
# 	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD).exe
# 	tar czvf ${BINARY}-win64-${VERSION}.tar.gz ./${BINARY}.exe
# 	$(GOCLEAN)


.PHONY: banner binary release
