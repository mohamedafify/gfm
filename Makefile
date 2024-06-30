GC = go

all: main

main: main.go terminal.go
	$(GC) build $^

clean:
	rm -f main
