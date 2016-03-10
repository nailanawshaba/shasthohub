// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io"
	"os"

	"github.com/keybase/client/go/logger"
)

type SafeWriter interface {
	GetFilename() string
	WriteTo(io.Writer) (int64, error)
}

// File defines a default SafeWriter implementation
type File struct {
	filename string
	data     []byte
	perm     os.FileMode
}

// NewFile returns a File
func NewFile(filename string, data []byte, perm os.FileMode) File {
	return File{filename, data, perm}
}

// Save file
func (f File) Save(log logger.Logger) error {
	return SafeWriteToFile(f, f.perm, log)
}

// GetFilename is for SafeWriter
func (f File) GetFilename() string {
	return f.filename
}

// WriteTo is for SafeWriter
func (f File) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.data)
	return int64(n), err
}

// SafeWriteToFile to safely write to a file. Use mode=0 for default permissions.
func SafeWriteToFile(t SafeWriter, mode os.FileMode, log logger.Logger) error {
	fn := t.GetFilename()
	log.Debug(fmt.Sprintf("+ Writing to %s", fn))
	tmpfn, tmp, err := OpenTempFile(fn, "", mode)
	log.Debug(fmt.Sprintf("| Temporary file generated: %s", tmpfn))
	if err != nil {
		return err
	}

	_, err = t.WriteTo(tmp)
	if err == nil {
		err = tmp.Close()
		if err == nil {
			err = os.Rename(tmpfn, fn)
		} else {
			log.Error(fmt.Sprintf("Error closing temporary file %s: %s", tmpfn, err))
			os.Remove(tmpfn)
		}
	} else {
		log.Error(fmt.Sprintf("Error writing temporary file %s: %s", tmpfn, err))
		tmp.Close()
		os.Remove(tmpfn)
	}
	log.Debug(fmt.Sprintf("- Wrote to %s -> %s", fn, errToOk(err)))
	return err
}

func errToOk(err error) string {
	if err == nil {
		return "ok"
	}
	return "ERROR: " + err.Error()
}
