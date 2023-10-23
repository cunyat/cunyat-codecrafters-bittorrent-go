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

func decodeBencodeValue(str string) (interface{}, string, error) {
	if strings.HasPrefix(str, "i") && len(str) > 2 {
		end := strings.IndexByte(str, 'e')
		digits := str[1:end]
		value, err := strconv.Atoi(digits)
		if err != nil {
			return nil, "", err
		}

		return value, str[end:], nil
	}

	if strings.HasPrefix(str, "l") {
		var items []interface{}
		var value interface{}
		var err error
		var rest = str[1:]

		for {
			value, rest, err = decodeBencodeValue(rest)
			if err != nil {
				return nil, "", err
			}

			items = append(items, value)
			if strings.HasPrefix(rest, "e") {
				return items, rest[1:], nil
			}
		}
	}

	if unicode.IsDigit(rune(str[0])) {
		sep := strings.IndexByte(str, ':')

		if sep == -1 {
			return nil, "", fmt.Errorf("missing string separator character")
		}

		length, err := strconv.Atoi(str[:sep])
		if err != nil {
			return nil, "", fmt.Errorf("could not parse string length: %w", err)
		}

		end := sep + 1 + length
		return str[sep+1 : end], str[end:], nil
	}

	return nil, "", fmt.Errorf("unexpected bencoded value format (%s)", str)
}

func decodeBencode(bencodedString string) (interface{}, error) {
	value, _, err := decodeBencodeValue(bencodedString)
	return value, err
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
