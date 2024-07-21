package main

import (
	"debug/pe"
	"fmt"
	"os"
)

// ReadSectionsPE reads the binary content of each section in the PE file
func ReadSectionsPE(file *os.File) (map[string][]byte, error) {
	sections := make(map[string][]byte)

	// Parse the PE file
	peFile, err := pe.NewFile(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing PE file: %v", err)
	}
	defer file.Close()

	// Iterate through each section in the file
	for _, sec := range peFile.Sections {
		// Read the content of the section
		content, err := readSectionBytes(sec, file)
		if err != nil {
			return nil, fmt.Errorf("error reading section %s: %v", sec.Name, err)
		}
		sections[sec.Name] = content
	}

	return sections, nil
}

// readSectionBytes reads the bytes of a specific section from a PE file
func readSectionBytes(sec *pe.Section, file *os.File) ([]byte, error) {
	// Seek to the start of the section
	if _, err := file.Seek(int64(sec.Offset), 0); err != nil {
		return nil, fmt.Errorf("error seeking to section %s: %v", sec.Name, err)
	}

	// Read the content of the section
	data := make([]byte, sec.Size)
	if _, err := file.Read(data); err != nil {
		return nil, fmt.Errorf("error reading section %s: %v", sec.Name, err)
	}

	return data, nil
}

// compareBytes compares bytes in the same position between two byte slices
// and returns the percentage of bytes that are equal.
func compareBytes(bytes1, bytes2 []byte) float64 {

	// Check if either slice is empty
	if len(bytes1) == 0 || len(bytes2) == 0 {
		return 0.0
	}
	
	// Determine the length of the shorter byte slice
	var totalBytes int
	if len(bytes1) < len(bytes2) {
		totalBytes = len(bytes1)
	} else {
		totalBytes = len(bytes2)
	}

	// Count the number of matching bytes
	matchingBytes := 0
	for i := 0; i < totalBytes; i++ {
		if bytes1[i] == bytes2[i] {
			matchingBytes++
		}
	}

	// Calculate the percentage of similarity
	similarity := float64(matchingBytes) / float64(totalBytes) * 100.0
	return similarity
}


func main() {
	// Check if two file paths are provided as command-line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./program file1 file2")
		return
	}

	// Open the PE files
	file1, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Error opening file1:", err)
		return
	}
	defer file1.Close()

	file2, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println("Error opening file2:", err)
		return
	}
	defer file2.Close()

	// Read the sections of file1
	sections1, err := ReadSectionsPE(file1)
	if err != nil {
		fmt.Println("Error reading sections from file1:", err)
		return
	}

	// Read the sections of file2
	sections2, err := ReadSectionsPE(file2)
	if err != nil {
		fmt.Println("Error reading sections from file2:", err)
		return
	}

	// Compare the sections of both files
	for name, content1 := range sections1 {
		content2, ok := sections2[name]
		if !ok {
			fmt.Printf("Error: Section %s not present in file2\n", name)
			continue
		}

		similarity := compareBytes(content1, content2)
		fmt.Printf("Similarity of section %s: %.2f%%\n", name, similarity)
	}
}
