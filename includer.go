package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

func includeStuff() {

	inputFile, err := os.Open("client.html")
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create("client_out.html")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	regex := regexp.MustCompile("#include \"(.*)\"")

	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "#include") {
			filePath := regex.FindStringSubmatch(scanner.Text())[1]
			includeFile, includeErr := os.Open(filePath)
			if includeErr != nil {
				log.Fatal(includeErr)
			}
			defer includeFile.Close()
			includeScanner := bufio.NewScanner(includeFile)
			for includeScanner.Scan() {
				outputFile.WriteString(includeScanner.Text() + "\n")
			}
			includeFile.Close()
		} else {
			outputFile.WriteString(scanner.Text() + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
