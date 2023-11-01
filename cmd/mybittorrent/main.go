package main

import (
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"os"
	"strings"
)

type TorrentFileInfo struct {
	Length      int
	Name        string
	PieceLength int `bencode:"piece length"`
	Pieces      string
}

type TorrentFile struct {
	Announce string
	Info     TorrentFileInfo
}

func ParseTorrentFile(filename string) (TorrentFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	info := TorrentFile{}
	if err := bencode.Unmarshal(file, &info); err != nil {
		return TorrentFile{}, err
	}

	return info, nil
}

func main() {
	command := os.Args[1]
	if command == "decode" {
		value := os.Args[2]
		decoded, err := bencode.Decode(strings.NewReader(value))
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		torrent, err := ParseTorrentFile(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Tracker URL: %s\nLength: %d\n", torrent.Announce, torrent.Info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
