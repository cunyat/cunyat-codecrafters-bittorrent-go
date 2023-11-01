package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

type TorrentFileInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
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

type GetPeersResponse struct {
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Incomplete  int    `bencode:"incomplete"`
	Complete    int    `bencode:"complete"`
	Peers       string `bencode:"peers"`
}

func (r GetPeersResponse) PeersAddr() []string {
	bytePeers := []byte(r.Peers)
	var addrs []string
	for i := 0; i < len(bytePeers); i += 6 {
		ip := net.IPv4(bytePeers[i], bytePeers[i+1], bytePeers[i+2], bytePeers[i+3])
		port := int(bytePeers[i+4]) << 8
		port |= int(bytePeers[i+5])
		addrs = append(addrs, fmt.Sprintf("%s:%d", ip.String(), port))
	}
	return addrs
}

func GetPeers(t TorrentFile) (GetPeersResponse, error) {
	trackerURL, err := url.Parse(t.Announce)
	if err != nil {
		return GetPeersResponse{}, fmt.Errorf("bad url for torrent announce: %w", err)
	}
	hash, err := TorrentFileHash(t)
	if err != nil {
		return GetPeersResponse{}, err
	}
	q := trackerURL.Query()
	q.Add("info_hash", string(hash))
	q.Add("peer_id", "18243745892367492361")
	q.Add("port", "6881")
	q.Add("uploaded", "0")
	q.Add("downloaded", "0")
	q.Add("left", fmt.Sprintf("%d", t.Info.Length))
	q.Add("compact", "1")
	trackerURL.RawQuery = q.Encode()

	fmt.Println("url", trackerURL.String())

	res, err := http.Get(trackerURL.String())
	if err != nil {
		return GetPeersResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return GetPeersResponse{}, fmt.Errorf("got error response: %d %s", res.StatusCode, res.Status)
	}

	content, err := io.ReadAll(res.Body)
	fmt.Printf("status: %d\n", res.StatusCode)
	fmt.Printf("res: %s\n", content)
	fmt.Printf("err: %s\n", err)

	var peers GetPeersResponse
	if err := bencode.Unmarshal(bytes.NewReader(content), &peers); err != nil {
		return GetPeersResponse{}, fmt.Errorf("unable to parse get peers response: %w", err)
	}
	return peers, nil
}

type HandshakeResponse struct {
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

		fmt.Printf("Tracker URL: %s", torrent.Announce)
		fmt.Printf("Length: %d\n", torrent.Info.Length)
		fmt.Printf("Info Hash: %x\n", hash)
		fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
		fmt.Println("Pieces Hashes:")
		for i := 0; i < len(torrent.Info.Pieces); i += 20 {
			fmt.Printf("%x\n", torrent.Info.Pieces[i:i+20])
		}
	case "peers":
		filename := os.Args[2]
		torrent, err := ParseTorrentFile(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		peers, err := GetPeers(torrent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, peer := range peers.PeersAddr() {
			fmt.Println(peer)
		}
	case "handshake":
		filename := os.Args[2]
		peerAddr := os.Args[3]
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

		conn, err := net.Dial("tcp", peerAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("connected")
		defer conn.Close()

		buf := bytes.Buffer{}

		buf.WriteByte(19)
		buf.WriteString("BitTorrent protocol")
		buf.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
		buf.Write(hash)
		buf.WriteString("00112233445566778899")

		fmt.Println("sending handshake")
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resBuf := bytes.Buffer{}
		fmt.Println("waiting response")
		n, err := resBuf.ReadFrom(conn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		res := resBuf.Bytes()
		fmt.Println("bytes: ", n)
		peerID := res[n-20 : n]
		fmt.Printf("Peer ID: %x\n", peerID)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
