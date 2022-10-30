package main

import (
	"os"
	"syscall"
)

func getDeviceID(path string, fi os.FileInfo) devID {
	var stat = fi.Sys().(*syscall.Stat_t)
	return devID(stat.Dev)
}
