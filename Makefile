-include artifacts/make/go.mk

artifacts/make/%.mk:
	@curl --create-dirs '-#Lo' "$@" "https://rinq.github.io/make/$*.mk?nonce=$(shell date +%s)"
