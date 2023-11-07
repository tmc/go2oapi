package go2oapi_test

import (
	"encoding/json"
	"log"
	"os"

	"github.com/tmc/go2oapi"
)

func ExampleParseFunction() {
	// Assuming we have a file named 'example.go' in the current directory
	// with a function 'HelloWorld' that we want to parse.
	filePath := "testdata/sample-a"
	functionName := "NewWidgetFactory"

	// Parse the function to get the details
	funcDetails, err := go2oapi.ParseFunction(filePath, functionName)
	if err != nil {
		log.Fatalf("Error parsing function: %v\n", err)
	}

	// Output the function details
	// (In a real test, this would be used to validate the output against expected results)
	json.NewEncoder(os.Stdout).Encode(funcDetails)

	// Output:
	// {"name":"NewWidgetFactory","description":"NewWidgetFactory creates a new widget factory.","parameters":{"type":"object","properties":{"factoryInfo":{"type":"object","properties":{"Category":{"type":"string","description":"Category","enum":["foo","bar"]},"FactoryName":{"type":"string","description":"The name of the factory"},"InventoryLevels":{"type":"array","description":"InventoryLevels","items":{"type":"integer"}},"Operational":{"type":"boolean"}}}},"required":["factoryInfo"]}}
}
