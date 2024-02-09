package fileUtil

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
)

func ReadFile(filename string) ([]string, error) {
	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func CompareFiles(file1 string, file2 string) ([]string, []string) {
	lines1, err := ReadFile(file1)
	if err != nil {
		fmt.Println("Error reading file 1:", err)
		return nil, nil
	}

	lines2, err := ReadFile(file2)
	if err != nil {
		fmt.Println("Error reading file 2:", err)
		return nil, nil
	}

	fmt.Println("Lines unique to file 1:")
	var uniqueLines1 []string
	for _, line := range lines1 {
		found := false
		for _, line2 := range lines2 {
			if line == line2 {
				found = true
				break
			}
		}
		if !found {
			uniqueLines1 = append(uniqueLines1, line)
			fmt.Println(line)
		}
	}

	fmt.Println("\nLines unique to file 2:")
	var uniqueLines2 []string
	for _, line := range lines2 {
		found := false
		for _, line1 := range lines1 {
			if line == line1 {
				found = true
				break
			}
		}
		if !found {
			uniqueLines2 = append(uniqueLines2, line)
			fmt.Println(line)
		}
	}

	return uniqueLines1, uniqueLines2
}

func FindAllMatchingFilenames(filenames []string, pattern string) []string {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return nil
	}

	var matchingFilenames []string
	for _, filename := range filenames {
		if regex.MatchString(filename) {
			matchingFilenames = append(matchingFilenames, filename)
		}
	}
	slices.Sort(matchingFilenames)
	return matchingFilenames
}
