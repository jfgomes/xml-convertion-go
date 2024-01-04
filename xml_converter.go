package main

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-test/convert_to_json"
	"go-test/convert_to_rosetta"
)

func xmlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error getting file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	compressedContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading the file", http.StatusInternalServerError)
		return
	}

	// Creating a Gzip reader from compressed content
	reader, err := gzip.NewReader(bytes.NewReader(compressedContent))
	if err != nil {
		http.Error(w, "Error creating Gzip reader", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Reading uncompressed content
	decompressedContent, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, "Error reading uncompressed content", http.StatusInternalServerError)
		return
	}

	// Adding a log to check the uncompressed XMLo
	jsonData, err := convert_to_json.ConvertXMLToJSON(decompressedContent)
	if err != nil {
		http.Error(w, "Error converting to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Creating a map to store the result
	var result map[string]interface{}

	// Decoding JSON to the map
	err2 := json.Unmarshal([]byte(jsonData), &result)
	if err2 != nil {
		fmt.Println("Error decoding JSON:", err2)
		return
	}

	// Converting to Rosetta
	rosettaXML, err := convert_to_rosetta.ConvertJSONToRosetta(result, w) // Assuming 'result' is your decoded JSON map
	if err != nil {
		http.Error(w, "Error converting to Rosetta: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get owner as the name to write to file
	ownerEmail := convert_to_rosetta.ConvertOwnerEmail(result)
	fmt.Println("Owner Email:", ownerEmail)

	err3 := SaveNewXml(rosettaXML, "converted/"+ownerEmail+".xml")
	if err3 != nil {
		// error
	}

	w.Header().Set("Content-Type", "application/json")
}

func SaveNewXml(rosettaXML string, filePath string) error {
	// Write content to file
	err := ioutil.WriteFile(filePath, []byte(rosettaXML), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	fmt.Println("Content saved with success!")
	return nil
}

func main() {
	http.HandleFunc("/convert", xmlHandler)

	// Serve static docs
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	fmt.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}

}
