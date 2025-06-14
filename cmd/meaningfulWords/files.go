package main

import (
	"os"
)



func appendWordsToFile(file *os.File, words []word) error {
	for _, w := range words {
		wordStr := w.word + "\n"
		_, err := file.WriteString(wordStr)
		if err != nil {
			return err
		}
	}
	return nil
}
