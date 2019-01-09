package main

// /*
// #include <unistd.h>
// */
// import "C"

import (
	"fmt"
	"runtime"
)

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	// fmt.Printf("Rss = %v MiB\n", bToMb(uint64(getRSSMemory())))
	fmt.Printf("Alloc = %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("TotalAlloc = %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("Sys = %v MiB\n", bToMb(m.Sys))
	fmt.Printf("NumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// func getRSSMemory() (rss int64) {
// 	fd, err := os.Open(`/proc/self/statm`)
// 	if err != nil {
// 		return 0
// 	}
// 	defer fd.Close()

// 	var tmp int64

// 	if _, err := fmt.Fscanf(fd, `%d %d`, &tmp, &rss); err != nil {
// 		return 0
// 	}

// 	pagesize := int64(C.sysconf(C._SC_PAGESIZE))

// 	rss *= pagesize

// 	return
// }
