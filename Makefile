# GIT ###

COMMIT_MESSAGE ?= .
TAG_NAME ?= .

push:
	@if [ "$(COMMIT_MESSAGE)" = "." ]; then \
		printf "\n!WRONG CMD\n\nUsage:\n   make push [ENVIRONMENT VARIABLES]\n\nENVIRONMENT VARIABLES:\n   COMMIT_MESSAGE - Specifies commit message\n"; \
	else \
		git add . && \
		git commit -m "$(COMMIT_MESSAGE)" && \
		git push origin; \
	fi

tag:
	@if [ "$(TAG_NAME)" = "." ]; then \
		printf "\n!WRONG CMD\n\nUsage:\n   make tag [ENVIRONMENT VARIABLES]\n\nENVIRONMENT VARIABLES:\n   TAG_NAME - Specifies tag name\n"; \
	else \
		git tag $(TAG_NAME) && \
		git push origin --tags; \
	fi

# TEST ##

TYPE ?= .

test:
	@case $(TYPE) in \
		.) go test -v ./internal/tests ;; \
		functional) go test -v -run _Functional ./internal/tests ;; \
		integration) go test -v -run _Integration ./internal/tests ;; \
	esac

# GO ###

EFFECTIVEMOBILE := cmd/effectivemobile/main.go 
MIGRATOR := cmd/migrator/main.go

run:
	go run $(EFFECTIVEMOBILE)

migrate:
	go run $(MIGRATOR)

build:
	go build -o build/effectivemobile $(EFFECTIVEMOBILE)