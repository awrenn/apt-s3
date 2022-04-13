GPG_KEY?=dummy
REGION?=us-west-2
ARCH?=amd64
BUCKET?=dummy-apt
VERSION?=1.0.0

MINIO_PORT?=9000
MINIO_NAME?=apt-s3_minio_tester

build: ./apt-s3

./apt-s3-amd64: $(shell find -name *.go)
	GOARCH=amd64 go build -o ./apt-s3-amd64 ./cmd/main.go

./apt-s3: $(shell find -name *.go)
	GOARCH=$(ARCH) go build -o ./apt-s3 ./cmd/main.go

simulate: ./apt-s3
	./apt-s3 -region $(REGION) -bucket dummy-apt -deb ./dummy.deb

simulate-sign: ./apt-s3
	./apt-s3 -region $(REGION) -bucket dummy-apt -deb ./dummy.deb -key $(GPG_KEY)

add-key: 
	gpg --export $(GPG_KEY) | sudo apt-key add -

test: 
	go test ./...

dummy: ./dummy_$(VERSION).deb

./dummy_$(VERSION).deb:
	mkdir -p ./dummy/dummy_$(VERSION)/DEBIAN
	cat ./files/dummy | sed -e "s|{{ VERSION }}|$(VERSION)|" > ./dummy/dummy_$(VERSION)/DEBIAN/control
	mkdir -p ./dummy/dummy_$(VERSION)/usr/local/bin
	echo "#!/bin/sh\necho 'test binary'" >  ./dummy/dummy_$(VERSION)/usr/local/bin/dummy
	chmod +x ./dummy/dummy_$(VERSION)/usr/local/bin/dummy
	dpkg-deb --build ./dummy/dummy_$(VERSION) ./dummy/dummy_$(VERSION).deb

repo:
	echo deb https://$(shell terraform output -json  | jq -r .dummy_repo_url.value) main stable

## We have to use the amd64 version to publish itself
publish: ./apt-s3-amd64 ./apt-s3_$(ARCH)_$(VERSION).deb
	./apt-s3-amd64 -region $(REGION) -bucket $(BUCKET) -deb ./apt-s3_$(ARCH)_$(VERSION).deb -key $(GPG_KEY)

./apt-s3_$(ARCH)_$(VERSION).deb: ./apt-s3
	mkdir -p ./out/DEBIAN
	cat files/control | sed \
		-e 's|{{ arch }}|$(ARCH)|' \
		-e 's|{{ version }}|$(VERSION)|' \
		> ./out/DEBIAN/control
	mkdir -p ./out/usr/bin
	cp ./apt-s3 ./out/usr/bin/apt-s3
	dpkg-deb --build ./out ./apt-s3_$(ARCH)_$(VERSION).deb
