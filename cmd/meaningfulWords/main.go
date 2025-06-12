package main

import (
	"flag"
	"github.com/wexlerdev/randomwords/internal/wordlist"
	"fmt"
	"os"
	"log"
	"context"
	"google.golang.org/genai"
	"github.com/joho/godotenv"
)

type Config struct {
	apiKey		string
}

func main() {
	src := flag.String("s", "", "source file to read from")

	flag.Parse()

	file, err := os.Open(*src)
	if err != nil {
		log.Fatalf("err opening file: %v", err)
		return
	}

	fmt.Println("converting file to array of words...")
	words, err := wordlist.FileToWordArray(file)
	if err != nil {
		log.Fatalf("err converting file to array, %v", err)
		return
	}
	
	fmt.Printf("Finished Converting File to array of words!\n")

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Err loading the env api key")
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")

	cfg := Config{
		apiKey: apiKey,
	}	

	err = cfg.askGeminiAboutWords(*words)
	if err != nil {
		log.Fatalf("err calling gemini, %v", err)
		return
	}
	
}

func (cfg * Config) askGeminiAboutWords(words []string) error {
	ctx := context.TODO()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil
	}

	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text("give a fun rap back about buffalo bills"), nil)
	if err != nil {
		return nil
	}
	fmt.Println(result.Text())
	return nil
}

