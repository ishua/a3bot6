.PHONY: build run templ clean

templ:
	templ generate

build: templ
	go build -o build/a3b-webui ./cmd/server

run: build
	./build/a3b-webui

clean:
	rm -rf build/
	rm -f *_templ.go
