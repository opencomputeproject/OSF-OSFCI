// Shadow picks up a 32MB rom file and duplicates it to create a 64MB version
// original file is destroyed/overwritten so please use a copy

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Printf("Flash image shadow command\n")
		dat, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		var output []byte
		fmt.Printf("%d\n", len(dat))
		for i := 0; i < len(dat); i++ {
			output = append(output, dat[i])
		}
		for i := 0; i < len(dat); i++ {
			output = append(output, dat[i])
		}
		err = ioutil.WriteFile(os.Args[1], output, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Please specify image.\n")
	}
}
