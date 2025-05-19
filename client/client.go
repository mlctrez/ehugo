package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mlctrez/ehugo/hueapi"
	"io"
	"net/http"
	"os"
)

func main() {
	apiHost := flag.String("host", "http://localhost", "Host address of the Hue API server")
	lightName := flag.String("name", "", "Name of the light to create")

	flag.Parse()

	if *lightName == "" {
		fmt.Println("Error: Light name is required")
		flag.Usage()
		os.Exit(1)
	}

	light := hueapi.LightInfo{Name: *lightName}

	jsonData, err := json.Marshal(light)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/username/lights", *apiHost)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned status %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	fmt.Printf("Successfully created light:\n")
	fmt.Printf("REPLY: %s\n", string(body))
}
