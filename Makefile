# gfm - go file manager

GC = go
BIN = gfm
SRC = *.go

all: ${BIN}

${BIN}: ${SRC}
	$(GC) build -o ${BIN} $^ 

clean:
	rm -f ${BIN}

test: ${BIN} 
	alacritty -e ./gfm &
	sleep .5
	alacritty -e tail /tmp/gfm.log -f &

install: ${BIN}
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f ${BIN} ${DESTDIR}${PREFIX}/bin
	chmod 755 ${DESTDIR}${PREFIX}/bin/${BIN}

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/${BIN}

.PHONY: all clean install uninstall
