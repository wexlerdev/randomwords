package main

import (
	"fmt"
	"context"
	"strings"
	"google.golang.org/genai"
	"encoding/json"
	"strconv"
	"os"
	"log"
	"path/filepath"
)

type geminiParams struct {
	client					*genai.Client
	genContentConfig		*genai.GenerateContentConfig
	modelStr				string
}

type wordInput struct {
	Word		string		`json:"word"`
	Rating		string		`json:"rating"`
}

type word struct {
	word	string
	rating	int
}


func newSchema() *genai.Schema{
	return &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema {
					"word": {Type: genai.TypeString},
					"rating": {
						Type: genai.TypeString,
						Enum: []string{"0", "0.25", "0.5", "0.75", "1"},
					},
				}, 
			},
		}

}

func getSystemInstructionString() string {
	return `As an expert evaluator for freestyle rap lexicon, rate each word 0, 0.25, 0.75, or 1 for its **Rap Suitability** to a general audience. **Only consider English words.**
0: **HARSHLY EXCLUDE**:
    * Very common/basic words ("the," "a," "is," "it," "and," "but").
    * Nonsense or gibberish.
    * Non-English words.
    * **HARSHLY EXCLUDE**: Extremely specialized, obscure, or advanced medical, scientific, or chemical terms that a general audience would *not* know or understand easily in a rap context.
    * **PUNISH SIMILAR FORMS**: If a word is merely a common inflectional variant (e.g., plural, different verb tense, simple adverb/adjective derived from a noun) of a word that is otherwise suitable, assign it a 0. **Prioritize the most common or base form of the word.** For example, if 'run' gets a 1, then 'running,' 'ran,' and 'runs' should all receive a 0.
0.25: Understandable but less common or more specific to certain contexts; might be slightly awkward to fit into a freestyle.
0.75: Broadly understood and adds significant meaning to a rap lyric; good potential but perhaps not universally frequent or doesn't offer exceptionally versatile rhyming options.
1: Broadly understood, frequently usable in rap, adds significant meaning, possesses strong evocative qualities, and is widely applicable and versatile for rhyming.`

}

func (cfg * Config) newGeminiParams(model string) (* geminiParams, error) {
	ctx := context.TODO()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	//config
	schema := newSchema()
	sysInstructionString:= getSystemInstructionString()
	systemInstructions := genai.NewContentFromText(sysInstructionString, genai.RoleUser)
	temp := float32(0.6)
	topp := float32(0.6)

	config := &genai.GenerateContentConfig{
		Temperature: &temp,
		TopP: &topp,
		SystemInstruction: systemInstructions,
		ResponseMIMEType: "application/json",
		ResponseSchema: schema,
	}

	geminiParams := geminiParams {
		client: client,
		genContentConfig: config,
		modelStr: model,
	}

	return &geminiParams, err
}

func chunkWordsIntoSlices(words []string, chunkSize int) [][]string {
	var chunks [][]string

	if chunkSize <= 0 {
		return chunks
	}

	totalWords := len(words)
	for i := 0; i < totalWords; i += chunkSize {
		end := i + chunkSize
		if end > totalWords {
			end = totalWords
		}
		chunks = append(chunks, words[i:end])
	}
	return chunks
}

