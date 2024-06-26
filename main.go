package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/umputun/go-flags"
	"google.golang.org/api/option"
)

type options struct {
	Model           string        `long:"model" env:"GEMINI_MODEL" default:"gemini-1.5-pro-preview-0409" description:"Model"`
	GoogleProjectID string        `long:"project" env:"GOOGLE_PROJECT_ID" description:"Google Project ID"`
	CredFile        string        `long:"cred" env:"GOOGLE_CRED_FILE" description:"Google Credential File"`
	InputDir        string        `long:"in" env:"INPUT_DIR" default:"./prompts" description:"Input directory"`
	OutputDir       string        `long:"out" env:"OUTPUT_DIR" default:"./out" description:"Output directory"`
	Temperature     float32       `long:"temp" env:"TEMPERATURE" default:"1" description:"Temperature"`
	MaxTokens       int           `long:"max_tokens" env:"MAX_TOKENS" default:"8000" description:"Max tokens"`
	Workers         int           `long:"workers" env:"WORKERS" default:"500" description:"Workers"`
	Delay           time.Duration `long:"delay" env:"DELAY" default:"500ms" description:"Delay between requests in ms. Shouldn't be more than 60000 / req per min limit (5 by default) / number of locations"`
	Timeout         time.Duration `long:"timeout" env:"TIMEOUT" default:"300s" description:"Timeout for each request"`
	Limit           int           `long:"limit" env:"LIMIT" default:"0" description:"Limit files to process. Can be used for testing. 0 means no limit"`
}

var locations = []string{
	"us-south1",
	"us-central1",
	"us-west4",
	"us-east1",
	"us-east4",
	"us-west1",
	"northamerica-northeast1",
	"southamerica-east1",
	"europe-west1",
	"europe-north1",
	"europe-west3",
	"europe-west2",
	"europe-southwest1",
	"europe-west8",
	"europe-west4",
	"europe-west9",
	"europe-central2",
	"europe-west6",
	"asia-east1",
	"asia-east2",
	"asia-south1",
	"asia-northeast3",
	"asia-southeast1",
	"australia-southeast1",
	"asia-northeast1",
	"me-west1",
}

const redColor = "\033[31m"
const grayColor = "\033[90m"
const greenColor = "\033[32m"
const resetColor = "\033[0m"

func main() {
	var opts options
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			logError("cli error: ", err)
		}
		os.Exit(2)
	}

	fileNames, err := getFilesList(opts.InputDir)
	if err != nil {
		logError(fmt.Sprintf("Error reading input directory %s: ", opts.InputDir), err)
		return
	}

	if len(fileNames) == 0 {
		fmt.Println("No .prompt files found in the input directory: ", opts.InputDir)
	}

	if err = os.MkdirAll(opts.OutputDir, 0755); err != nil {
		logError("Error creating directory: ", err)
		return
	}

	clientPool := []*genai.Client{}
	for _, location := range locations {
		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
		client, err := genai.NewClient(ctx, opts.GoogleProjectID, location,
			option.WithCredentialsFile(opts.CredFile),
		)
		if err != nil {
			logError("Error creating client: ", err)
			return
		}
		clientPool = append(clientPool, client)
	}

	semaphore := make(chan struct{}, opts.Workers)
	results := make(chan error)
	ticker := time.NewTicker(opts.Delay)
	completed := 0
	total := len(fileNames)
	files := map[string]struct{}{}
	outputFile := ""
	clientIndex := 0
	for i, fName := range fileNames {
		outputFileName := strings.TrimSuffix(fName, ".prompt")
		outputFile = fmt.Sprintf("%s/%s", opts.OutputDir, outputFileName)
		if _, err := os.Stat(outputFile); err == nil {
			fmt.Printf(grayColor+"File %s already exists. Skipping.\n"+resetColor, outputFile)
			completed++
			continue
		}

		// Skip if the file has already been processed
		if _, ok := files[outputFile]; ok {
			fmt.Printf(grayColor+"File %s already processing. Skipping.\n"+resetColor, outputFile)
			completed++
			continue
		}

		// Get prompt from outputFilePath
		prompt, err := readFile(fmt.Sprintf("%s/%s", opts.InputDir, fName))
		if err != nil {
			logError("Error read file error: ", err)
			completed++
			continue
		}

		// Mark the file as being processed
		files[outputFile] = struct{}{}

		client := clientPool[clientIndex]
		clientIndex++
		if clientIndex >= len(clientPool) {
			clientIndex = 0
		}

		go func(count int, prompt, outputFile string) {
			ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)

			defer func() {
				cancel()
				<-semaphore
				results <- nil
			}()

			fmt.Printf("%d/%d %s Processing: %s\n", i+1, total, nowDateTime(), outputFile)
			output, err := QueryGemini(ctx, client, opts.Model,
				i+1, &opts.Temperature, opts.MaxTokens, prompt)
			if err != nil {
				results <- err
				return
			}

			err = ioutil.WriteFile(outputFile, []byte(output), 0644)
			if err != nil {
				results <- err
			}
		}(i, prompt, outputFile)

		<-ticker.C // Respect rate limit
		semaphore <- struct{}{}

	}

	// Close the results channel when all cities have been processed
	if completed == total {
		close(results)
	}

	// Gathering and handling errors
	errs := 0
	for err := range results {
		if err != nil {
			logError("Error:", err)
			errs++
		}
		completed++
		if completed == total {
			close(results)
		}
	}
	fmt.Println("Processing complete.")
}

func logError(pre string, err error) {
	if err != nil {
		fmt.Println(redColor + pre + " " + err.Error() + resetColor)
	}
}

func QueryGemini(ctx context.Context, client *genai.Client, model string, jobN int, temperature *float32, maxTokens int, prompt string) (string, error) {
	gemini := client.GenerativeModel(model)
	gemini.GenerationConfig.SetMaxOutputTokens(int32(maxTokens))
	if temperature != nil {
		gemini.Temperature = temperature
	}
	gemini.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryUnspecified,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
	}
	resp, err := gemini.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error in job %d: %v", jobN, err)
	}

	if resp == nil {
		return "", fmt.Errorf("empty response, job %d", jobN)
	}

	res := ""

	fmt.Printf(greenColor+"%d %s Job done\n"+resetColor, jobN, nowDateTime())

	if len(resp.Candidates) > 0 && resp.Candidates[0] != nil && resp.Candidates[0].Content != nil &&
		len(resp.Candidates[0].Content.Parts) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			res += fmt.Sprintf("%v", part)
		}
	}
	return res, nil
}

func nowDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getFilesList(path string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		if !file.IsDir() && file.Size() > 0 && strings.HasSuffix(file.Name(), ".prompt") {
			files = append(files, file.Name())
		}
	}
	return files, nil
}

func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
