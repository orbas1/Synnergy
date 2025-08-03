package main

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[EXPLORER] ", log.LstdFlags|log.Lmicroseconds)
