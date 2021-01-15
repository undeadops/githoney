package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitlabPost(t *testing.T) {
	data := []byte(`{
		"employees":{
		   "protected":false,
		   "address":{
			  "street":"22 Saint-Lazare",
			  "postalCode":"75003",
			  "city":"Paris",
			  "countryCode":"FRA",
			  "country":"France"
		   },
		   "employee":[
			  {
				 "id":1,
				 "first_name":"Jeanette",
				 "last_name":"Penddreth"
			  },
			  {
				 "id":2,
				 "firstName":"Giavani",
				 "lastName":"Frediani"
			  }
		   ]
		}
	 }`)

	c := Config{
		Port:            "",
		HoneycombApiKey: "",
		GitlabAuthToken: "1234",
	}

	req, err := http.NewRequest("POST", "/gitlab", bytes.NewBuffer(data))
	checkError(err, t)

	rr := httptest.NewRecorder()

	http.HandlerFunc(c.gitlabWebhook).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected an HTTP %d. Got %d", http.StatusOK, status)
	}

	expected := "\"OK\""
	assert.Equal(t, expected, rr.Body.String(), "Response body differs")

}

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Errorf("An error occured. %v", err)
	}
}
