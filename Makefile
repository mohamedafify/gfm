GC = go

all: main

main: main.go terminal.go log.go
	$(GC) build $^

clean:
	rm -f main
