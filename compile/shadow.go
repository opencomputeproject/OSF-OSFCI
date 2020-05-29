package main
import (
        "fmt"
        "io/ioutil"
	"os"
)
func main() {
        fmt.Printf("Flash image shadow command\n")
        dat, _ := ioutil.ReadFile(os.Args[1])
        var output []byte
        fmt.Printf("%d\n", len(dat))
        for i := 0 ; i < len(dat) ; i++ {
                output = append(output, dat[i])
        }
        for i := 0 ; i < len(dat) ; i++ {
                output = append(output, dat[i])
        }
        _ = ioutil.WriteFile(os.Args[1], output, 0644)
}
