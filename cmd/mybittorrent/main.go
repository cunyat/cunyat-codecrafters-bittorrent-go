package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// decodeString parses a string value in bencode format
// assumes that the first byte in buf is a digit
func decodeString(buf string) (string, string, error) {
	sep := strings.IndexByte(buf, ':')
	if sep == -1 {
		return "", "", fmt.Errorf("missing string separator character")
	}

	length, err := strconv.Atoi(buf[:sep])
	if err != nil {
		return "", "", fmt.Errorf("could not parse string length: %w", err)
	}

	end := sep + 1 + length
	return buf[sep+1 : end], buf[end:], nil
}

// decodeInteger parses an integer value in bencode format
// assumes that the first byte in buf is an integer identifier 'i'
func decodeInteger(buf string) (int, string, error) {
	end := strings.IndexByte(buf, 'e')
	if end == -1 {
		return 0, "", fmt.Errorf("could not find integer end character")
	}

	digits := buf[1:end]
	value, err := strconv.Atoi(digits)
	if err != nil {
		return 0, "", err
	}

	return value, buf[end+1:], nil
}

// decodeList parses a list of values in bencode format
// assumes that the first byte in buf is a list identifier 'l'
func decodeList(buf string) ([]interface{}, string, error) {
	var items = []interface{}{}
	var value interface{}
	var err error
	var rest = buf[1:]

	for !strings.HasPrefix(rest, "e") {
		value, rest, err = decodeBencodeValue(rest)
		if err != nil {
			return nil, "", err
		}

		items = append(items, value)
	}

	return items, rest[1:], nil
}

// decodeDictionary parses a dictionary in bencode format
// assumes that the first byte in buf is a dictionary identifier 'd'
func decodeDictionary(buf string) (map[string]interface{}, string, error) {
	dict := map[string]interface{}{}
	rest := buf[1:]
	for !strings.HasPrefix(rest, "e") {
		// dictionary key must be a string
		key, r, err := decodeString(rest)
		if err != nil {
			return nil, "", fmt.Errorf("could not decode dictionary key, maybe not an string: %w", err)
		}
		rest = r
		value, r, err := decodeBencodeValue(rest)
		if err != nil {
			return nil, "", fmt.Errorf("error parsing decode value for key '%s': %w", key, err)
		}

		dict[key] = value
		rest = r
	}

	return dict, rest, nil
}

func decodeBencodeValue(buf string) (interface{}, string, error) {
	if 2 > len(buf) {
		// todo: this is right?
		return nil, "", fmt.Errorf("invalid bencoded value, it must have at least two bytes")
	}

	first := rune(buf[0])
	switch true {
	case first == 'l':
		return decodeList(buf)
	case first == 'd':
		return decodeDictionary(buf)
	case first == 'i':
		return decodeInteger(buf)
	case unicode.IsDigit(first):
		return decodeString(buf)
	default:
		return nil, "", fmt.Errorf("unexpected bencoded value format (%s)", buf)
	}
}

func decodeBencode(buf string) (interface{}, error) {
	value, _, err := decodeBencodeValue(buf)
	return value, err
}

type TorrentInfo struct {
	Announce string
	Length   int
}

func ParseTorrentInfo(filename string) (TorrentInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return TorrentInfo{}, err
	}
	defer file.Close()

	value, err := io.ReadAll(file)
	if err != nil {
		return TorrentInfo{}, err
	}

	decoded, err := decodeBencode(string(value))
	if err != nil {
		return TorrentInfo{}, err
	}

	dict, ok := decoded.(map[string]interface{})
	if !ok {
		return TorrentInfo{}, fmt.Errorf("expected torrent file to contain a dictionary, got %T", decoded)
	}

	announce, ok := dict["announce"]
	if !ok {
		return TorrentInfo{}, fmt.Errorf("missing announce field in torrent file")
	}

	announceStr, ok := announce.(string)
	if !ok {
		return TorrentInfo{}, fmt.Errorf("announce field must be an string, got %T", announce)
	}

	info, ok := dict["info"]
	if !ok {
		return TorrentInfo{}, fmt.Errorf("missing info field in torrent file")
	}

	infoDict, ok := info.(map[string]interface{})
	if !ok {
		return TorrentInfo{}, fmt.Errorf("info field is not a dictionary")
	}

	length, ok := infoDict["length"]
	if !ok {
		return TorrentInfo{}, fmt.Errorf("missing length field in torrent file's info field")
	}

	lengthInt, ok := length.(int)
	if !ok {
		return TorrentInfo{}, fmt.Errorf("length field must be an integer, got %T", length)
	}

	return TorrentInfo{Announce: announceStr, Length: lengthInt}, nil
}

func main() {
	command := os.Args[1]
	if command == "decode" {
		value := os.Args[2]
		decoded, err := decodeBencode(value)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		info, err := ParseTorrentInfo(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Tracker URL: %s\nLength: %d\n", info.Announce, info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
