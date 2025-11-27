package mpesa

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// CallbackResponse represents the structure of the M-Pesa callback JSON
type CallbackResponse struct {
	Body struct {
		stkCallback `json:"stkCallback"`
	} `json:"Body"`
}

type stkCallback struct {
	MerchantRequestID string           `json:"MerchantRequestID"`
	CheckoutRequestID string           `json:"CheckoutRequestID"`
	ResultCode        int              `json:"ResultCode"`
	ResultDesc        string           `json:"ResultDesc"`
	CallbackMetadata  callbackMetadata `json:"CallbackMetadata,omitempty"`
}

type callbackMetadata struct {
	Item []callbackItem `json:"Item"`
}

type callbackItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value,omitempty"`
}

// ProcessCallback handles the incoming M-Pesa callback request
func (c *Client) ProcessCallback(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading callback body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log raw body for debugging
	log.Printf("Received M-Pesa Callback: %s", string(body))

	// Parse JSON
	var callback CallbackResponse
	if err := json.Unmarshal(body, &callback); err != nil {
		log.Printf("Error parsing callback JSON: %v", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Process the result
	resultCode := callback.Body.stkCallback.ResultCode
	resultDesc := callback.Body.stkCallback.ResultDesc
	merchantID := callback.Body.stkCallback.MerchantRequestID
	checkoutID := callback.Body.stkCallback.CheckoutRequestID

	if resultCode == 0 {
		log.Printf("✅ Transaction Successful! [MerchantID: %s, CheckoutID: %s]", merchantID, checkoutID)
		// Extract metadata items if needed
		for _, item := range callback.Body.stkCallback.CallbackMetadata.Item {
			log.Printf("   - %s: %v", item.Name, item.Value)
		}
	} else {
		log.Printf("❌ Transaction Failed! [MerchantID: %s, CheckoutID: %s] Reason: %s", merchantID, checkoutID, resultDesc)
	}

	// Respond to M-Pesa
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ResultCode": 0, "ResultDesc": "Accepted"}`))
}
