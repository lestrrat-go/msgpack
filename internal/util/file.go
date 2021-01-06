package util

import (
	"fmt"
	"go/format"
	"os"

	"github.com/pkg/errors"
)

func WriteFormattedFile(fn string, src []byte) error {
	formatted, err := format.Source(src)
	if err != nil {
		fmt.Println(string(src))
		return err
	}

	dst, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, `failed to open file`)
	}
	defer dst.Close()

	dst.Write(formatted)
	return nil
}
