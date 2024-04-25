package main

import (
	"fmt"

	"github.com/bxelab/runestone"
)

func main() {
	runeName := "STUDYZY.GMAIL.COM"
	myRune, err := runestone.SpacedRuneFromString(runeName)
	if err != nil {
		fmt.Println(err)
		return
	}
	etching := &runestone.Etching{
		Rune: &myRune.Rune,
	}
	r := runestone.Runestone{Etching: etching}
	data, err := r.Encipher()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Enciphered Etching data: %x\n", data)
}
