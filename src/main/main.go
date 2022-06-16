package main

import (
	"FileShare/src/fileshare"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

var m *fileshare.SwarmMaster
var r *mux.Router

func pingFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(`Pong`)
}

func createPeer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nEndpoint Hit: CreatePeer")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter id in order to create new Peer")
	}

	resBytes := []byte(reqBody)        // Converting the string "res" into byte array
	var jsonRes map[string]interface{} // declaring a map for key names as string and values as interface
	_ = json.Unmarshal(resBytes, &jsonRes)

	var id int = int(jsonRes["id"].(float64))
	testDirectory := "testdirs/peer" + strconv.Itoa(id)
	const nodeIdPrefix = 60120
	port := ":" + strconv.Itoa(nodeIdPrefix+id)

	os.MkdirAll(testDirectory, 0777)

	p1 := fileshare.MakePeer(id, testDirectory, port)
	nodes := m.GetActiveNodes()
	if len(nodes) > 0 {
		randomNodeId := rand.Int() % len(nodes)
		fmt.Println("Node ", id, " will connect with ", nodes[randomNodeId])
		p1.ConnectPeer(":"+strconv.Itoa(nodeIdPrefix+nodes[randomNodeId]), nodes[randomNodeId])
	}
	p1.ConnectServer()

	destination := testDirectory + "/output.json"
	fileshare.CopyManifest(destination)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p1)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10)

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	defer file.Close()

	fileExtension := filepath.Ext(handler.Filename)
	pattern := "file.*" + fileExtension
	tempFile, err := ioutil.TempFile("", pattern)

	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	fileshare.CreateChunksAndEncrypt(tempFile.Name(), m, handler.Filename, fileExtension)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(`Successfully Uploaded File`)
}

func getActiveNodes(w http.ResponseWriter, r *http.Request) {
	nodes := m.GetActiveNodes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func getChunkByKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	setupHeader(w)
	data := fileshare.GetChunkByKey(key)
	w.Write([]byte(data))
}

func setupHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	m = fileshare.MakeSwarmMaster()
	m.MasterTest() // this isn't really needed - we can move the code to
	setupRoutes()
}

func searchFile(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	ownername := r.URL.Query().Get("ownername")

	setupHeader(w)
	files := fileshare.SearchFiles(filename, ownername)
	json.NewEncoder(w).Encode(files)
}

func decryptFile(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	ownername := r.URL.Query().Get("ownername")

	setupHeader(w)

	fileExtension := fileshare.ConvertDecryptFiles(filename, ownername)

	tempFileName := "final" + fileExtension
	files := fileshare.ReadFile("./testdirs/" + tempFileName)
	w.Write(files)
	// os.Remove("./testdirs/" + tempFileName)
}

func setupRoutes() {
	r = mux.NewRouter()

	r.HandleFunc("/ping", pingFunc).Methods("GET")
	r.HandleFunc("/getActivePeers", getActiveNodes).Methods("GET")
	r.HandleFunc("/upload", upload).Methods("POST")
	r.HandleFunc("/createPeer", createPeer).Methods("POST")
	r.HandleFunc("/getChunkByKey/{key}", getChunkByKey).Methods("GET")
	r.HandleFunc("/searchFile/{filename}", searchFile).Methods("GET")
	r.HandleFunc("/decryptFile/{filename}", decryptFile).Methods("GET")

	log.Fatal(http.ListenAndServe(":5001", r))
}