func (cfg * Config) askGeminiAboutWords(words []string) error {
	ctx := context.TODO()

	modelChoice := "models/gemini-2.0-flash"	
	geminiParams, err := cfg.newGeminiParams(modelChoice)
	if err != nil {
		return err
	}

	wordChunks := chunkWordsIntoSlices(words, 1200)

	logFile, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	// Ensure the file is closed when main exits
	defer logFile.Close()

	// 2. Set the output of the default logger to the file.
	log.SetOutput(logFile)

	// Optional: Configure log flags (date, time, short file name)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("[APP] ") // Optional: Add a custom prefix

	// Now, any log messages will go to "application.log"
	log.Println("This is a log message that goes to the file.")

	// --- NEW: Single output file setup ---
	outputDir := "output_rated_words"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}
	outputFilePath := filepath.Join(outputDir, "all_rated_words.txt")

	outputFile, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file '%s': %w", outputFilePath, err)
	}
	defer outputFile.Close() // Ensure the single output file is closed

	log.Printf("[askGeminiAboutWords] All processed words will be written to: %s", outputFilePath)

	for _, words := range wordChunks {
		//put words in a string seperated by ,
		request := strings.Join(words, ", ") 

		//send request to gemini
		result, err := geminiParams.client.Models.GenerateContent(ctx, modelChoice, genai.Text(request), geminiParams.genContentConfig)
		if err != nil {
			return err
		}
		log.Printf("RESULT: %v", result.Text())

		
		//words input
		var wordsInput []wordInput
		err = json.Unmarshal([]byte(result.Text()), &wordsInput)
		if err != nil {
			return err
		}
		log.Printf("RESULTWORDSLICE: %v", wordsInput[:10])
		
		fmt.Println("after unmarshaling!")

		wordSlice, err := convertToWord(wordsInput)
		if err != nil {
			return err
		}
		log.Printf("word slice: %v", wordSlice[:10])

		//write to files
		//err = addWordsToDir("testNew", wordSlice)
		err = addWordsToFile(outputFile, wordSlice)
		if err != nil {
			return err
		}

	}

	fmt.Printf("amount of words: %v\n", len(words))

	fmt.Println("got result!!")
	return nil
}

func addWordsToFile(file *os.File, words []word) error {
	for _, w := range words {
		line := fmt.Sprintf("%s (Rating: %d)\n", w.word, w.rating) // Include rating for clarity
		log.Printf("[addWordsToFile] Writing line: %s", strings.TrimSpace(line)) // Log what's being written
		_, err := file.WriteString(line)
		if err != nil {
			return fmt.Errorf("failed to write word '%s' to file: %w", w.word, err)
		}
		// Force data to be written to disk immediately after each word
		err = file.Sync()
		if err != nil {
			return fmt.Errorf("failed to sync file after writing word '%s': %w", w.word, err)
		}
	}
	return nil
}


func closeFilesInMap(files map[int]*os.File) {
	for _, file := range files {
		file.Close()
	} 
}

func addWordsToDir(dirPath string, words []word) error {
	fileMap, err := getFileMap(dirPath)
	if err != nil {
		return err
	}
	defer closeFilesInMap(fileMap)

	for _, w := range words {
		// Directly look up the file using the word's rating as the key
		if file, ok := fileMap[w.rating]; ok {
			_, err := file.WriteString(w.word + "\n")
			if err != nil {
				return fmt.Errorf("failed to write word '%s' to file for rating %v: %w", w.word, w.rating, err)
			}
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync file after writing word '%s' (rating %d) to file '%s': %w", w.word, w.rating, file.Name(), err)
			}
		} else {
			// This block executes if a word's rating doesn't exactly match
			// one of your predefined rating choices (0, 0.25, 0.5, 0.75, 1).
			// This could indicate an unexpected rating from Gemini or a precision issue.
			log.Printf("Warning: Word '%s' has an unexpected rating %v. Skipping file write.", w.word, w.rating)
		}
	}
	return nil
}

func getFileMap(dirPath string) (map[int]*os.File, error) {
	ratingChoices := []int{0, 25, 50, 75, 100}
	fileNameMap := make(map[int]string)
	//make sure dirPath exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to ensure directory '%s' exists: %w", dirPath, err)
	}

	for _, rChoice := range ratingChoices {
		fileNameMap[rChoice] = fmt.Sprintf("%s/words_%v.txt",dirPath, rChoice)
		log.Printf("fileName for rchoice: %v, %v\n", rChoice, fileNameMap[rChoice])
	}


	fileMap := make(map[int]*os.File)
	for _, rChoice := range ratingChoices {
		file, err := os.OpenFile(fileNameMap[rChoice], os.O_APPEND | os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}
		fileMap[rChoice] = file
	}
	return fileMap, nil

}

func convertToWord(wordsIn []wordInput) ([]word, error) {
	wordSlice := make([]word, 0, len(wordsIn))
	for _, wIn := range wordsIn {
		float64Rating, err := strconv.ParseFloat(wIn.Rating, 32)
		if err != nil {
			return nil, err
		}
		wurd := word {
			word: wIn.Word,
			rating: int(float64Rating * 100),
		}
		wordSlice = append(wordSlice, wurd)
	}
	return wordSlice, nil

}

