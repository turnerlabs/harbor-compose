package cmd

import "log"

func check(e error) {
	if e != nil {
		log.Fatal("ERROR:", e)
	}
}
