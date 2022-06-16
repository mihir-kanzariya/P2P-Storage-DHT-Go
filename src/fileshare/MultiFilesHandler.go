package fileshare

import (
	"fmt"
	"os"
)

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

	divided := chunks(buffer, 10192)
	fmt.Println("Total ", len(divided), " chunks created")
	return divided
}

func RetrieveFilesFromChunk(allFiles []string) []byte {
	allBuffer := make([]byte, 0)

	for _, element := range allFiles {
		file, err := os.Open(element)
		if err != nil {
			fmt.Println(err)
			continue
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
		allBuffer = append(allBuffer, []byte(DecryptFile(string(buffer)))...)
	}
	return allBuffer
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
