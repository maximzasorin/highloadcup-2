package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/pkg/errors"
)

func main() {
	var (
		dataset = flag.String("dataset", "/tmp/data", "Dataset")
		addr    = flag.String("addr", ":80", "Addr")
	)
	flag.Parse()

	setGCPercent(15)

	fmt.Println("Create parser")
	dicts := NewDicts()
	parser := NewParser(dicts)

	fmt.Println("Read options")
	now, rating := readOptions(*dataset + "/options.txt")

	fmt.Println("Create store")
	store := NewStore(dicts, now, rating)

	server := NewServer(store, parser, dicts, &ServerOptions{
		Addr: *addr,
	})

	go func() {
		for range time.Tick(5 * time.Second) {
			printAppStatus(store, server)
		}
	}()

	fmt.Println("Trying read archive")

	startTime := time.Now()
	readArchive(*dataset+"/data.zip", parser, store)
	// readDir(dataset+"/data", store)

	fmt.Println("Total accounts found =", store.Count(), "in", time.Now().Sub(startTime).Round(time.Millisecond))
	runtime.GC()

	fmt.Println("Create indexes")
	startIndex := time.Now()
	j := 1
	store.Iterate(func(account *Account) bool {
		store.AppendToIndex(account)
		if j%10000 == 0 {
			printAppStatus(store, server)
			fmt.Println(j)
		}
		j++
		return true
	})
	store.UpdateIndex()
	fmt.Println("Index created in", time.Now().Sub(startIndex).Round(time.Millisecond))

	// fmt.Println("Update country cities")
	// dicts.UpdateCountryCities(store)

	// a := store.Get(uint32(rand.Int31n(30000)))
	// fmt.Printf("Example = %+v, size = %db\n", a, unsafe.Sizeof(Account{}))

	// fmt.Println("Total with premium =", store.WithPremium())
	// fmt.Println("Count likes =", store.CountLikes())
	// fmt.Println("Count sex_f =", store.CountSexF())
	// fmt.Println("Total fnames =", len(dicts.GetFnames()))
	// fmt.Println("Total snames =", len(dicts.GetSnames()))
	// fmt.Println("Total countries =", len(dicts.GetCountries()))
	// fmt.Println("Total cities =", len(dicts.GetCities()))
	// fmt.Println("Total interests =", len(dicts.GetInterests()))

	printMemUsage()

	fmt.Println("Start server")

	// setGCPercent(25)

	fmt.Println("Run GC...")
	runtime.GC()

	err := server.Handle()
	if err != nil {
		log.Fatal(err)
	}
}

func readOptions(filename string) (uint32, bool) {
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
	ui64, err := strconv.ParseUint(strings.TrimSpace(line), 10, 32)
	if err != nil {
		log.Fatal(err)
	}

	line, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	rating, err := strconv.ParseUint(strings.TrimSpace(line), 10, 1)
	if err != nil {
		log.Fatal(err)
	}

	return uint32(ui64), rating == 1
}

func readArchive(filename string, parser *Parser, store *Store) {
	data, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

	fmt.Println("Founded files =", len(data.File))

	for _, f := range data.File {
		fmt.Println("parse file " + f.Name)

		reader, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		parseFile(reader, parser, store)
		reader.Close()
	}
}

func parseFile(reader io.ReadCloser, parser *Parser, store *Store) {
	rawAccounts, err := parser.DecodeAccounts(reader)
	if err != nil {
		log.Fatal("Cannot parse file")
	}

	for _, rawAccount := range rawAccounts {
		_, err := store.Add(rawAccount, false, false)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Can not add account"))
		}
	}
}

func printAppStatus(store *Store, server *Server) {
	fmt.Println("------------------------------------")
	fmt.Println("Total accounts =", store.Count())
	server.stats.Sort()
	fmt.Println(server.stats.Format())
	printMemUsage()
	fmt.Println("------------------------------------")
}

func setGCPercent(gcPercent int) {
	fmt.Println("SetGCPercent =", gcPercent)
	debug.SetGCPercent(gcPercent)
}
