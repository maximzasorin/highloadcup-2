package main

import (
	"fmt"
	"runtime"
)

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\nTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\nSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\nNumGC = %v\n", m.NumGC)
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
