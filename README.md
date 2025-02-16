# Configurable Snowflake ID Generator

This project is a configurable Snowflake ID generator implemented in Go. It allows you to generate unique IDs based on a custom epoch, machine ID, and sequence number.

## Features

- Configurable machine ID, epoch start time, and bit allocation for machine ID, sequence number, and time.
- Safe for concurrent use.
- Generates unique, monotonic IDs.

## Installation

To install the package, use `go get`:

```sh
go get github.com/opoccomaxao/go-snowflake
```

## Usage

Here is an example of how to use the Snowflake ID generator:

```go
package main

import (
    "fmt"
    "log"

    snowflake "github.com/opoccomaxao/go-snowflake"
)

func main() {
    generator, err := snowflake.New(snowflake.Config{
        MachineID: 1,
    })
    if err != nil {
        log.Fatalf("failed to create generator: %v", err)
    }

    id := generator.Next()
    fmt.Printf("Generated ID: %d\n", id)
}
```
