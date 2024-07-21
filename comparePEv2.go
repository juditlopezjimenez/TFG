package main

import (
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
//	"strings"
//	"github.com/xuri/excelize/v2"
//	"math"
	"strconv"
	"encoding/json"
	"encoding/csv"
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

// comparePEFiles compares two PE files and returns the similarity of their sections
func comparePEFiles(file1, file2 string) (map[string]float64, error) {
	// Open the PE files
	f1, err := os.Open(file1)
	if err != nil {
		return nil, fmt.Errorf("error opening file1: %v", err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return nil, fmt.Errorf("error opening file2: %v", err)
	}
	defer f2.Close()

	// Read the sections of file1
	sections1, err := ReadSectionsPE(f1)
	if err != nil {
		return nil, fmt.Errorf("error reading sections from file1: %v", err)
		
	}

	// Read the sections of file2
	sections2, err := ReadSectionsPE(f2)
	if err != nil {
		return nil, fmt.Errorf("error reading sections from file2: %v", err)
		
	}

	// Compare the sections of both files
	results := make(map[string]float64)
	for name, content1 := range sections1 {
		content2, ok := sections2[name]
		if !ok {
			continue
		}

		similarity := compareBytes(content1, content2)
		results[name] = similarity
	}

	return results, nil
}

// compareAllPEFiles compares all PE files in a folder and returns the similarity matrix
func compareAllPEFiles(folder string) (map[string]map[string]map[string]float64, error) {
    similarityMatrix := make(map[string]map[string]map[string]float64)

    // Walk through the folder to find PE files
    err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            // Initialize the inner map for the current file
            similarityMatrix[info.Name()] = make(map[string]map[string]float64)

            // Compare the current file with all other files
            err := filepath.Walk(folder, func(path2 string, info2 os.FileInfo, err error) error {
                if err != nil {
                    return err
                }
                if !info2.IsDir() && info.Name() != info2.Name() {
                    // Compare the files
                    similarity, err := comparePEFiles(filepath.Join(folder, info.Name()), filepath.Join(folder, info2.Name()))
                    if err != nil {
                        return err
                    }
                    // Store the similarity value in the matrix
                    similarityMatrix[info.Name()][info2.Name()] = similarity
                }
                return nil
            })
            if err != nil {
                return err
            }
        }
        return nil
    })

    if err != nil {
        return nil, err
    }

    return similarityMatrix, nil
}

func writeSimilarityMatrixToCSV(filename string, innerMap map[string]map[string]float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Generate headers dynamically
	var headers []string
	for innerKey := range innerMap {
		headers = append(headers, innerKey)
	}
	headers = append([]string{""}, headers...)

	// Write CSV header
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for outerKey, outerMap := range innerMap {
		row := make([]string, len(headers))
		row[0] = outerKey
		for i, header := range headers[1:] {
			row[i+1] = strconv.FormatFloat(outerMap[header], 'f', -1, 64)
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}


func transformDataset(original map[string]map[string]map[string]float64) map[string]map[string]map[string]float64 {
	transformed := make(map[string]map[string]map[string]float64)

	for outerKey, innerMap := range original {
		for innerKey, innerInnerMap := range innerMap {
			for innerInnerKey, value := range innerInnerMap {
				if _, exists := transformed[innerInnerKey]; !exists {
					transformed[innerInnerKey] = make(map[string]map[string]float64)
				}
				if _, exists := transformed[innerInnerKey][innerKey]; !exists {
					transformed[innerInnerKey][innerKey] = make(map[string]float64)
				}
				transformed[innerInnerKey][innerKey][outerKey] = value
			}
		}
	}

	return transformed
}


func printJSON(data map[string]map[string]map[string]float64) {
	// Convert map to JSON
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Print JSON
	fmt.Println(string(jsonData))
}


func main() {
	// Check if a folder path is provided as a command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./program folder_path")
		return
	}

	folderPath := os.Args[1]

	// Compare all PE files in the folder
	similarityMatrix, err := compareAllPEFiles(folderPath)
	if err != nil {
		fmt.Println("Error comparing PE files:", err)
		return
	}

	newsimilarityMatrix := transformDataset(similarityMatrix)
	printJSON(newsimilarityMatrix)

	// Write the similarity matrix to an Excel file
	for key, value := range newsimilarityMatrix {
		filename := key + ".csv"
		if err := writeSimilarityMatrixToCSV(filename, value); err != nil {
			panic(err)
		}
	}
}

