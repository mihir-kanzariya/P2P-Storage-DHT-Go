package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Info struct {
	Filename    string `json:"Filename"`
	ChunkName   string `json:"ChunkName"`
	FileKey     int    `json:"FileKey"`
	NoOfChunks  int    `json:"NoOfChunks"`
	ChunkIndex  int    `json:"ChunkIndex"`
	OwnerNodeId int    `json:"OwnerNodeId"`
}
type ChunkInfo struct {
	ChunkName  string `json:"ChunkName"`
	ChunkIndex int    `json:"ChunkIndex"`
}

const (
	ManifestPath = "demo.json"
)

func main() {
	// GetFileKeys("test.txt", 68)
	// GetFileChunks("test.txt", 68)
	// insertData(3, "test.txt", "chunk-2dsdsd", 25, 5, 68)
}

// Check if manifest file exists or not
// If not exists then just create new one
func CheckManifestExists() bool {
	manifest := fmt.Sprint(ManifestPath)

	_, err := os.Stat(manifest)
	if err != nil {
		err = os.Mkdir(manifest, 0755)
		if err != nil {
			return false
		}
	}
	return true
}

// Get Manifest file instance
func GetMenifest() []Info {
	if !CheckManifestExists() {
		return nil
	}
	jsonFile, err := os.Open(ManifestPath)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var nodeInfo []Info
	json.Unmarshal(byteValue, &nodeInfo)
	return nodeInfo
}

func GetFileKeys(filename string, nodeId int) {
	if !CheckManifestExists() {
		return
	}
	var nodeInfo = GetMenifest()
	keys := make([]int, 0)

	for i := 0; i < len(nodeInfo); i++ {
		data := nodeInfo[i]
		if data.Filename == filename {
			keys = append(keys, data.FileKey)
		}
	}
	fmt.Println("keys are ", keys)
}

func GetFileChunks(filename string, ownerNodeId int) []ChunkInfo {
	if !CheckManifestExists() {
		return nil
	}
	var nodeInfo = GetMenifest()
	chunkInfo := make([]ChunkInfo, 0)

	for i := 0; i < len(nodeInfo); i++ {
		data := nodeInfo[i]
		if data.Filename == filename {
			data := ChunkInfo{ChunkName: data.ChunkName, ChunkIndex: data.ChunkIndex}
			chunkInfo = append(chunkInfo, data)
		}
	}
	fmt.Println("chunkInfo are ", chunkInfo)

	return chunkInfo
}

func insertData(chunkIndex int, filename string, chunkName string, fileKey int, noOfChunks int, ownerNodeId int) {
	if !CheckManifestExists() {
		return
	}
	var info = Info{Filename: filename, ChunkIndex: chunkIndex, ChunkName: chunkName, FileKey: fileKey, NoOfChunks: noOfChunks, OwnerNodeId: ownerNodeId}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(info)

	file, err := ioutil.ReadFile(ManifestPath)
	if err != nil {
		log.Fatal(err)
	}
	var data []Info
	err = json.Unmarshal(file, &data)
	data = append(data, info)

	reqBodyBytes2 := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes2).Encode(data)
	ioutil.WriteFile(ManifestPath, reqBodyBytes2.Bytes(), 0644)

}
