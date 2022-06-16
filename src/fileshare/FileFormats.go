package fileshare

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ReadPdf(path string) []string {

	allChunks := make([]string, 0)
	f, r, err := pdf.Open(path)
	Handle(err)
	defer f.Close()
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	Handle(err)

	buf.ReadFrom(b)
	data := buf.String()

	info, err := os.Stat(path)
	Handle(err)

	var m int64 = info.Size()
	nSize := int(m)
	chunkSize := nSize / 10000 // 10 KB per chunk
	lengthPerSplit := len(data) / chunkSize

	for i := 0; i < chunkSize; i++ {
		if i+1 == chunkSize {
			chunkTexts := data[i*lengthPerSplit:]
			allChunks = append(allChunks, chunkTexts)
		} else {
			chunkTexts := data[i*lengthPerSplit : (i+1)*lengthPerSplit]
			allChunks = append(allChunks, chunkTexts)
		}
	}
	return allChunks
}

func ReadTxt(filepath string) []string {
	allChunks := make([]string, 0)

	file, err := os.Open(filepath)
	Handle(err)
	defer file.Close()

	info, err := os.Stat(filepath)
	Handle(err)
	chunkSize := (int(info.Size()) / chunkFileSize) + 1

	scanner := bufio.NewScanner(file)
	texts := make([]string, 0)
	for scanner.Scan() {
		text := scanner.Text()
		texts = append(texts, text)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	lengthPerSplit := len(texts) / chunkSize

	for i := 0; i < chunkSize; i++ {
		if i+1 == chunkSize {
			chunkTexts := texts[i*lengthPerSplit:]
			allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
		} else {
			chunkTexts := texts[i*lengthPerSplit : (i+1)*lengthPerSplit]
			allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
		}
	}
	return allChunks
}
