package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/piquette/finance-go/quote"
)

func main() {
	// Get the current price of Apple Inc. (AAPL)
	q, err := quote.Get("OPK")
	if err != nil {
		// Uh-oh.
		panic(err)
	}

	// Success!
	for {
		startCall(fmt.Sprintf("%.2f", q.Bid))
		time.Sleep(5 * time.Second)
	}
}

type SpeakRequest struct {
	Payload  string `json:"payload"`
	Voice    string `json:"voice"`
	Language string `json:"language"`
}

type CallRequest struct {
	ConnectionID string `json:"connection_id"`
	To           string `json:"to"`
	From         string `json:"from"`
}

func startCall(price string) {

	apiKey := os.Getenv("TELNYX_API_KEY")

	url := "https://api.telnyx.com/v2/calls"

	callReq := CallRequest{
		ConnectionID: "1720750306398045239",
		To:           "+13105316926", // The phone number to call
		From:         "+14422018839", // Your Telnyx phone number

	}

	// Marshal the request body to a []byte slice
	callReqBytes, err := json.Marshal(callReq)
	if err != nil {
		fmt.Printf("Error marshaling request body: %v\n", err)
		return
	}

	// Define the HTTP request for making a call
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(callReqBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Set the API key as the authentication header for the request
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the call
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(context.Background()))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Print the response body
	//fmt.Println("Response body:", string(body))

	var response Response

	// Unmarshal the response body into the Response struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshaling response body: %v\n", err)
		return
	} else {
		tries := 0
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			if getCallStatus(response.Data.CallControlID, apiKey) {
				//log.Println("EDD call still live")
				time.Sleep(5 * time.Second)
				break
			} else {
				log.Println("still dialing, or busy")
			}
			tries += i
			if tries == 25 {
				log.Println("tried to get Call status too many times, recalling")
				time.Sleep(5 * time.Second)
				return
			}
		}

		time.Sleep(5 * time.Second)

		startTalking(response.Data.CallControlID, apiKey, price)
	}
}

type StartCallResponse struct {
	Data struct {
		CallControlID string `json:"call_control_id"`
		CallLegID     string `json:"call_leg_id"`
		CallSessionID string `json:"call_session_id"`
		IsAlive       bool   `json:"is_alive"`
		RecordType    string `json:"record_type"`
	} `json:"data"`
}

func getCallStatus(id, key string) bool {
	BaseUrl := "https://api.telnyx.com/v2/calls/" + id

	req, err := http.NewRequest("GET", BaseUrl, nil)
	if err != nil {
		return false
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("here is the error in addCaller", err)
		return false
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var jsonResp StartCallResponse
	err = json.Unmarshal(bodyBytes, &jsonResp)
	if err != nil {
		log.Println("err unmarshalling in getStatus: ", err)
		log.Println(string(bodyBytes))
	} else {
		//log.Println("getCallStatus status: ", jsonResp.Data)
		//un comment to see call statuses
	}
	return jsonResp.Data.IsAlive
}

func startTalking(id, key, price string) {

	// Define the API endpoint for sending audio from text during a call
	url := "https://api.telnyx.com/v2/calls/" + id + "/actions/speak"

	// Define the request body for sending audio from text during a call
	speakReq := SpeakRequest{
		Payload:  "Hello, this is a message from a shareholder. You're current stock price is now " + price,
		Voice:    "female",
		Language: "en-US",
	}

	// Marshal the request body to a []byte slice
	speakReqBytes, err := json.Marshal(speakReq)
	if err != nil {
		fmt.Printf("Error marshaling request body: %v\n", err)
		return
	}

	// Define the HTTP request for sending audio from text during a call
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(speakReqBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Set the API key as the authentication header for the request
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(context.Background()))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Print the response body
	fmt.Println("Response body:", string(body))

}

func dialDigits(eddId, digitsRAW, key string) bool {

	BaseUrl := "https://api.telnyx.com/v2/calls/" + eddId + "/actions/send_dtmf"

	type Data struct {
		Digits string `json:"digits"`
	}
	digits := &Data{Digits: digitsRAW}
	jsonStr, err := json.Marshal(digits)

	req, err := http.NewRequest("POST", BaseUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("error posting digits", err)
		return false
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("here is the error in addCaller", err)
		return false
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	var jsonResp BridgeResponse
	err = json.Unmarshal(bodyBytes, &jsonResp)
	if err != nil {
		log.Println("err unmarshalling in dialingDigits: ", err)
		log.Println(string(bodyBytes))
		return false
	}
	//fmt.Println(jsonResp.Data)
	if jsonResp.Data.Result == "ok" {
		//log.Println("****Digits dialed****")
		return true
	}

	return false

}

type BridgeResponse struct {
	Data struct {
		Result string `json:"result"`
	} `json:"data"`
}

type Response struct {
	Data struct {
		CallControlID string      `json:"call_control_id"`
		CallLegID     string      `json:"call_leg_id"`
		CallSessionID string      `json:"call_session_id"`
		ClientState   interface{} `json:"client_state"`
		IsAlive       bool        `json:"is_alive"`
		RecordType    string      `json:"record_type"`
	} `json:"data"`
}

type TelnyxPost struct {
	ConnectionID string `json:"connection_id"`
	To           string `json:"to"`
	From         string `json:"from"`
}
