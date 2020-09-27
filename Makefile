IMAGE=simplinicqueue

.PHONY: build
build:
	@docker build -t ${IMAGE} .

.PHONY: run
run: check-CONFIG
	@if [ -z "${DEST}" ]; \
	then \
		docker run -v ${CONFIG}:/app/config/${notdir ${CONFIG}} \
		-e SRC_DIR=/app/out \
		${IMAGE} --config /app/config/${notdir ${CONFIG}}; \
	else \
		docker run -v ${CONFIG}:/app/config/${notdir ${CONFIG}} \
		-e SRC_DIR=/app/out -v ${DEST}:/app/out \
		${IMAGE} --config /app/config/${notdir ${CONFIG}}; \
	fi


.PHONY: usage
usage: build
	@docker run ${IMAGE}

.PHONY: check-%
check-%:
	@if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi