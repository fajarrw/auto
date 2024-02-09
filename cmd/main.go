package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fajarrw/auto/pkg/fileUtil"
	"github.com/fajarrw/auto/pkg/terminal"
)

func folderExists(folderPath string) bool {
	_, err := os.Stat(folderPath)
	return !os.IsNotExist(err)
}

func createFolderIfNotExists(folderPath string) error {
	if !folderExists(folderPath) {
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}
		fmt.Printf("[INFO] Folder '%s' created.\n", folderPath)
	}
	return nil
}

func scanSubdomain(folderPath string, domainName string) {
	// Step 2: Run amass enum
	terminal.RunCommand(folderPath, "amass", "enum", "-passive", "-norecursive", "-noalts", "-d", domainName, "-o", filepath.Join(folderPath, "amass.txt"))

	// Step 3: Run crtsh
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("crtsh -q %%.%s -o | sort | uniq | anew %s", domainName, filepath.Join(folderPath, "crtsh.txt")))

	// Step 4: Run haktrails subdomains
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("echo %s | haktrails subdomains > %s", domainName, filepath.Join(folderPath, "security-trails.txt")))

	// Step 5: Run subfinder
	terminal.RunCommand(folderPath, "subfinder", "-d", domainName, "-o", filepath.Join(folderPath, "subfinder.txt"))

	// Step 6: Combine and filter unique subdomains
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("cat amass.txt crtsh.txt security-trails.txt subfinder.txt | sed 's/\\*.//g' | sort | uniq > %s", filepath.Join(folderPath, "sub.txt")))

	// Step 7: Run httpx to check live subdomains
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("httpx -l %s | sort | anew %s", filepath.Join(folderPath, "sub.txt"), filepath.Join(folderPath, "live_sub.txt")))

	// Step 8: Sort the live_sub.txt
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("cat live_sub.txt | sort | anew live_sub_%s.txt", time.Now().Format("01_02_15_04")))

	// Step 9: Delete live_sub.txt
	fmt.Println("[INFO] Deleting live_sub.txt")
	if _, err := os.Stat(filepath.Join(folderPath, "live_sub.txt")); err == nil {
		// File exists, so remove it
		err := os.Remove(filepath.Join(folderPath, "live_sub.txt"))
		if err != nil {
			fmt.Println("[ERR] Error removing file:", err)
			return
		}
		fmt.Println("[INFO] live_sub.txt removed successfully.")
	} else if os.IsNotExist(err) {
		// File doesn't exist
		fmt.Println("[INFO] live_sub.txt does not exist. So, it is not removed")
	} else {
		// Other error occurred
		fmt.Println("[ERR] Error checking live_sub.txt existence:", err)
	}

	fmt.Println("[DONE] Process completed for domain:", domainName)
}

func aggregateLines(folderPath string, filename1 string, filename2 string) {
	terminal.RunCommand(folderPath, "bash", "-c", fmt.Sprintf("sort -u %s %s | anew %s", filename1, filename2, filepath.Join(folderPath, fmt.Sprintf("agg_%s_%s.txt", strings.TrimSuffix(filename1, ".txt"), strings.TrimSuffix(filename2, ".txt")))))
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[ERR] Error getting user's home directory:", err)
		return
	}

	// Open the text file containing domain names
	file, err := os.Open(".secret/domains.txt")
	if err != nil {
		fmt.Println("[ERR] Error opening domains.txt:", err)
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read each line from the file
	for scanner.Scan() {
		// Extract domain name from the line
		domainName := strings.TrimSpace(scanner.Text())

		// Create folder for the domain if it doesn't exist
		folderPath := filepath.Join(homeDir+"/Desktop", domainName)
		if err := createFolderIfNotExists(folderPath); err != nil {
			fmt.Printf("[ERR] Error creating folder '%s': %v\n", folderPath, err)
			continue
		}

		// Search for valid subdomain
		scanSubdomain(folderPath, domainName)

		// Aggregate newly generated file with the older one
		// Open the folder
		folder1, err := os.Open(folderPath)
		if err != nil {
			fmt.Println("[ERR] Error opening folder: ", err)
			return
		}
		defer folder1.Close()

		fileInfos1, err := folder1.Readdir(-1)
		if err != nil {
			fmt.Println("[ERR] Error reading folder contents: ", err)
			return
		}

		// Iterate over the directory entries
		var filenames1 []string
		fmt.Println("[INFO] Filenames in folder:")
		for _, fileInfo := range fileInfos1 {
			filenames1 = append(filenames1, fileInfo.Name())
		}
		slices.Sort(filenames1)
		fmt.Println(filenames1)

		// Define the regular expression pattern
		liveSubPattern := `^live_sub.*`
		liveSubFilenames := fileUtil.FindAllMatchingFilenames(filenames1, liveSubPattern)
		fmt.Println("[INFO] Matching live_sub filenames:")
		fmt.Println(liveSubFilenames)

		aggPattern := `^agg`
		aggFilenames1 := fileUtil.FindAllMatchingFilenames(filenames1, aggPattern)
		fmt.Println("[INFO] Matching agg filenames:")
		fmt.Println(aggFilenames1)

		for i := 0; i < len(liveSubFilenames)-1; i++ {
			found := false
			for _, element := range aggFilenames1 {
				if element == fmt.Sprintf("agg_%s_%s.txt", strings.TrimSuffix(liveSubFilenames[i], ".txt"), strings.TrimSuffix(liveSubFilenames[i+1], ".txt")) {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("[INFO] agg_%s_%s.txt already exists. Thus, it is skipped.\n", strings.TrimSuffix(liveSubFilenames[i], ".txt"), strings.TrimSuffix(liveSubFilenames[i+1], ".txt"))
				continue
			}
			fmt.Println("[INFO] Creating aggregate file")
			aggregateLines(folderPath, liveSubFilenames[i], liveSubFilenames[i+1])
		}

		folder2, err := os.Open(folderPath)
		if err != nil {
			fmt.Println("[ERR] Error opening folder: ", err)
			return
		}
		defer folder2.Close()

		fileInfos2, err := folder2.Readdir(-1)
		if err != nil {
			fmt.Println("[ERR] Error reading folder contents:", err)
			return
		}

		// Iterate over the UPDATED directory entries
		var filenames2 []string
		fmt.Println("[INFO] Filenames in folder (UPDATED):")
		for _, fileInfo := range fileInfos2 {
			filenames2 = append(filenames2, fileInfo.Name())
		}
		slices.Sort(filenames2)
		fmt.Println(filenames2)

		//aggPattern := `^agg`
		aggFilenames2 := fileUtil.FindAllMatchingFilenames(filenames2, aggPattern)
		fmt.Println("[INFO] Matching agg filenames:")
		fmt.Println(aggFilenames2)

		// Check for difference between newly generated live)sub.txt and the older one
		for i := 0; i < len(aggFilenames2)-1; i++ {
			uniqueLines1, uniqueLines2 := fileUtil.CompareFiles(aggFilenames2[i], aggFilenames2[i+1])

			file1, err := os.Create(folderPath + "/" + fmt.Sprintf("diff1_%s_%s.txt", aggFilenames2[i], aggFilenames2[i+1]))
			if err != nil {
				fmt.Println("[ERR] Error creating file:", err)
				return
			}
			defer file1.Close()

			// Write each string to the file, separated by newline
			for _, str := range uniqueLines1 {
				_, err := fmt.Fprintln(file1, str)
				if err != nil {
					fmt.Println("[ERR] Error writing to file:", err)
					return
				}
			}

			file2, err := os.Create(folderPath + "/" + fmt.Sprintf("diff2_%s_%s.txt", aggFilenames2[i], aggFilenames2[i+1]))
			if err != nil {
				fmt.Println("[ERR] Error creating file:", err)
				return
			}
			defer file2.Close()

			// Write each string to the file, separated by newline
			for _, str := range uniqueLines2 {
				_, err := fmt.Fprintln(file2, str)
				if err != nil {
					fmt.Println("[ERR] Error writing to file:", err)
					return
				}
			}

			fmt.Println("[INFO] File diff2 created successfully.")
		}
	}

	// Check for any errors encountered during reading the file
	if err := scanner.Err(); err != nil {
		fmt.Println("[ERR] Error reading file:", err)
		return
	}
}
