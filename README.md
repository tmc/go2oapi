# go2oapi - Go to OpenAPI Converter

go2oapi is a tool for converting Go function declarations into OpenAPI (Swagger) definitions.

It parses your Go source files, identifies function parameters, and generates associated OpenAPI
definitions suitable for use in [OpenAI Function
Calling](https://platform.openai.com/docs/guides/function-calling).

## Features

- **Automated Parsing**: Automatically parses Go source files to find function declarations.
- **Rich OpenAPI Definitions**: Creates comprehensive OpenAPI definitions including types, descriptions, and enums.
- **Error Handling**: Provides clear error messages for unsupported types and parsing issues.
- **Reflection-Free**: Operates without the need for Go reflection, ensuring type safety and straightforward code.

## Installation

You can directly install the cli version of the tool via `go get` (presuming you have go installed correctly).

```bash
go get github.com/tmc/go2oapi/cmd/go2oapi
```

## Usage

You can use go2oapi either as a package or a command line tool.

### Usage as a library

```go
package main

import (
    "github.com/tmc/go2oapi"
    "log"
)

func main() {
    details, err := go2oapi.ParseFunction("path/to/your/go/program/", "YourFunctionName")
    if err != nil {
        log.Fatalf("Error parsing function: %s\n", err)
    }

    // Use `details` for further processing or output
}
```

### go2oapi command line

To generate the OpenAPI tool signature for `strings.Join`:
```
$ go2oapi -src $(go env GOROOT)/src/strings -func Join
{
  "name": "Join",
  "description": "Join concatenates the elements of its first argument to create a single string. The separator string sep is placed between elements in the resulting string.",
  "parameters": {
    "type": "object",
    "properties": {
      "elems": {
        "type": "array",
        "items": {
          "type": "string"
        }
      },
      "sep": {
        "type": "string"
      }
    },
    "required": [
      "elems",
      "sep"
    ]
  }
}
```
