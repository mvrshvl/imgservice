package main

import (
	"bufio"
	"fmt"
	"log"
	"nir/geth"
	"os"
	"strings"
)

func main() {
	// open the file
	file, err := os.Open(geth.StaticDirectory + "/exchanges.csv")

	//handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
	}

	fileWriting, err := os.Create(geth.StaticDirectory + "/exchanges2.csv")
	//handle errors while opening
	if err != nil {
		log.Fatalf("Error when creating file: %s", err)
	}

	fileScanner := bufio.NewScanner(file)

	// read line by line
	for fileScanner.Scan() {
		fmt.Println(fileScanner.Text())
		line := strings.Split(fileScanner.Text(), ",")
		_, err := fileWriting.WriteString(fmt.Sprintf("%s, %s\n", line[0], line[1]))
		if err != nil {
			log.Fatal(err)
		}
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
	}

	file.Close()
}
