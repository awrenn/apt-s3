build:
	go build -o ./apt-s3 ./cmd/main.go


test: ./dummy.deb
	go test ./...

./dummy.deb:
	dpkg-deb --build ./dummy ./dummy.deb
