package centrifuge

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/galaco/bsp"
	"github.com/pkg/errors"
)

func Pakfile(f *bsp.Bsp) (*zip.Reader, error) {
	b := f.RawLump(bsp.LumpPakfile).RawContents()

	r := bytes.NewReader(b)

	zipR, err := zip.NewReader(r, int64(len(b)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open Pakfile lump")
	}

	return zipR, nil
}

func extractFile(f *zip.File, targetDir string) error {
	r, err := f.Open()
	if err != nil {
		return errors.Wrapf(err, "failed to open DDS file %q from archive", f.Name)
	}
	defer r.Close()

	ddsFileName := filepath.Join(targetDir, filepath.Base(f.Name))
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

func ExtractDdsFiles(f *bsp.Bsp, targetDir string) error {
	err := os.MkdirAll(targetDir, 0744)
	if err != nil {
		return errors.Wrapf(err, "failed to create target dir %q", targetDir)
	}

	zipR, err := Pakfile(f)

	for _, pakF := range zipR.File {
		if filepath.Ext(pakF.Name) == ".dds" {
			err := extractFile(pakF, targetDir)
			if err != nil {
				return errors.Wrapf(err, "failed to extract radar image %q", pakF.Name)
			}
		}
	}

	return nil
}

var radarOverviewInfoFilePattern = regexp.MustCompile("resource/overviews/([^\\.]+)\\.txt")

var ErrRadarImageNotFound = errors.New("failed to find radar overview image in BSP file")

func GetMapName(pakfile *zip.Reader) (string, error) {
	for _, pakF := range pakfile.File {
		matches := radarOverviewInfoFilePattern.FindStringSubmatch(pakF.Name)

		if len(matches) > 0 {
			return matches[1], nil
		}
	}

	return "", ErrRadarImageNotFound
}

func GetRadarImage(pakfile *zip.Reader) (radarInfoReader io.ReadCloser, radarOverviewImageReader io.ReadCloser, err error) {
	for _, pakF := range pakfile.File {
		matches := radarOverviewInfoFilePattern.FindStringSubmatch(pakF.Name)

		if len(matches) > 0 {
			radarInfoReader, err := pakF.Open()
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to open radar-info .txt file: %q", pakF.Name)
			}

			ddsPath := fmt.Sprintf("resource/overviews/%s_radar.dds", matches[1])

			fOverview, err := pakfile.Open(ddsPath)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to open radar-overview .dds file: %q", ddsPath)
			}

			return radarInfoReader, fOverview, nil
		}
	}

	return nil, nil, ErrRadarImageNotFound
}
