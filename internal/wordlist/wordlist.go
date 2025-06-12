package wordlist

import (
	"os"
	"fmt"
	"bufio"
)

func AppendWords(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file '%s': %w", srcPath, err)
	}
	defer src.Close()

	// Using 0644 permissions (read/write for owner, read-only for group/others)
	dst, err := os.OpenFile(dstPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file '%s' in append mode: %w", dstPath, err)
	}
	defer dst.Close()
	newWords, err := FileToWordArray(src)
	if err != nil {
		return err
	}
	fmt.Printf("DEBUG: Appending to actual file path: %s\n", dst.Name())


	for _, newWord := range *newWords {
		fmt.Printf("%v, ", newWord)
		dst.WriteString(newWord + "\n")
	}
	fmt.Print("\n")

	fmt.Printf("wrote %v new words to file %v, from file %v", len(*newWords), dstPath, srcPath)
	fmt.Println()

	return nil
}

func FileToWordArray(file *os.File) (*[]string, error) {
	scanner := bufio.NewScanner(file)
	wordSlice := make([]string, 0, 50000)
	lineNum := 1
	for scanner.Scan() {
		wurd := scanner.Text()
		wordSlice = append(wordSlice, wurd)
		lineNum++
	}
	return &wordSlice, nil
}
