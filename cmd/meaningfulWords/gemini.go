package main

import (
	"fmt"
	"context"
	"strings"
	"google.golang.org/genai"
	"encoding/json"
	"strconv"
)

type geminiParams struct {
	client					*genai.Client
	genContentConfig		*genai.GenerateContentConfig
	modelStr				string
}

type wordInput struct {
	word		string
	rating		string
}

type word struct {
	word	string
	rating	float32
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

	for _, words := range wordChunks {
		//put words in a string seperated by ,
		request := strings.Join(words, ", ") 

		//send request to gemini
		result, err := geminiParams.client.Models.GenerateContent(ctx, modelChoice, genai.Text(request), geminiParams.genContentConfig)
		if err != nil {
			return err
		}
		resultBytes, err := result.MarshalJSON()
		if err != nil {
			return err
		}
		//words input
		var wordsInput []wordInput
		err = json.Unmarshal(resultBytes, wordsInput)
		if err != nil {
			return err
		}
		wordSlice, err := convertToWord(wordsInput)
		if err != nil {
			return err
		}
		//write to files

	}

	fmt.Printf("amount of words: %v\n", len(request))

	fmt.Println("got result!!")
	fmt.Println(result.Text())
	return nil
}

func convertToWord(wordsIn []wordInput) ([]word, error) {
	wordSlice := make([]word, 0, len(wordsIn))
	for _, wIn := range wordsIn {
		float64Rating, err := strconv.ParseFloat(wIn.rating, 32)
		if err != nil {
			return nil, err
		}
		wurd := word {
			word: wIn.word,
			rating: float32(float64Rating),
		}
		wordSlice = append(wordSlice, wurd)
	}
	return wordSlice, nil

}

