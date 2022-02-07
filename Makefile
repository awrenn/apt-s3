build: ./apt-s3

./apt-s3: $(shell find -name *.go)
	go build -o ./apt-s3 ./cmd/main.go

simulate: ./apt-s3
	./apt-s3 -region us-west-2 -bucket dummy-apt -deb ./dummy.deb

test: ./dummy.deb
	go test ./...

./dummy.deb:
	dpkg-deb --build ./dummy ./dummy.deb

repo:
	echo deb [trusted=yes] https://$(shell terraform output -json  | jq -r .dummy_repo_url.value) main stable
