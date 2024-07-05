package main

import (
	"log"
)

func main() {
	InitLogger()
	log.Print("Starting...")
	term := new(Terminal)
	term.New()
	term.HandleInput()
	defer term.quit()
}
