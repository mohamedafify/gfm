GC = go

all: main

main: gfm.go terminal.go log.go
	$(GC) build $^

clean:
	rm -f main
