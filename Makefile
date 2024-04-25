ALL_SRC := $(shell find . -name '*.go' -type f | sort)
install-tools:
	@which golangci-lint > /dev/null || (echo 'install golangci-lint' && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3)
	go install github.com/google/addlicense@latest

add-license:
	addlicense -c 'The BxELab studyzy Authors' $(ALL_SRC)

check-license:
	@ADD_LICENSE_OUT=`addlicense -check $(ALL_SRC) 2>&1`; \
		if [ "$$ADD_LICENSE_OUT" ]; then \
			echo "addlicense FAILED => add License errors:\n"; \
			echo "$$ADD_LICENSE_OUT\n"; \
			echo "Use 'make add-license' to fix this."; \
			exit 1; \
		else \
			echo "Check License finished successfully"; \
		fi

build:
	cd ./cmd/runestonecli && go build -o ../../bin/runestonecli .

ut:
	go test -v ./...

lint:
	golangci-lint run ./...