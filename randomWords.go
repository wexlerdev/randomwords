package main
import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"context"
)
					

func (cfg * Config) populateDbFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Println("reading words from file!")
	//make todo context
	todoContext := context.TODO()

	for scanner.Scan() {
		word := scanner.Text()
		word = strings.TrimSpace(word)
		if word == "" {
			break
		}

		//add word to db
		cfg.dbQueries.CreateWord(todoContext, word)

	}

	if err = scanner.Err(); err != nil {
		return err
	}
	fmt.Println("finished reading file")
	return nil

}
