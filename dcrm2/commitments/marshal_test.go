package commitments

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

type myStruct struct {
	Data string `json:"data"`
}

func (s *myStruct) MarshalJSON() ([]byte, error) {
	return []byte(`{"data":"charlie"}`), nil
}

func (s *myStruct) UnmarshalJSON(b []byte) error {
	// Insert the string directly into the Data member
	return json.Unmarshal(b, &s.Data)
}

func TestMarshal(t *testing.T) {
	// Create a struct with initial content "alpha"
	ms := &myStruct{"alpha"}

	// Replace content with "bravo" using custom UnmarshalJSON() (SUCCESSFUL)
	if err := json.NewDecoder(bytes.NewBufferString(`"bravo"`)).Decode(&ms); err != nil {
		log.Fatal(err)
	}

	// Trying another method (UNSUCCESSFUL)
	if ret, err := json.Marshal(ms); err != nil {
		log.Fatal(err)
	} else {
		t.Log(string(ret))
	}

	// Verify that the Marshaler interface is correctly implemented
	var marsh json.Marshaler
	marsh = ms
	ret, _ := marsh.MarshalJSON()
	t.Log(string(ret)) // Prints "charlie"
}
