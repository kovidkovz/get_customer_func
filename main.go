package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Scope        any    `json:"scope"`
}

func fetchCustomerData() ([]byte, error) {
	// STEP 1: Login API Call
	
	// prepare payload for the api
	loginPayload := map[string]string{
		"username": os.Getenv("USERNAME"),
		"password": os.Getenv("PASSWORD"),
	}

	// convert into bytes
	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login payload: %v", err)
	}

	// prepare login request
	loginReq, err := http.NewRequest("POST", "https://online.v3.staging.traxmate.io/api/auth/login", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %v", err)
	}

	// set headers
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("Accept", "application/json, text/plain, */*")

	// create an http client
	client := &http.Client{}

	// make request to the api
	loginResp, err := client.Do(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make login request: %v", err)
	}
	defer loginResp.Body.Close()

	// extract response body
	body, _ := io.ReadAll(loginResp.Body)
	if loginResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: %s", body)
	}

	// unmarshal the byte data into the struct
	var loginData LoginResponse
	if err := json.Unmarshal(body, &loginData); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %v", err)
	}

	// STEP 2: Use token to call customers API
	customerReq, err := http.NewRequest("GET", "https://online.v3.staging.traxmate.io/api/customers/all", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer request: %v", err)
	}

	// Extract the token from the loginData struct
	authHeader := "Bearer " + loginData.Token

	// set the header for the get all customer api
	customerReq.Header.Set("x-authorization", authHeader)

	// call the api
	customerResp, err := client.Do(customerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer data: %v", err)
	}
	defer customerResp.Body.Close()

	// extract the response body
	customerBody, _ := io.ReadAll(customerResp.Body)
	if customerResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("customer API failed: %s", customerBody)
	}

	return customerBody, nil
}

func callCustomerAPI() {
	fmt.Println("Calling fetchCustomerData at:", time.Now().Format(time.RFC1123))

	custData, err := fetchCustomerData()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("customer_data:", string(custData))
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Create a ticker that ticks every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop() // clean up the ticker at the end of the main func

	// call the wrapper function as soon as the server starts
	callCustomerAPI()

	// Loop: call function every 5 minutes
	go func() {
		for range ticker.C {
			callCustomerAPI()
		}
	}()

	select{} //keep the main func running..... just block the function here, so that the goroutine keeps running in the background
}
