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
	if len(os.Args) <= 1 {
		fmt.Printf("Please specify image.\n")
		return
	}
	if err := run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

// ndup n-plicates a stream ; O(n) is M*N
func ndup(input []byte, output []byte, multiplier int) []byte {
	for ; multiplier > 0; multiplier-- {
		for i := 0; i < len(input); i++ {
			output = append(output, input[i])
		}
	}
	return output
}

func run(fname string) error {
	fmt.Printf("Flash image shadow command..\n")
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	var output []byte
	fmt.Printf("%d\n", len(dat))
	output = ndup(dat, output, 2)
	err = ioutil.WriteFile(os.Args[1], output, 0644)
	if err != nil {
		return err
	}
	return nil
}
