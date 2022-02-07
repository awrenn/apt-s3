GPG_KEY?=dummy
build: ./apt-s3

./apt-s3: $(shell find -name *.go)
	go build -o ./apt-s3 ./cmd/main.go

simulate: ./apt-s3
	./apt-s3 -region us-west-2 -bucket dummy-apt -deb ./dummy.deb

simulate-sign: ./apt-s3
	./apt-s3 -region us-west-2 -bucket dummy-apt -deb ./dummy.deb -key $(GPG_KEY)

add-key: 
	gpg --export $(GPG_KEY) | sudo apt-key add -

test: ./dummy.deb
	go test ./...

./dummy.deb:
	dpkg-deb --build ./dummy ./dummy.deb

repo:
	echo deb https://$(shell terraform output -json  | jq -r .dummy_repo_url.value) main stable
