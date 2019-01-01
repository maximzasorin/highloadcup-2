package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/francoispqt/gojay"
)

func main() {
	dataset := os.Getenv("DATASET")
	// port := os.Getenv("PORT")

	fmt.Println("Create store")
	store := NewStore()

	fmt.Println("Trying read archive")

	startTime := time.Now()
	readArchive(dataset+"/data.zip", store)
	// readDir(dataset+"/data", store)

	fmt.Println("Total accounts found =", store.Count(), "in", time.Now().Sub(startTime).Round(time.Millisecond))

	fmt.Println("Read options")
	readOptions(dataset+"/options.txt", store)

	printMemUsage()

	fmt.Println("Run GC...")
	runtime.GC()

	printMemUsage()

	a := store.Get(uint32(rand.Int31n(30000)))
	fmt.Printf("Example = %+v, size = %vb\n", a, unsafe.Sizeof(*a))

	// printMemUsage()

	// fmt.Println("Start server")

	// server := NewServer(store, &ServerOptions{
	// 	Addr: ":" + port,
	// })
	// err := server.Handle()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func readOptions(filename string, store *Store) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	now, err := strconv.ParseUint(strings.TrimSpace(line), 10, 32)
	if err != nil {
		log.Fatal(err)
	}

	store.SetNow(uint32(now))
}

func readArchive(filename string, store *Store) {
	data, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

	fmt.Println("Founded files", len(data.File))

	for _, f := range data.File {
		fmt.Println("Parse file", f.Name)

		reader, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		parseFile(reader, store)

		reader.Close()
	}
}

func parseFile(reader io.ReadCloser, store *Store) {
	// gojay
	dec := gojay.NewDecoder(reader)
	err := dec.Object(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "accounts":
			return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
				// start := time.Now()
				account := Account{}
				err := dec.Object(&account)
				if err != nil {
					return err
				}
				err = store.Add(&account)
				if err != nil {
					log.Fatal("Can not add account to store")
				}
				// fmt.Println("Account", account.ID, "was parsed for", time.Now().Sub(start))
				// printMemUsage()
				return nil
			}))
		}
		return errors.New("Unknown key in accounts file")
	}))

	if err != nil {
		log.Fatal(err)
	}
}
