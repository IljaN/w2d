package deepl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestValidateResponse(t *testing.T) {
	var errorText string
	testErrorMes := map[string]string{
		"message": "test message",
	}
	testBody, err := json.Marshal(testErrorMes)
	if err != nil {
		t.Fatalf("Error within json.Marshal")
	}
	testResponse := http.Response{
		Status:     "200 test",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(testBody)),
	}

	for c := 100; c < 512; c++ {
		if http.StatusText(c) == "" {
			// unused stasus code
			continue
		}
		testResponse.Status = fmt.Sprintf("%d test", c)
		testResponse.StatusCode = c
		err := ValidateResponse(&testResponse)
		if c >= 200 && c < 300 {
			errorText = "If status code is between 200 and 299, no error should be returned"
			if err != nil {
				t.Fatalf("%s\nActual: %s", errorText, err.Error())
			}
			continue
		}
		if err == nil {
			t.Fatalf("If Status code is not 200 <= c < 300, should occur error\nStatus code: %d, Response: %v", c, testResponse)
		} else {
			if !strings.Contains(err.Error(), http.StatusText(c)) {
				errorText = fmt.Sprintf("Error text should include Status Code(%s)\nActual: %s",
					http.StatusText(c), err.Error())
			}
			if statusText, ok := KnownErrors[c]; ok { // errored
				if !strings.Contains(err.Error(), statusText) {
					errorText = fmt.Sprintf("If stasus code is knownded, the text should include it's error text\nExpected: %s\nActual: %s",
						statusText, err.Error())
					t.Fatalf("%s", errorText)
				}
			}
		}
	}
	// test when body is valid/invalid as json
	invalidResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("test"))),
	}
	err = ValidateResponse(&invalidResp)
	if err == nil {
		t.Fatalf("If status code is invalid, should occur error\nActual: %d", invalidResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), "]") {
		t.Fatalf("If body is invalid as json, error is formated as %s",
			"`Invalid response [statuscode statustext]`")
	}

	expectedMessage := "This is test"
	validResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(`{"message": "%s"}`, expectedMessage)))),
	}
	err = ValidateResponse(&validResp)
	if err == nil {
		t.Fatalf("If is status code invalid, should occur error\nActual: %d", validResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), expectedMessage) {
		t.Fatalf("If body is valid as json, error suffix should have `%s`\nActual: %s",
			expectedMessage, err.Error())
	}
}

func TestParseResponse(t *testing.T) {
	baseResponse := http.Response{
		Status:     "200 test",
		StatusCode: 200,
	}
	{
		input := map[string][]map[string]string{
			"Translations": make([]map[string]string, 3),
		}
		sampleRes := map[string]string{
			"detected_source_language": "test",
			"text":                     "test text",
		}
		input["Translations"][0] = sampleRes
		input["Translations"][1] = sampleRes
		input["Translations"][2] = sampleRes
		b, err := json.Marshal(input)
		if err != nil {
			t.Fatal("Error within json.Marshal")
		}
		baseResponse.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		res, err := parseResponse[TranslateResponse](&baseResponse)
		if err != nil {
			t.Fatalf("If input is valid, should not occur error\n%s", err.Error())
		}

		if len(res.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(res.Translations), len(input["Translations"]))
		}
	}
	{
		input := map[string][]map[string]string{
			"Translations": make([]map[string]string, 1),
		}
		sampleRes := map[string]string{
			"detected_source_language": "test",
			"text":                     "test text",
			"this will be ignored":     "test",
		}
		input["Translations"][0] = sampleRes
		b, err := json.Marshal(input)
		if err != nil {
			t.Fatal("Error within json.Marshal")
		}
		baseResponse.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		res, err := parseResponse[TranslateResponse](&baseResponse)
		if err != nil {
			t.Fatalf("If input is valid, should not occur error\n%s", err.Error())
		}
		if len(res.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(res.Translations), len(input["Translations"]))
		}
		resType := reflect.ValueOf(res.Translations[0]).Type()
		expectedNumOfField := 2
		if resType.NumField() != expectedNumOfField {
			t.Fatalf("Length of Translated's filed should be equal %d\nActual: %d", expectedNumOfField, resType.NumField())
		}
	}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTranslatePassesAllParams(t *testing.T) {
	c := client{
		Endpoint: ProEndpoint,
		AuthKey:  "abc",
		client: NewTestClient(func(req *http.Request) *http.Response {
			if req.URL.String() != ProEndpoint+"translate" {
				t.Fatalf("Url %s expected, got %s", ProEndpoint+"translate", req.URL.String())
			}

			_ = req.ParseForm()
			expParams := map[string]string{
				"auth_key":    "abc",
				"target_lang": "ru",
				"text":        "Hallo Welt!",
			}

			for k, v := range expParams {
				if !req.Form.Has(k) {
					t.Fatalf("Parameter %s missing but was expected", k)
				}

				if req.Form.Get(k) != v {
					t.Fatalf("Parameter %s was expected to be %s, got %s", k, expParams[k], v)
				}
			}

			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}

		}),
	}

	_, _ = c.Translate("Hallo Welt!", "ru", "")
	_, _ = c.TranslateToString("Hallo Welt!", "ru", "")

}

func TestEndpointForAuthKey(t *testing.T) {
	// Demo keys end with :fx
	demoKey := "abcdefg-ewfwfew-weffew-few:fx"
	if exp, got := FreeEndpoint, DetermineEndpoint(demoKey); exp != got {
		t.Fatalf("Expected endpoint to be free (%s), got %s", exp, got)
	}

	proKey := "gwegewgwg-ewfewf-fewfewf-32242"
	if exp, got := ProEndpoint, DetermineEndpoint(proKey); exp != got {
		t.Fatalf("Expected endpoint to be free (%s), got %s", exp, got)
	}
}
