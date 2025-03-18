package result

import (
	"encoding/xml"
	"fmt"
	"os"
)

func GetResult(path string) (*Testsuites, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results Testsuites
	if err := xml.NewDecoder(file).Decode(&results); err != nil {
		fmt.Println("Error decoding XML:", err)
		return nil, err
	}

	return &results, nil
}
