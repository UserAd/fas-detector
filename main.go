package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := os.Stdin
	writer := os.Stdout

	defer reader.Close()
	defer writer.Close()

	vars := make(map[string]string)

	s := bufio.NewScanner(reader)
	for s.Scan() {
		if s.Text() == "" {
			break
		}

		terms := strings.SplitN(s.Text(), ":", 2)
		if len(terms) == 2 {
			vars[strings.TrimSpace(terms[0])] = strings.TrimSpace(terms[1])
		}
	}

	filename, ok := vars["agi_arg_1"]

	if !ok {
		writer.WriteString("VERBOSE Filename ARG not sent\n")
		return
	}

	file, err := os.Open(fmt.Sprintf("/tmp/%s.wav", filename))

	if err != nil {
		writer.WriteString(fmt.Sprintf("VERBOSE \"Can not open file %s\"\n", filename))
		return
	}

	defer file.Close()

	detector := NewDetector(file)

	_, detected, err := detector.Detect()

	if err != nil {
		writer.WriteString(fmt.Sprintf("VERBOSE \"Failed to detect FAS: %s\"\n", err))
		return
	}

	writer.WriteString(fmt.Sprintf("SET VARIABLE FAS_DETECTED %t\n", detected))

}
