MAIN    ?= server/main.go
RUN_ARGS = "-http-addr=localhost:2000 "

#PLATFORMS = linux/arm64 linux/amd64
PLATFORMS = linux/amd64

# Build the project
run:
	go run cmd/$(MAIN) $(RUN_ARGS)

run-test-config:
	RUN_ARGS="$(RUN_ARGS) -config-file ./test-connectors.yaml" \
	$(MAKE) run

.PHONY: schemas
schemas:
	@echo ">> Downloading schemas"
	@{ \
		VERSIONS="1.1.0 1.2.0 1.2.1 1.3.0 1.4.0 1.5.0 1.6.0" ; \
		URL="https://raw.githubusercontent.com/microsoft/AdaptiveCards/refs/heads/main/schemas/\{\}/adaptive-card.json" ; \
		mkdir -p schemas ; \
		echo $$VERSIONS | xargs -n 1 -I {} curl -sSL $$URL -o schemas/adaptive-card-{}.json ; \
	}

include common.Makefile
