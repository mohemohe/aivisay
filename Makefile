build: say.go
	go build -o say-go say.go

.PHONY: clean
clean:
	rm -f say-go