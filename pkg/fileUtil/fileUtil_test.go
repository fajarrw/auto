package fileUtil

import "testing"

func TestCompareFiles(t *testing.T) {
	uniqueLines1, uniqueLines2 := CompareFiles("file1.txt", "file2.txt")
	if uniqueLines1[0] != "a" {
		t.Errorf("CompareFiles(file1.txt, file2.txt) returned uniqueLines1: %s, expected %s", uniqueLines1[0], "a")
	}
	if uniqueLines2[0] != "d" {
		t.Errorf("CompareFiles(file1.txt, file2.txt) returned uniqueLines1: %s, expected %s", uniqueLines2[0], "d")
	}
}

func TestFindAllMatchingFilenames()
