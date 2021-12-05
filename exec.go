package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	piston "github.com/milindmadhukar/go-piston"
)

// TODO: make this configurable
const USERAGENT = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"

// TODO: if this is well written, this should be ported to its own package/wrapper
// POST /api/v2/execute
type ExecuteRequest struct {
	Language           string   `json:"language"`             // required, language of code
	Version            string   `json:"version"`              // required, language of version
	Files              []File   `json:"files"`                // required, files of code
	Stdin              string   `json:"stdin"`                // input to code
	Args               []string `json:"args"`                 // program arguments
	CompileTimeout     int      `json:"compile_timeout"`      // max time for compiling; default: 10000 MS
	RunTimeout         int      `json:"run_timeout"`          // max run time; default: 3000 MS
	CompileMemoryLimit int      `json:"compile_memory_limit"` // max memory for compile: -1
	RunMemoryLimit     int      `json:"run_memory_limit"`     // max memory for run; default: -1
}

type ExecuteResponse struct {
	Language string         `json:"language"`
	Version  string         `json:"version"`
	Run      ExecuteResults `json:"run"`
	Message  string         `json:"message"` // means something bad happened...
}

type ExecuteResults struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Output string `json:"output"`
	Code   int    `json:"code"`
	Signal string `json:"signal"` // what is this??
}

type File struct {
	Name     string `json:"name"`     // name of upload
	Content  string `json:"content"`  // required, content of file
	Encoding string `json:"encoding"` // encoding used; default: utf8; options base64, hex
}

// TODO: runtime endpoints
func Exec(lang string, version string, code string) (string, error) {
	execRequest := ExecuteRequest{
		Language: lang,
		Version:  version,
		Files: []File{
			{
				Content: code,
			},
		},
	}
	if version == "" {
		latest, err := GetLatestVersion(lang)
		if err != nil {
			return "", err
		}
		execRequest.Version = latest
	}

	body, err := json.Marshal(execRequest)
	if err != nil {
		return "", err
	}

	res, err := Request("POST", PISTON_URL+"execute", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var results ExecuteResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&results)

	return results.Run.Output, err
}

func GetRuntimes() (*piston.Runtimes, error) {
	httpClient := http.DefaultClient
	client := piston.New("", httpClient, PISTON_URL)

	runtimes, err := client.GetRuntimes()
	if err != nil {
		return nil, err
	}

	return runtimes, nil
}

// TODO: there should be a static list of runtimes which the bot refers to; any issues if the runtimes change??
func GetLatestVersion(language string) (string, error) {
	runtimes, err := GetRuntimes()
	if err != nil {
		return "", err
	}

	if runtimes == nil {
		return "", errors.New("no runtimes found")
	}

	for _, runtime := range *runtimes {
		if language == runtime.Language || isPresent(runtime.Aliases, language) {
			return runtime.Version, nil
		}
	}

	return "", errors.New("Could not find a version for the language " + language)
}

// "Borrowed" from that piston wrapper library
// Returns a boolean value checking if a string is found in the slice or not.
func isPresent(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func Request(method string, path string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", USERAGENT)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)

	return res, err
}
