main.wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm main.go

serve: main.wasm
	go run server.go

clean:
	rm -f main.wasm
