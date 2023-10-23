package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
    if bencodedString[0] == 'i' {
        if len(bencodedString) <= 2 || bencodedString[len(bencodedString)-1] != 'e' {
            return "", fmt.Errorf("unexpected integer format (%s)", bencodedString)
        }

        digits := bencodedString[1:len(bencodedString)-1]
        value, err := strconv.Atoi(digits)
        if err != nil {
            return nil, err
        }

        return value, nil

    }
	if unicode.IsDigit(rune(bencodedString[0])) {
		firstColonIndex := strings.IndexByte(bencodedString, ':')

		if firstColonIndex == -1 {
			return "", fmt.Errorf("missing separator character")
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else {
		return "", fmt.Errorf("Only strings and integers are supported at the moment")
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		bencodedValue := os.Args[2]
		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
