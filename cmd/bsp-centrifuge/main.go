package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/galaco/bsp"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/saiko-tech/bsp-centrifuge/pkg/centrifuge"
	"github.com/saiko-tech/bsp-centrifuge/pkg/steamapi"
)

func pathToBsp(path string) (*bsp.Bsp, error) {
	if path == "-" {
		bspF, err := bsp.ReadFromStream(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read BSP data from stdin")
		}

		return bspF, nil
	}

	bspF, err := bsp.ReadFromFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read BSP data from file: %q", path)
	}

	return bspF, nil
}

func extractPakfile(bspFile, targetFile string) error {
	bspF, err := pathToBsp(bspFile)
	if err != nil {
		return errors.Wrap(err, "failed to read BSP data")
	}

	b := bspF.RawLump(bsp.LumpPakfile).RawContents()
	r := bytes.NewReader(b)

	var w io.Writer
	if targetFile == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(targetFile)
		if err != nil {
			return errors.Wrapf(err, "failed to create target file: %q", targetFile)
		}
		defer f.Close()

		w = f
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to extract/copy pakfile data")
	}

	return nil
}

func extractFile(zipR *zip.Reader, file, targetFile string) error {
	f, err := zipR.Open(file)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %q in zip", file)
	}
	defer f.Close()

	fOut, err := os.Create(targetFile)
	if err != nil {
		return errors.Wrapf(err, "failed to create target file %q", targetFile)
	}
	defer fOut.Close()

	io.Copy(fOut, f)

	return nil
}

func extractRadarOverview(bspFile, targetDir string) error {
	bspF, err := pathToBsp(bspFile)
	if err != nil {
		return errors.Wrap(err, "failed to read BSP data")
	}

	pakfile, err := centrifuge.Pakfile(bspF)
	if err != nil {
		return errors.Wrap(err, "failed to read pakfile data")
	}

	mapName, err := centrifuge.GetMapName(pakfile)
	if err != nil {
		return errors.Wrap(err, "failed to get map name from pakfile")
	}

	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		return errors.Wrapf(err, "failed to create target dir %q", targetDir)
	}

	ddsPath := fmt.Sprintf("resource/overviews/%s_radar.dds", mapName)
	err = extractFile(pakfile, ddsPath, filepath.Join(targetDir, fmt.Sprintf("%s_radar.dds", mapName)))
	if err != nil {
		return errors.Wrapf(err, "failed to extract file %q from pakfile", ddsPath)
	}

	txtPath := fmt.Sprintf("resource/overviews/%s.txt", mapName)
	err = extractFile(pakfile, txtPath, filepath.Join(targetDir, fmt.Sprintf("%s.txt", mapName)))
	if err != nil {
		return errors.Wrapf(err, "failed to extract file %q from pakfile", txtPath)
	}

	return nil
}

func download(workshopFileID int, targetFile string) error {
	var (
		w   io.Writer
		err error
	)

	if targetFile == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(targetFile)
		if err != nil {
			return errors.Wrapf(err, "failed to create target file: %q", targetFile)
		}
		defer f.Close()

		w = f
	}

	err = steamapi.DownloadWorkshopItem(workshopFileID, w)
	if err != nil {
		return errors.Wrapf(err, "failed to download workshop item with ID %q", workshopFileID)
	}

	return nil
}

func main() {
	var (
		bspFile     string
		bspFileFlag = &cli.StringFlag{
			Name:        "bsp-file",
			Value:       "-",
			Usage:       "BSP file from which to extract data",
			Destination: &bspFile,
		}
		targetFile     string
		targetFileFlag = &cli.StringFlag{
			Name:        "target-file",
			Value:       "-",
			Usage:       "Target file to which to save the data, if applicable",
			Destination: &targetFile,
		}
		targetDir     string
		targetDirFlag = &cli.StringFlag{
			Name:        "target-dir",
			Value:       "out",
			Usage:       "Target directory to which to save the data, if applicable",
			Destination: &targetDir,
		}
		workshopFileID int
	)

	var ()

	app := &cli.App{
		Name:  "bsp-centrifuge",
		Usage: "extract interesting data from BSP (Binary-Space-Partition - source-engine maps) files",
		Commands: []*cli.Command{
			{
				Name:    "pakfile",
				Aliases: []string{"pak"},
				Usage:   "extract the Pakfile zip",
				Flags:   []cli.Flag{bspFileFlag, targetFileFlag},
				Action: func(c *cli.Context) error {
					return extractPakfile(bspFile, targetFile)
				},
			},
			{
				Name:    "radar-image",
				Aliases: []string{"radar"},
				Usage:   "extract radar overview image (.dds file) and the corresponding info (.txt file)",
				Flags:   []cli.Flag{bspFileFlag, targetDirFlag},
				Action: func(c *cli.Context) error {
					return extractRadarOverview(bspFile, targetDir)
				},
			},
			{
				Name:    "download",
				Aliases: []string{"dl"},
				Usage:   "download a file from the steam workshop",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "workshop-file-id",
						Value:       0,
						Usage:       "Steam workshop file id to download, if applicable",
						Destination: &workshopFileID,
					},
				},
				Action: func(c *cli.Context) error {
					return download(workshopFileID, targetFile)
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
