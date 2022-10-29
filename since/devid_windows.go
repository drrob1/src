package main

import (
	"os"
)

//func getDeviceID(path string, fi os.FileInfo) devID {
//	var stat = fi.Sys().(*syscall.Stat_t)
//	return devID(stat.Dev)
//}

// getDeviceID will always return 0 on Windows, so it will compile but have no effect.
func getDeviceID(path string, fi os.FileInfo) devID {
	return 0
}
