package main

import (
	"log"

	piston "github.com/milindmadhukar/go-piston"
)

func Exec(lang string, code string) (string, error) {
	client := piston.CreateDefaultClient()
	output, err := client.Execute(lang, "", // Passing language. Since no version is specified, it uses the latest supported version.
		[]piston.Code{
			{
				Content: code,
			},
		},
		// piston.Stdin("hello world"), // Passing input as "hello world".
	)
	if err != nil {
		log.Fatal(err)
		return "", nil
	}
	return output.GetOutput(), nil
}
