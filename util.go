package runit

import (
	"log"
)

func handleErr(err error, do int) int {
	out := -1
	if err != nil {
		switch do {
		case 0:
			log.Printf("err: %v\n", err)
			panic(err)
		case 1:
			log.Printf("error: %v\n", err)
			out = 1
		case 2:
			out = 1
		}
	}

	return out

}
