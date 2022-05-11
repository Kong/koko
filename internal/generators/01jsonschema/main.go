package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/generator"
	_ "github.com/kong/koko/internal/resource"
)

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

	schema := generator.GlobalSchema()
	for name, schema := range schema.Definitions {
		jsonSchema, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		filepath := fmt.Sprintf("%s/%s.json", schemasDir, name)
		err = ioutil.WriteFile(filepath, jsonSchema, fs.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
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

//go:embed schemas/*
var KongSchemas embed.FS
`
