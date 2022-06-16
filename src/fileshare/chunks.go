package fileshare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

const (
	dbPath            = "./database"
	cryptoKey         = "teteteteteetesdsdsdsdsdt"
	chunkFileSize int = 256 // bytes
	BUFFERSIZE    int = 500
)

var Colors = fileExtentions()

func fileExtentions() *fileExtention {
	return &fileExtention{
		txt:  ".txt",
		pdf:  ".pdf",
		docx: ".docx",
	}
}

type fileExtention struct {
	txt  string
	pdf  string
	docx string
}

func ReadDir(dirname string) []os.FileInfo {

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func ReadFile(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	file.Read(buffer)
	return buffer
}

func GetEncryptedFiles(fileName string, ownername string) []File {

	var chunks []File

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		if string(c.Ownername) == ownername && strings.HasPrefix(c.FileName, fileName) {
			chunks = append(chunks, c)
		}
	}
	return chunks
}

func ConvertDecryptFiles(fileName string, ownername string) string {

	chunks := GetEncryptedFiles(fileName, ownername)
	sort.SliceStable(chunks, func(i, j int) bool {
		return chunks[i].ChuckIndex < chunks[j].ChuckIndex
	})
	allFilePaths := make([]string, 0)
	for _, elem := range chunks {
		allFilePaths = append(allFilePaths, string(elem.FilePath))
	}
	allBufferData := RetrieveFilesFromChunk(allFilePaths)

	tempfile := "./testdirs/" + "final" + chunks[0].FileExtension

	file, err := os.Create(tempfile)
	if err != nil {
		log.Fatal(err)
	}

	file.Write(allBufferData)

	defer file.Close()
	return chunks[0].FileExtension
}

func CreateChunksAndEncrypt(filepath string, m *SwarmMaster, name string, fileExtension string) {

	writefile(CreateFileChunks(filepath), filepath, m, name, fileExtension)
}

func writefile(data [][]byte, filePath string, m *SwarmMaster, name string, fileExtension string) {

	nodes := m.GetActiveNodes()

	nodesLen := len(nodes)
	counter := 0
	for index, chunk := range data {

		// IF NUMBER OF CHUNKS ARE MORE THAN NUMBER OF CONNECTED PEERS
		if (counter == nodesLen) || (counter > nodesLen) {
			counter = 0
		}

		fileChunk := name + "-" + strconv.Itoa(index) + "-chunks-" + uuid.New().String() + fileExtension
		path := "../main/testdirs/peer" + strconv.Itoa(registerPeers[counter].PeerID) + "/" + fileChunk
		file, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		registerPeers[counter].RegisterFile(name)

		// MAINTAIN MENIFEST FILE
		var chunks File

		chunks.Chunkname = fileChunk
		chunks.FilePath = path
		chunks.FileExtension = fileExtension
		chunks.FileName = name
		chunks.Ownername = "StorageTeam"
		chunks.NodeAddress = strconv.Itoa(registerPeers[counter].PeerID)
		chunks.BlockHash = []byte("SomeHash")
		chunks.ChuckIndex = index
		chunks.Port = registerPeers[counter].Port

		SaveFileInfo(chunks)

		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		file.Write([]byte(EncryptFile(string(chunk))))
		counter++
	}

	// upload manifest file to all nodes
	updateManifest(m)
}

func updateManifest(m *SwarmMaster) {
	nodes := m.GetActiveNodes()
	for _, b := range nodes {
		destination := "testdirs/peer" + strconv.Itoa(b) + "/output.json"
		CopyManifest(destination)
	}
}

func CopyManifest(dst string) error {

	src := "output.json"

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	//remove file if already file exist
	_, err = os.Stat(dst)
	if err == nil {
		err := os.Remove(dst)
		if err != nil {
			log.Fatal(err)
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}

func SearchFiles(filename string, ownername string) []File {

	chunks := GetChunksByPrefix(filename, ownername)
	return chunks
}

type File struct {
	Chunkname     string
	FilePath      string
	Ownername     string
	FileName      string
	FileExtension string
	NodeAddress   string
	BlockHash     []byte
	ChuckIndex    int
	Port          string
}

type FileDB struct {
	Database *badger.DB
}

func GetDBinstacnce() *FileDB {
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	blockchain := FileDB{Database: db}
	return &blockchain
}

func SaveFileInfo(chunk File) []File {

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(chunk)
	// key := chunk.Chunkname

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)
	data = append(data, chunk)

	reqBodyBytes2 := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes2).Encode(data)
	ioutil.WriteFile("output.json", reqBodyBytes2.Bytes(), 0644)

	return data
}

func GetChunkByKey(key string) string {

	var finalResponse []byte
	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		newKey := string(c.Chunkname)
		if newKey == key {
			reqBodyBytes2 := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes2).Encode(c)
			finalResponse = reqBodyBytes2.Bytes()
		}
	}
	if len(finalResponse) == 0 {
		return string(`{"message": "No data found"}`)
	}
	return string(finalResponse)
}

func GetChunksByPrefix(prefix string, ownername string) []File {

	var chunks []File

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		if string(c.Ownername) == ownername && strings.HasPrefix(c.FileName, prefix) {
			chunks = append(chunks, c)
		}
	}
	return chunks
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func isSameSentence(t1, t2 pdf.Text) bool {
	if t1.Font == t2.Font && t1.FontSize == t2.FontSize {
		return true
	}
	return false
}

func readPdf2(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	defer f.Close()
	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		var lastTextStyle pdf.Text
		texts := p.Content().Text
		for _, text := range texts {
			if isSameSentence(text, lastTextStyle) {
				lastTextStyle.S = lastTextStyle.S + text.S
			} else {
				fmt.Printf("Font: %s, Font-size: %f, x: %f, y: %f, content: %s \n", lastTextStyle.Font, lastTextStyle.FontSize, lastTextStyle.X, lastTextStyle.Y, lastTextStyle.S)
				lastTextStyle = text
			}
		}
	}
	return "", nil
}
