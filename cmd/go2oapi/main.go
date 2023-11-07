package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tmc/go2oapi"
)

var (
	srcDir      = flag.String("src", ".", "The directory to scan for Go source files.")
	handlerFunc = flag.String("func", "", "The name of the function to generate OpenAPI definitions for.")
	outputFile  = flag.String("output", "-", "The output file where the function details will be written.")
)

func main() {
	// Parse the provided flags.
	flag.Parse()
	// Parse the source directory for the specific function.
	funcDetails, err := go2oapi.ParseFunction(*srcDir, *handlerFunc)
	if err != nil {
		fmt.Printf("Error parsing function: %v\n", err)
		os.Exit(1)
	}
	// Generate the JSON output from the parsed function details.
	jsonOutput, err := generateJSON(funcDetails)
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		os.Exit(1)
	}
	// Output the JSON to the specified file.
	err = outputJSON(jsonOutput, *outputFile)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %v\n", err)
		os.Exit(1)
	}
}

// GenerateJSON takes the details of a function and generates a JSON representation.
func generateJSON(details *go2oapi.FunctionDetails) ([]byte, error) {
	return json.MarshalIndent(details, "", "  ")
}

// OutputJSON writes the JSON output to the specified file.
func outputJSON(jsonData []byte, filename string) error {
	if filename == "-" {
		_, err := io.Copy(os.Stdout, bytes.NewReader(jsonData))
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}
