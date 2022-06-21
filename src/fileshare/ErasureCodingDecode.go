package fileshare

import (
	"fmt"
	"io"
	"os"

	"github.com/klauspost/reedsolomon"
)

func erasureDecoding(dataShards int, parShards int, fname []File, outFile string) {

	// Create matrix
	enc, err := reedsolomon.NewStream(dataShards, parShards)
	checkErr(err)

	// Open the inputs
	shards, size, err := openInput(dataShards, parShards, fname)
	fmt.Println("\n\nshards---> ", shards)
	checkErr(err)

	// Verify the shards
	ok, err := enc.Verify(shards)
	if ok {
		fmt.Println("No reconstruction needed")
	} else {
		fmt.Println("Verification failed. Reconstructing data")
		shards, size, err = openInput(dataShards, parShards, fname)
		checkErr(err)
		// Create out destination writers
		out := make([]io.Writer, len(shards))
		for i := range out {
			if shards[i] == nil {
				// outfn := fmt.Sprintf("%s.%d", fname[i].FilePath)
				fmt.Println("\n\nCreating", fname[i].FilePath)
				out[i], err = os.Create(fname[i].FilePath)
				checkErr(err)
			}
		}
		err = enc.Reconstruct(shards, out)
		if err != nil {
			fmt.Println("Reconstruct failed -", err)
			os.Exit(1)
		}
		// Close output.
		for i := range out {
			if out[i] != nil {
				err := out[i].(*os.File).Close()
				checkErr(err)
			}
		}
		shards, size, err = openInput(dataShards, parShards, fname)
		ok, err = enc.Verify(shards)
		if !ok {
			fmt.Println("Verification failed after reconstruction, data likely corrupted:", err)
			os.Exit(1)
		}
		checkErr(err)
	}

	// Join the shards and write them
	outfn := outFile
	// if outfn == "" {
	// 	outfn = fname
	// }

	fmt.Println("Writing data to", outfn)
	f, err := os.Create(outfn)
	checkErr(err)

	shards, size, err = openInput(dataShards, parShards, fname)
	checkErr(err)

	// We don't know the exact filesize.
	err = enc.Join(f, shards, int64(dataShards)*size)
	checkErr(err)
}

func openInput(dataShards, parShards int, fname []File) (r []io.Reader, size int64, err error) {
	// Create shards and load the data.
	shards := make([]io.Reader, dataShards+parShards)

	for i, infn := range fname {
		f, err := os.Open(infn.FilePath)
		if err != nil {
			fmt.Println("Error reading file", err)
			shards[i] = nil
			continue
		} else {
			shards[i] = f
		}
		stat, err := f.Stat()
		checkErr(err)
		if stat.Size() > 0 {
			size = stat.Size()
		} else {
			shards[i] = nil
		}
	}
	return shards, size, nil
}
