package main

import (
	"net/http"

	piston "github.com/milindmadhukar/go-piston"
)

func Exec(lang string, version string, code string) (string, error) {
	httpClient := http.DefaultClient
	client := piston.New("", httpClient, PISTON_URL)
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

func GetRuntimes() *piston.Runtimes {
	httpClient := http.DefaultClient
	client := piston.New("", httpClient, PISTON_URL)
	runtimes, err := client.GetRuntimes()
	if err != nil {
		return nil
	}
	return runtimes
}
