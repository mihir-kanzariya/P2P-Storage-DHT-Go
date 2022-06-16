package fileshare

import (
	"fmt"
	"os"
	"path/filepath"

	"io"

	"github.com/klauspost/reedsolomon"
)


func erasureEncoding(dataShards int,parShards int, inputFile string, outputFilePath string, outputFileName string){

	if (dataShards + parShards) > 256 {
		fmt.Fprintf(os.Stderr, "Error: sum of data and parity shards cannot exceed 256\n")
		os.Exit(1)
	}

	encodingStream, err := reedsolomon.NewStream(dataShards, parShards)
	checkErr(err)

	fmt.Println("Opening", inputFile)
	f, err := os.Open(inputFile)
	checkErr(err)

	instat, err := f.Stat()
	fmt.Print(instat)
	checkErr(err)

	shards := dataShards + parShards
	out := make([]*os.File, shards)

	// // Create the resulting files.
	// dir, file := filepath.Split(outputFilePath)
	// // if *outDir != "" {
	// // 	dir = *outDir
	// // }

	for i := range out {
		outfn := fmt.Sprintf("%s.%d", outputFileName, i)
		fmt.Println("Creating", outfn)
		out[i], err = os.Create(filepath.Join(outputFilePath, outfn))
		checkErr(err)
	}

	// Split into files.
	data := make([]io.Writer, dataShards)
	for i := range data {
		data[i] = out[i]
	}
	// Do the split
	err = encodingStream.Split(f, data, instat.Size())
	checkErr(err)

	// Close and re-open the files.
	input := make([]io.Reader, dataShards)

	for i := range data {
		out[i].Close()
		f, err := os.Open(out[i].Name())
		checkErr(err)
		input[i] = f
	//	updateManifestFile()
		defer f.Close()
	}

	// Create parity output writers
	parity := make([]io.Writer, parShards)
	for i := range parity {
		parity[i] = out[dataShards+i]
		defer out[dataShards+i].Close()
	}

	// Encode parity
	err = encodingStream.Encode(input, parity)
	checkErr(err)
	fmt.Printf("File split into %d data + %d parity shards.\n", dataShards, parShards)
    



}

func updateManifestFile(filePath string,nodeId string,fileName string,peerID string,fileHash []byte,fileIndex int){
        	var chunks File   
			dir, file := filepath.Split(filePath)
			fileExtension := filepath.Ext(filePath)
			chunks.Chunkname = file
			chunks.FilePath = dir
			chunks.FileExtension = fileExtension
			chunks.FileName = fileName
			chunks.Ownername = "StorageTeam"
			chunks.NodeAddress = peerID
			chunks.BlockHash = fileHash
			chunks.ChuckIndex = fileIndex
			chunks.Port = peerID
	
			SaveFileInfo(chunks)
}

//func()

func getLocalStorage(path string,fileName string){

	//return path.fileName
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(2)
	}
}

