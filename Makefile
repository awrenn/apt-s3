GPG_KEY?=dummy
REGION?=us-west-2
ARCH?=amd64
BUCKET?=dummy-apt

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

./dummy.deb:
	dpkg-deb --build ./dummy ./dummy.deb

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
