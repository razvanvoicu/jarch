package main

import (
	"container/list"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var defaultPhotoDir = "/media/r/"

var samsungDVPhotoAndVideoPattern = "/DCIM/[0-9]+PHOTO/SAM_"
var sonyPhotoPattern = "/DCIM/[0-9]+MSDCF/DSC[0-9]+[.]JPG"
var sonyVideoPattern = "/MP_ROOT/[0-9]+ANV[0-9]+/MAH[0-9]+[.]MP4"
var ekenPhotoPattern = "/PHOTO/FHD[0-9]+[.]JPG"
var ekenVideoPattern = "/VIDEO/FHD[0-9]+[.]MOV"
var djiPattern = "/DCIM/[0-9]+MEDIA/DJI_[0-9]+[.]((MP4)|(JPG))"

var patterns = []string {
	samsungDVPhotoAndVideoPattern,
	sonyPhotoPattern,
	sonyVideoPattern,
	ekenPhotoPattern,
	ekenVideoPattern,
	djiPattern,
}

func main() {
	searchDir := photoDir()
	photoPaths := list.New()
	for p := range patterns {
		searchFiles(searchDir, patterns[p], photoPaths, detectorWalker)
	}

	l := photoPaths.Len()
	for i, e := 0, photoPaths.Front(); e != nil; i, e = i+1, e.Next() {
		fmt.Println("Copying", e.Value, "to", dstName(e.Value.(string)), "(", i+1, "of", l, ")")
		fileCopy(e.Value.(string))
	}
}

func photoDir() string {
	dir := os.Getenv("JOURNAL_PHOTO_PATH")
	if dir == "" {
		dir = defaultPhotoDir
	}
	return dir
}

func searchFiles(dir string, pattern string, paths *list.List, walker func(string, *list.List) filepath.WalkFunc) {
	filepath.Walk(dir, walker(pattern, paths))
}

func detectorWalker(pattern string, path_holder *list.List) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		r, _ := regexp.MatchString(pattern, path)
		if r {
			path_holder.PushBack(path)
		}
		return nil
	}
}

func fileCopy(src string) (int64, error) {
	dst := dstName(src)
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func dstName(src string) string {
	dir := os.Getenv("JOURNAL_PATH")
	if dir == "" {
		dir = "/mnt/odin/data/journal/"
	}
	info, _ := os.Stat(src)
	ts := info.ModTime().UTC()
	year := ts.Format("2006")
	month := ts.Format("01")
	day := ts.Format("02")
	time := ts.Format("150405")
	suffix := strings.Split(src, ".")[1]
	os.MkdirAll(dir+year+"/"+month+"/"+day, 066)
	return dir + year + "/" + month + "/" + day + "/" + year + month + day + "-" + time + "." + suffix
}
