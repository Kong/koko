package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)

// This generator generates the directory structure for the gen/jsonschema
// package. This is done before generating the schemas because if gen/jsonschema
// package doesn't exist, the jsonschema generator won't compile.
//go:generate go run main.go
func main() {
	genDir := "../../gen"
	err := ensureDir(genDir)
	if err != nil {
		log.Fatalln(err)
	}
	jsonSchemaDir := genDir + "/jsonschema"
	err = ensureDir(jsonSchemaDir)
	if err != nil {
		log.Fatalln(err)
	}
	schemasDir := jsonSchemaDir + "/schemas"
	err = ensureDir(schemasDir)
	if err != nil {
		log.Fatalln(err)
	}

	embedFilename := "embed.go"
	filepath := fmt.Sprintf("%s/%s", jsonSchemaDir, embedFilename)
	err = ioutil.WriteFile(filepath, []byte(embedContent), fs.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}

func ensureDir(dir string) error {
	err := os.Mkdir(dir, fs.ModePerm)
	if err != nil {
		pathErr, ok := err.(*os.PathError)
		if !ok {
			panic(err)
		}
		if errors.Is(pathErr, os.ErrExist) {
			return nil
		}
	}
	return err
}

var embedContent = `package jsonschema

import "embed"

var KongSchemas embed.FS
`
