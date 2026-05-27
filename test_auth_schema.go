package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	authReq := map[string]string{
		"identity": "admin@chantry.local",
		"password": "chantry_admin_123!",
	}
	authBytes, _ := json.Marshal(authReq)
	
	resp, err := http.Post("http://localhost:12090/api/admins/auth-with-password", "application/json", bytes.NewBuffer(authBytes))
	if err != nil {
		fmt.Println("Auth Error:", err)
		return
	}
	defer resp.Body.Close()
	
	var authResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&authResp)
	token := authResp["token"].(string)
	
	req, _ := http.NewRequest("GET", "http://localhost:12090/api/collections/guilds", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp2, _ := http.DefaultClient.Do(req)
	defer resp2.Body.Close()
	
	var coll map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&coll)
	
	b, _ := json.MarshalIndent(coll["schema"], "", "  ")
	fmt.Println(string(b))
}
