IMAGE=simplinicqueue

.PHONY: build
build:
	@docker build -t ${IMAGE} .

.PHONY: run
run: check-CONFIG build
	docker run -v ${CONFIG}:/app/config/${notdir ${CONFIG}} ${IMAGE} --config /app/config/${notdir ${CONFIG}}
	
.PHONY: usage
usage: build
	@docker run ${IMAGE}

.PHONY: check-%
check-%:
	@if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi