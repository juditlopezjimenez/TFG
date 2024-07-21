package main

import (
    "debug/elf"
    "fmt"
    "os"
    "math"
)

// ReadSections reads the binary content of each section in the ELF file
func ReadSections(file *elf.File) (map[string][]byte, error) {
    sections := make(map[string][]byte)

    // Iterate through each section in the file
    for _, sec := range file.Sections {
        // Read the content of the section
        data, err := readSectionBytes(sec)
        if err != nil {
            return nil, err
        }
        sections[sec.Name] = data
    }

    return sections, nil
}

// readSectionBytes reads the bytes of a specific section from a binary file
func readSectionBytes(sec *elf.Section) ([]byte, error) {
    // Check if the section type is SHT_NOBITS
    if sec.Type == elf.SHT_NOBITS {
        // For SHT_NOBITS sections, return an empty slice
        return nil, nil
    }

    // Read the content of the section
    data, err := sec.Data()
    if err != nil {
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

    // Read the contents of file1
    file1, err := elf.Open(os.Args[1])
    if err != nil {
        fmt.Println("Error reading file1:", err)
        return
    }
    defer file1.Close()

    // Read the contents of file2
    file2, err := elf.Open(os.Args[2])
    if err != nil {
        fmt.Println("Error reading file2:", err)
        return
    }
    defer file2.Close()

    sectionsf1, err := ReadSections(file1)
    if err != nil {
        fmt.Println("Error reading sections from file1:", err)
        return
    }

    sectionsf2, err := ReadSections(file2)
    if err != nil {
        fmt.Println("Error reading sections from file2:", err)
        return
    }

    // Compare bytes between file1 and file2
    var totalSimilarity float64
    var numSections int
    for name, content1 := range sectionsf1 {
        content2, ok := sectionsf2[name]
        if !ok {
            fmt.Printf("Error: Section %s not present in both files\n", name)
            continue
        }

        // Skip empty sections
        if content1 == nil || content2 == nil {
            continue
        }

        similarity := compareBytes(content1, content2)
        if !math.IsNaN(similarity) {
            totalSimilarity += similarity
            numSections++
            fmt.Printf("Similarity of section %s: %.2f%%\n", name, similarity)
        } else {
            fmt.Printf("Error: Unable to calculate similarity of section %s\n", name)
        }
    }

    if numSections == 0 {
        fmt.Println("Error: No common sections present in both files")
        return
    }
}
