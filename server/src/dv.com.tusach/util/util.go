package util

import (
	"errors"
	"log"
	"os"
)

func ExtractError(err interface{}) error {
	// find out what exactly is err
	switch x := err.(type) {
	case string:
		return errors.New(x)
	case error:
		return x
	default:
		return errors.New("Unknow error")
	}
}

func SaveFile(filename string, data []byte) error {
	log.Println("saving file: ", filename)
	fo, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = fo.Write(data)
	if err != nil {
		return err
	}
	return nil
}
