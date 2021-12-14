package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/galaco/bsp"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/saiko-tech/csgo-centrifuge/pkg/bsputil"
	"github.com/saiko-tech/csgo-centrifuge/pkg/crc"
	"github.com/saiko-tech/csgo-centrifuge/pkg/steamapi"
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

func extractPakfile(bspPath, outPath string) error {
	bspF, err := pathToBsp(bspPath)
	if err != nil {
		return errors.Wrap(err, "failed to read BSP data")
	}

	b := bspF.RawLump(bsp.LumpPakfile).RawContents()
	r := bytes.NewReader(b)

	var w io.Writer
	if outPath == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(outPath)
		if err != nil {
			return errors.Wrapf(err, "failed to create out file: %q", outPath)
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

func crc32Bsp(bspPath string) error {
	bspF, err := pathToBsp(bspPath)
	if err != nil {
		return errors.Wrap(err, "failed to read BSP data")
	}

	crc, err := bspF.CRC32()
	if err != nil {
		return errors.Wrap(err, "failed to calculate CRC32 sum for BSP file")
	}

	fmt.Println(crc)

	return nil
}

func extractFile(zipR *zip.Reader, file, outPath string) error {
	f, err := zipR.Open(file)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %q in zip", file)
	}
	defer f.Close()

	fOut, err := os.Create(outPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create out file %q", outPath)
	}
	defer fOut.Close()

	io.Copy(fOut, f)

	return nil
}

func extractRadarOverview(bspPath, outDirPath string) error {
	bspF, err := pathToBsp(bspPath)
	if err != nil {
		return errors.Wrap(err, "failed to read BSP data")
	}

	pakfile, err := bsputil.Pakfile(bspF)
	if err != nil {
		return errors.Wrap(err, "failed to read pakfile data")
	}

	mapName, err := bsputil.GetMapName(pakfile)
	if err != nil {
		return errors.Wrap(err, "failed to get map name from pakfile")
	}

	err = os.MkdirAll(outDirPath, 0777)
	if err != nil {
		return errors.Wrapf(err, "failed to create out dir %q", outDirPath)
	}

	ddsPath := fmt.Sprintf("resource/overviews/%s_radar.dds", mapName)
	err = extractFile(pakfile, ddsPath, filepath.Join(outDirPath, fmt.Sprintf("%s_radar.dds", mapName)))
	if err != nil {
		return errors.Wrapf(err, "failed to extract file %q from pakfile", ddsPath)
	}

	txtPath := fmt.Sprintf("resource/overviews/%s.txt", mapName)
	err = extractFile(pakfile, txtPath, filepath.Join(outDirPath, fmt.Sprintf("%s.txt", mapName)))
	if err != nil {
		return errors.Wrapf(err, "failed to extract file %q from pakfile", txtPath)
	}

	return nil
}

func download(workshopFileID int, outPath string) error {
	var (
		w   io.Writer
		err error
	)

	if outPath == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(outPath)
		if err != nil {
			return errors.Wrapf(err, "failed to create out file: %q", outPath)
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

func extractCRCTable(engineClientSOPath, outPath string) error {
	r, err := os.Open(engineClientSOPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open enginge_client.so file %q", engineClientSOPath)
	}

	var w io.Writer
	if outPath == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(outPath)
		if err != nil {
			return errors.Wrapf(err, "failed to create output file: %q", outPath)
		}
		defer f.Close()

		w = f
	}

	tab, err := crc.ExtractCRCTable(r)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return errors.Wrapf(err, "failed to extract CRC table from engine_client.so file %q", engineClientSOPath)
	}

	err = json.NewEncoder(w).Encode(tab)
	if err != nil {
		return errors.Wrapf(err, "failed to encode CRC table as JSON to output file %q", outPath)
	}

	return nil
}

func main() {
	var (
		inFile     string
		inFileFlag = &cli.StringFlag{
			Name:        "in-file",
			Value:       "-",
			Usage:       "Input file from which to extract data",
			Destination: &inFile,
		}
		outFile     string
		outFileFlag = &cli.StringFlag{
			Name:        "out-file",
			Value:       "-",
			Usage:       "Output file to which to save the data",
			Destination: &outFile,
		}
		outDir     string
		outDirFlag = &cli.StringFlag{
			Name:        "output-dir",
			Value:       "out",
			Usage:       "Output directory to which to save the data",
			Destination: &outDir,
		}
		workshopFileID int
	)

	var ()

	app := &cli.App{
		Name:  "csgo-centrifuge",
		Usage: "process CSGO game files in (hopefully) interesting ways",
		Commands: []*cli.Command{
			{
				Name:    "crc-table",
				Aliases: []string{"crc"},
				Usage:   "extract the CRC table from bin/linux64/engine_client.so",
				Flags:   []cli.Flag{inFileFlag, outFileFlag},
				Action: func(c *cli.Context) error {
					return extractCRCTable(inFile, outFile)
				},
			},
			{
				Name:  "bsp",
				Usage: "extract interesting data from BSP (Binary-Space-Partition - source-engine maps) files",
				Subcommands: []*cli.Command{
					{
						Name:    "pakfile",
						Aliases: []string{"pak"},
						Usage:   "extract the Pakfile zip",
						Flags:   []cli.Flag{inFileFlag, outFileFlag},
						Action: func(c *cli.Context) error {
							return extractPakfile(inFile, outFile)
						},
					},
					{
						Name:    "radar-image",
						Aliases: []string{"radar"},
						Usage:   "extract radar overview image (.dds file) and the corresponding info (.txt file)",
						Flags:   []cli.Flag{inFileFlag, outDirFlag},
						Action: func(c *cli.Context) error {
							return extractRadarOverview(inFile, outDir)
						},
					},
					{
						Name:    "crc32",
						Usage:   "calculate CRC32 sum of .bsp file",
						Flags:   []cli.Flag{inFileFlag},
						Action: func(c *cli.Context) error {
							return crc32Bsp(inFile)
						},
					},
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
					outFileFlag,
				},
				Action: func(c *cli.Context) error {
					return download(workshopFileID, outFile)
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
