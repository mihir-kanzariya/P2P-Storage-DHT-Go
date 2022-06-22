package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	rann "math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var mainMenu = `
1) Enter the filename to store
2) Enter the filename to retrieve
3) Exit
`

var hasher = fnv.New32a()
var ringCapacity uint32 = 127

// Returns the id of a node (given its full address) or key of a file (given its name).
func hsh(in string) int {
	hasher.Write([]byte(in))
	digest := hasher.Sum32()
	hasher.Reset()
	return int(digest % ringCapacity)
}

// Connects to the peer at the given address.
func connectToPeer(address string) (net.Conn, *bufio.Reader) {
	address = strings.TrimSpace(address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("Could not connect to the peer.")
		log.Fatalln(err)
	}
	// Create a buffered reader.
	reader := bufio.NewReader(conn)
	return conn, reader
}

// The server sends responses in the following form:
// OK/ERR <msg>\n
// This method returns these separately(e.g. "ERR No file found\n" => "ERR", "No file found.")
func extractServerResponse(resp string) (string, string) {
	resp = strings.TrimSpace(resp)
	var prefix string
	var msg string
	if strings.HasPrefix(resp, "OK") {
		prefix = "OK"
		if len(resp) > 2 {
			msg = resp[3:]
		}
	} else if strings.HasPrefix(resp, "ERR") {
		prefix = "ERR"
		if len(resp) > 3 {
			msg = resp[4:]
		}
	}
	return prefix, msg
}

// Constructs a store request with the file name to store, then sends the file.
// (1) finds the successor (owner) of the file through the given peer.
// (2) uploads the file to the owner of the file.
func storeFile(fileName string, peerAddr string) {
	// Find the successor (owner) of the file.
	fileKey := hsh(fileName)
	succAddr := askForSuccesor(fileKey, peerAddr)
	// Begin trying to store the file on the successor.
	conn, reader := connectToPeer(succAddr)
	defer conn.Close()
	srcFile, err := os.Open(fileName)
	defer srcFile.Close()
	if err != nil {
		log.Println("Could not send store request")
		log.Fatalln(err)
	}
	fileInfo, _ := srcFile.Stat()
	fileSize := fileInfo.Size()
	// Send the store request.
	storeRequest := fmt.Sprintf("STORE %s %d\n", fileName, fileSize)
	conn.Write([]byte(storeRequest))
	// Read the response.
	serverResponse, _ := reader.ReadString('\n')
	respType, respMsg := extractServerResponse(serverResponse)
	// Response: ERR <error msg>
	if respType != "OK" {
		fmt.Println("> Server response:", respMsg)
		return
	}
	// Response: OK
	io.Copy(conn, srcFile)
	// Response: OK
	fmt.Println("File successfully stored.")
}

// Retrieves the given file from the peer.
// (1) finding the successor of the file through the peer.
// (2) downloading the file through that successor.
func retrieveFile(fileName string, peerAddr string) {
	// Find the successor (owner) of the file.
	fileKey := hsh(fileName)
	succAddr := askForSuccesor(fileKey, peerAddr)
	// Begin trying to retrieve the file.
	conn, reader := connectToPeer(succAddr)
	defer conn.Close()
	// Construct the request.
	retrieveRequest := fmt.Sprintf("RETRIEVE %s %s\n", fileName, succAddr)
	// Send the retrieve request.
	conn.Write([]byte(retrieveRequest))
	// Retrieve the size of the file from the connection.
	serverResponse, _ := reader.ReadString('\n')
	respType, respMsg := extractServerResponse(serverResponse)
	// Response: ERR <error msg>
	if respType != "OK" {
		fmt.Println("> Server response:", respMsg)
		return
	}
	// Response: OK <file size>
	fileSize, _ := strconv.Atoi(strings.TrimSpace(respMsg))
	// Create the local file.
	dstFile, _ := os.Create(fileName)
	defer dstFile.Close()
	// Retrieve the file from the connection.
	io.CopyN(dstFile, reader, int64(fileSize))
	// Read the next response.
	serverResponse, _ = reader.ReadString('\n')
	respType, respMsg = extractServerResponse(serverResponse)
	// Response: ERR <error msg>
	if respType != "OK" {
		fmt.Println("> Server response:", respMsg)
		return
	}
	// Response: OK
	fmt.Println("File retrieved successfully.")
}

// Constructs a successor request with the given id and sends it to the given address.
// Returns the answer to the request (i.e. the address of the successor).
// SUCC <id> => <succ addr>
func askForSuccesor(id int, peerAddr string) string {
	// Initiate a connection with the given peer address.
	conn, reader := connectToPeer(peerAddr)
	defer conn.Close()
	// Send the successor request.
	succRequest := fmt.Sprintf("SUCC %d\n", id)
	conn.Write([]byte(succRequest))
	// Wait for an answer.
	answer, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Could not get the successor.")
		log.Fatalln(err)
	}
	// The answer will only contain the address of the successor.
	return answer
}

func saveFile(fileName string, peerAddr string) {

	chunks := CreateFileChunks(fileName)

	for index, chunk := range chunks {

		fileExtension := filepath.Ext(fileName)
		name := "chunk-" + strconv.Itoa(rann.Int()) + strconv.Itoa(index) + fileExtension
		tempFile, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		tempFile.Write([]byte(EncryptFile(string(chunk))))
		storeFile(tempFile.Name(), peerAddr)
		os.Remove(name)
	}
}
func main() {
	storeIP := os.Args[1]
	storePort := os.Args[2]
	storeAddr := storeIP + ":" + storePort
	// Show the main menu.
	fmt.Println(mainMenu)
	for {
		// Ask the user for a selection.
		fmt.Print("> Please select an option: ")
		var input string
		fmt.Scanln(&input)
		selectedOption, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid choice.")
			continue
		}
		// Act accordingly.
		switch selectedOption {
		case 1:
			// Ask the filename to hash.
			fmt.Print("> Enter the file name to store: ")
			var fileName string
			fmt.Scanln(&fileName)
			start := time.Now()
			saveFile(fileName, storeAddr)
			// storeFile(fileName, storeAddr)
			elapsed := time.Since(start)
			fmt.Println("Transfer took", elapsed.Microseconds(), "us")
		case 2:
			// Ask the filename to hash.
			fmt.Print("> Enter the file name to retrieve: ")
			var fileName string
			fmt.Scanln(&fileName)
			start := time.Now()
			retrieveFile(fileName, storeAddr)
			elapsed := time.Since(start)
			fmt.Println("Transfer took", elapsed.Microseconds(), "us")
		case 3:
			fmt.Println("Goodbye!")
			return
		}
	}
}

//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//
//------------------------------------------------ CREATING CHUNKS------------------------------------------//

func CreateFileChunks(pathName string) [][]byte {
	file, err := os.Open(pathName)
	if err != nil {
		fmt.Println(err)
		return nil
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

	divided := chunks(buffer, 60)
	fmt.Println("Total ", len(divided), " chunks created")
	return divided
}

func chunks(xs []byte, chunkSize int) [][]byte {
	if len(xs) == 0 {
		return nil
	}
	divided := make([][]byte, (len(xs)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(xs) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = xs[prev:next]
		prev = next
		i++
	}
	divided[i] = xs[prev:]
	return divided
}

//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION ------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//
//------------------------------------------------ CREATING ENCRYPTIONS AND DECRYPTION------------------------------------------//

const (
	cryptoKey = "teteteteteetesdsdsdsdsdt"
)

func DecryptFile(cipherstring string) string {

	keystring := cryptoKey
	ciphertext := []byte(cipherstring)
	key := []byte(keystring)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("Text is too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext)
}

func EncryptFile(plainstring string) string {

	keystring := cryptoKey
	plaintext := []byte(plainstring)
	key := []byte(keystring)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)

	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return string(ciphertext)
}
