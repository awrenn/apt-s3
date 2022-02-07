GPG_KEY?=dummy
REGION?=us-west-2
ARCH?=amd64
BUCKET?=dummy-apt
VERSION?=1.0.0

build: ./apt-s3

./apt-s3: $(shell find -name *.go)
	go build -o ./apt-s3 ./cmd/main.go

simulate: ./apt-s3
	./apt-s3 -region $(REGION) -bucket dummy-apt -deb ./dummy.deb

simulate-sign: ./apt-s3
	./apt-s3 -region $(REGION) -bucket dummy-apt -deb ./dummy.deb -key $(GPG_KEY)

add-key: 
	gpg --export $(GPG_KEY) | sudo apt-key add -

test: ./dummy.deb
	go test ./...

dummy: ./dummy_$(VERSION).deb

./dummy_$(VERSION).deb:
	mkdir -p ./dummy_$(VERSION)/DEBIAN
	mkdir -p ./dummy
	cat ./files/dummy | sed -e "s|{{ VERSION }}|$(VERSION)|" > ./dummy_$(VERSION)/DEBIAN/control
	mkdir -p ./dummy_$(VERSION)/usr/local/bin
	echo "#!/bin/sh\necho 'test binary'" >  ./dummy_$(VERSION)/usr/local/bin/dummy
	chmod +x ./dummy_$(VERSION)/usr/local/bin/dummy
	dpkg-deb --build ./dummy_$(VERSION) ./dummy/dummy_$(VERSION).deb

repo:
	echo deb https://$(shell terraform output -json  | jq -r .dummy_repo_url.value) main stable

publish: ./apt-s3_$(ARCH).deb
	./apt-s3 -region $(REGION) -bucket $(BUCKET) -deb ./apt-s3_$(ARCH).deb -key $(GPG_KEY)

./apt-s3_$(ARCH).deb: ./apt-s3
	mkdir -p ./out/DEBIAN
	cp files/control ./out/DEBIAN/control
	mkdir -p ./out/usr/bin
	cp ./apt-s3 ./out/usr/bin/apt-s3
	dpkg-deb --build ./out ./apt-s3_$(ARCH).deb
