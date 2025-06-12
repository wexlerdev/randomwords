package wordlist

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	destination := flag.String("d", "master.txt", "destination to add on to")	
	source := flag.String("s", "new.txt", "new words to add")

	flag.Parse()

	err := AppendWords(*source, *destination)
	if err != nil {
		fmt.Printf("err: %v", err)
		os.Exit(1)
	}

}

