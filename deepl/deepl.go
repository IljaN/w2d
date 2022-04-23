package deepl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Client for the DeepL.com translation API. See: https://www.deepl.com/en/docs-api/
//
// To access the API an AuthKey is required which can be obtained by creating "DeepL.com API account" (Free or Pro) at
// https://www.deepl.com/pro . The Client only covers a subset of all supported API functionality.
type Client interface {
	// Translate the given text from sourceLang to targetLang. Set sourceLang to "" (empty-string) to use automatic source-
	// language detection. Use the SupportedLanguages method to query possible values for targetLang and sourceLang.
	Translate(text, targetLang, sourceLang string) ([]string, error)
	// TranslateToString same as Translate but returns the concatenated text
	TranslateToString(text, targetLang, sourceLang string) (string, error)
	// SupportedLanguages returns the list of supported source languages if target is set to false. Otherwise,
	// the supported target languages are returned.
	SupportedLanguages(target bool) (map[string]SupportedLanguage, error)
}

const (
	ProEndpoint  = "https://api.deepl.com/v2/"
	FreeEndpoint = "https://api-free.deepl.com/v2/"
)

type client struct {
	Endpoint string
	AuthKey  string
	client   *http.Client
}

func NewClient(authKey string) Client {
	return &client{
		Endpoint: DetermineEndpoint(authKey),
		AuthKey:  authKey,
		client:   http.DefaultClient,
	}
}

type TranslateResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	}
}

// TranslateToString is a helper which calls Translate and concatenates the result in to a single string
func (c *client) TranslateToString(text, targetLang, sourceLang string) (string, error) {
	s, err := c.Translate(text, targetLang, sourceLang)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	sb.Grow(2028)

	for k := range s {
		sb.WriteString(s[k])
	}

	return sb.String(), nil
}

// Translate the given text from sourceLang to targetLang. Set sourceLang to "" (empty-string) to use automatic source-
// language detection. Use the SupportedLanguages method to query possible values for targetLang and sourceLang.
func (c *client) Translate(text, targetLang, sourceLang string) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)
	params.Add("target_lang", targetLang)
	params.Add("text", text)
	if sourceLang != "" {
		params.Add("source_lang", sourceLang)
	}

	ep := c.Endpoint + "translate"
	resp, err := c.client.PostForm(ep, params)

	if err := validateResponse(resp); err != nil {
		return []string{}, err
	}
	parsed, err := parseResponse[TranslateResponse](resp)
	if err != nil {
		return []string{}, err
	}
	r := []string{}
	for _, translated := range parsed.Translations {
		r = append(r, translated.Text)
	}
	return r, nil
}

type SupportedLanguageResponse []SupportedLanguage

type SupportedLanguage struct {
	Language          string `json:"language"`
	Name              string `json:"name"`
	SupportsFormality bool   `json:"supports_formality"`
}

// SupportedLanguages returns the list of supported source languages if target is set to false. Otherwise,
// the supported target languages are returned.
func (c *client) SupportedLanguages(target bool) (map[string]SupportedLanguage, error) {
	ep := c.Endpoint + "languages"
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)

	if target {
		params.Add("target", "target")
	}

	resp, err := c.client.PostForm(ep, params)

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	parsed, err := parseResponse[SupportedLanguageResponse](resp)
	if err != nil {
		return nil, err
	}

	supportedLangs := make(map[string]SupportedLanguage, len(parsed))

	for k := range parsed {
		supportedLangs[parsed[k].Language] = parsed[k]
	}

	return supportedLangs, nil
}

var KnownErrors = map[int]string{
	400: "Bad request. Please check error message and your parameters.",
	403: "Authorization failed. Please supply a valid auth_key parameter.",
	404: "The requested resource could not be found.",
	413: "The request size exceeds the limit.",
	414: "The request URL is too long. You can avoid this error by using a POST request instead of a GET request, and sending the parameters in the HTTP body.",
	429: "Too many requests. Please wait and resend your request.",
	456: "Quota exceeded. The character limit has been reached.",
	503: "Resource currently unavailable. Try again later.",
	529: "Too many requests. Please wait and resend your request.",
} // this from https://www.deepl.com/docs-api/accessing-the-api/error-handling/

func validateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var data map[string]interface{}
		baseErrorText := fmt.Sprintf("Invalid response [%d %s]",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
		if t, ok := KnownErrors[resp.StatusCode]; ok {
			baseErrorText += fmt.Sprintf(" %s", t)
		}
		e := json.NewDecoder(resp.Body).Decode(&data)
		if e != nil {
			return fmt.Errorf("%s", baseErrorText)
		} else {
			return fmt.Errorf("%s, %s", baseErrorText, data["message"])
		}
	}
	return nil
}

// parseResponse parses json response in to target type R
func parseResponse[R any](resp *http.Response) (R, error) {
	var parsed R
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		err := fmt.Errorf("%s (occurred while parse response)", err.Error())
		return parsed, err
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		err := fmt.Errorf("%s (occurred while parse response)", err.Error())
		return parsed, err
	}
	return parsed, err
}

// DetermineEndpoint returns the base api-endpoint depending on whether authKey belongs to a free or pro account.
func DetermineEndpoint(authKey string) string {
	if strings.HasSuffix(authKey, ":fx") {
		return FreeEndpoint
	}

	return ProEndpoint
}
