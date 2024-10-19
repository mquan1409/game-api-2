bootstrap: main.go ../../internal
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

run: bootstrap
	./bootstrap

clean:
	rm -f bootstrap