package main

import (
	piston "github.com/milindmadhukar/go-piston"
)

func Exec(lang string, version string, code string) (string, error) {
	client := piston.CreateDefaultClient()
	output, err := client.Execute(lang, version,
		[]piston.Code{
			{
				Content: code,
			},
		},
		// piston.Stdin("hello world"), // Passing input as "hello world".
	)
	if err != nil {
		return "", err
	}
	return output.GetOutput(), nil
}
