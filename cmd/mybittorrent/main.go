package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"os"
	"strings"
)

type TorrentFileInfo struct {
	Length      int         `bencode:"length"`
	Name        string      `bencode:"name"`
	PieceLength int         `bencode:"piece length"`
	Pieces      interface{} `bencode:"pieces"`
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

func TorrentFileHash(file TorrentFile) ([]byte, error) {
	s := sha1.New()
	if err := bencode.Marshal(s, file.Info); err != nil {
		return nil, err
	}
	return s.Sum(nil), nil
}

func main() {
	command := os.Args[1]
	switch command {
	case "decode":
		value := os.Args[2]
		decoded, err := bencode.Decode(strings.NewReader(value))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	case "info":
		filename := os.Args[2]
		torrent, err := ParseTorrentFile(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		hash, err := TorrentFileHash(torrent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %x\n", torrent.Announce, torrent.Info.Length, hash)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
