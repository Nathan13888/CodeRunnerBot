package main

import (
	piston "github.com/milindmadhukar/go-piston"
	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Msg("error while executing code")
		return "", err
	}
	return output.GetOutput(), nil
}
