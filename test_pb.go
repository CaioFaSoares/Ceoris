package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	resp, err := http.Get("http://localhost:12090/api/collections/guilds/records")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}
