buildserver:
	go build -o build/server main.go
run:
	./build/server
fun: buildserver run