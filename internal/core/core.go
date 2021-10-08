package core

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/galaco/bsp"
	"github.com/pkg/errors"
)

func extractRadarImage(f *zip.File) error {
	r, err := f.Open()
	if err != nil {
		return errors.Wrapf(err, "failed to open DDS file %q from archive", f.Name)
	}
	defer r.Close()

	ddsFileName := filepath.Base(f.Name)
	ddsF, err := os.Create(ddsFileName)
	if err != nil {
		return errors.Wrapf(err, "failed to create DDS file %q on disk", ddsFileName)
	}
	defer ddsF.Close()

	_, err = io.Copy(ddsF, r)
	if err != nil {
		return errors.Wrapf(err, "failed to write DDS data to file %q", ddsFileName)
	}

	return nil
}

func ExtractRadarImages(f *bsp.Bsp) error {
	b := f.RawLump(bsp.LumpPakfile).RawContents()
	r := bytes.NewReader(b)
	zipR, err := zip.NewReader(r, int64(len(b)))
	if err != nil {
		return errors.Wrapf(err, "failed to open Pakfile lump")
	}

	for _, pakF := range zipR.File {
		if filepath.Ext(pakF.Name) == ".dds" {
			err := extractRadarImage(pakF)
			if err != nil {
				return errors.Wrapf(err, "failed to extract radar image %q", pakF.Name)
			}
		}
	}

	return nil
}
