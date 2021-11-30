package main

import (
	"net/http"

	piston "github.com/milindmadhukar/go-piston"
)

var httpClient = http.DefaultClient
var client = piston.New("", httpClient, PISTON_URL)

func Exec(lang string, version string, code string) (string, error) {
	output, err := client.Execute(lang, version,
		[]piston.Code{
			{
				Content: code,
			},
		},
	)
	if err != nil {
		return "", err
	}
	return output.GetOutput(), nil
}

func GetLanguages() *[]string {
	languages := client.GetLanguages()
	return languages
}
