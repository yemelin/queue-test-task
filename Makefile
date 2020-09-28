IMAGE=simplinicqueue

.PHONY: build
build:
	@docker build -t ${IMAGE} .

.PHONY: run
run: check-CONFIG build
	@if [ -z "${DEST}" ]; \
	then \
		docker run --rm -v ${CONFIG}:/app/config/${notdir ${CONFIG}} \
		-e SRC_DIR=/app/out -it --name simpliniqueue\
		${IMAGE} --config /app/config/${notdir ${CONFIG}}; \
	else \
		docker run --rm -v ${CONFIG}:/app/config/${notdir ${CONFIG}} \
		-e SRC_DIR=/app/out -v ${DEST}:/app/out -it --name simpliniqueue\
		${IMAGE} --config /app/config/${notdir ${CONFIG}}; \
	fi

.PHONY: stop
stop:
	docker stop simpliniqueue

.PHONY: usage
usage: build
	@docker run ${IMAGE}

.PHONY: test
test:
	go test -tags=async

.PHONY: check-%
check-%:
	@if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi