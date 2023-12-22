package utilities

import (
	"fmt"
	"io"
)

func MustRead(in io.Reader, data []byte) error {
	c, err := io.ReadFull(in, data)

	if err != nil {
		return err
	}

	if c != len(data) {
		return fmt.Errorf("short read %d/%d", c, len(data))
	}
	return nil
}

func MustWrite(out io.Writer, data []byte) error {
	c, err := out.Write(data)

	if err != nil {
		return err
	}

	if c != len(data) {
		return fmt.Errorf("short write")
	}
	return nil
}
