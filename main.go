package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"

	"google.golang.org/api/option"
)

var model = os.Getenv("GEMINI_MODEL")

var googleProjectID = os.Getenv("GOOGLE_PROJECT_ID")

var credFile = os.Getenv("GOOGLE_CRED_FILE")

const limit = 0

const workers = 500

var delay = (60000 / 5 / 26) * time.Millisecond

const timeout = 5 * time.Minute

var inputDir = os.Getenv("INPUT_DIR")
var outputDir = os.Getenv("OUTPUT_DIR")

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
const greenColor = "\033[32m"
const resetColor = "\033[0m"

func main() {
	locationIndex := 0

	maxTokens := 8000
	tok, err := strconv.Atoi(os.Getenv("MAX_TOKENS"))
	if err == nil {
		fmt.Printf("Setting max tokens to %d\n", tok)
		maxTokens = tok
	} else {
		fmt.Printf("Using default max tokens: %d\n", maxTokens)
	}

	fileNames, err := getFilesList(inputDir)
	if err != nil {
		logError("Error reading input directory: ", err)
		return
	}

	if len(fileNames) == 0 {
		fmt.Println("No .prompt files found in the input directory: ", inputDir)
	}

	if err = os.MkdirAll(outputDir, 0755); err != nil {
		logError("Error creating directory: ", err)
		return
	}

	semaphore := make(chan struct{}, workers)
	results := make(chan error)
	ticker := time.NewTicker(delay)
	completed := 0
	total := len(fileNames)
	files := map[string]struct{}{}
	outputFile := ""
	for i, fName := range fileNames {
		outputFileName := strings.TrimSuffix(fName, ".prompt")
		outputFile = fmt.Sprintf("%s/%s", outputDir, outputFileName)
		if _, err := os.Stat(outputFile); err == nil {
			fmt.Printf("File %s already exists. Skipping.\n", outputFile)
			completed++
			continue
		}

		// Skip if the file has already been processed
		if _, ok := files[outputFile]; ok {
			fmt.Printf("File %s already processing. Skipping.\n", outputFile)
			completed++
			continue
		}

		// Get prompt from outputFilePath
		prompt, err := readFile(fmt.Sprintf("%s/%s", inputDir, fName))
		if err != nil {
			logError("Error read file error: ", err)
			completed++
			continue
		}

		// Mark the file as being processed
		files[outputFile] = struct{}{}

		curLocation := locations[locationIndex]
		locationIndex++
		if locationIndex >= len(locations) {
			locationIndex = 0
		}

		go func(count int, prompt, outputFile string) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			defer func() {
				cancel()
				<-semaphore
				results <- nil
			}()

			fmt.Printf("%d/%d %s Processing: %s\n", i+1, total, nowDateTime(), outputFile)
			output, err := QueryGemini(ctx, i+1, maxTokens, curLocation, prompt)
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
		if completed == total || (limit > 0 && errs >= limit) {
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

func QueryGemini(ctx context.Context, jobN, maxTokens int, location, prompt string) (string, error) {
	client, err := genai.NewClient(ctx, googleProjectID, location,
		option.WithCredentialsFile(credFile),
	)
	if err != nil {
		return "", err
	}
	gemini := client.GenerativeModel(model)
	gemini.GenerationConfig.SetMaxOutputTokens(int32(maxTokens))
	resp, err := gemini.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	res := ""

	if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReasonMessage != "" {
		fmt.Printf(redColor+"Prompt blocked: %s\n"+resetColor, resp.PromptFeedback.BlockReasonMessage)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].FinishMessage != "" {
		fmt.Printf(redColor+"Prompt finished: %s\n"+resetColor, resp.Candidates[0].FinishMessage)
	}

	fmt.Printf(greenColor+"%s Job %d done\n"+resetColor, nowDateTime(), jobN)

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
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
