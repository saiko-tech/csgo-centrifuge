package core_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/galaco/bsp"
	"github.com/saiko-tech/csgo-overviews-all-versions/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestCore(t *testing.T) {
	const bspFilePath = "../../test/data/de_train.bsp"

	f, err := bsp.ReadFromFile(bspFilePath)
	assert.NoErrorf(t, err, "failed to open BSP file %q", bspFilePath)

	err = core.ExtractRadarImages(f, filepath.Join(os.TempDir(), "radar-overviews"))
	assert.NoError(t, err, "failed to extract radar images from BSP file %q", bspFilePath)
}

type crcTable struct {
	Maps []struct {
		Name       string `json:"map_name"`
		Crc        uint32 `json:"map_crc"`
		WorkshopID int    `json:"workshop_id"`
	} `json:"maps"`
}

func TestDownload(t *testing.T) {
	const crcTablePath = "../../test/data/crc_table.json"

	f, err := os.Open(crcTablePath)
	assert.NoErrorf(t, err, "failed to open CRC table file %q", crcTablePath)
	defer f.Close()

	var tab crcTable
	dec := json.NewDecoder(f)
	err = dec.Decode(&tab)
	assert.NoErrorf(t, err, "failed to decode CRC table %q as JSON", crcTablePath)

	fDownload, err := os.Create(filepath.Join(os.TempDir(), fmt.Sprintf("%s.bsp.zip", tab.Maps[0].Name)))
	assert.NoErrorf(t, err, "failed to create target file for download %q", fDownload.Name())
	defer f.Close()

	err = core.DownloadWorkshopItem(tab.Maps[0].WorkshopID, fDownload)
	assert.NoErrorf(t, err, "failed to download workshop item %q", tab.Maps[0].WorkshopID)
}

func TestE2E(t *testing.T) {
	const crcTablePath = "../../test/data/crc_table.json"

	f, err := os.Open(crcTablePath)
	assert.NoErrorf(t, err, "failed to open CRC table file %q", crcTablePath)
	defer f.Close()

	var tab crcTable
	dec := json.NewDecoder(f)
	err = dec.Decode(&tab)
	assert.NoErrorf(t, err, "failed to decode CRC table %q as JSON", crcTablePath)

	var buf bytes.Buffer

	workshopID := tab.Maps[0].WorkshopID
	err = core.DownloadWorkshopItem(workshopID, &buf)
	assert.NoErrorf(t, err, "failed to download workshop item %q", workshopID)

	b := buf.Bytes()
	r := bytes.NewReader(b)

	zipR, err := zip.NewReader(r, int64(len(b)))
	assert.NoErrorf(t, err, "failed to open zip reader for workshop file with ID %q", workshopID)

	for _, zipF := range zipR.File {
		if filepath.Ext(zipF.Name) == ".bsp" {
			bspR, err := zipF.Open()
			assert.NoErrorf(t, err, "failed to open BSP file in zip %q", zipF.Name)

			bsp, err := bsp.ReadFromStream(bspR)
			assert.NoErrorf(t, err, "failed to read BSP data from zip file stream %q", zipF.Name)

			err = core.ExtractRadarImages(bsp, filepath.Join(os.TempDir(), "radar-overviews"))
			assert.NoErrorf(t, err, "failed to extract radar images from BSP data of file %q", zipF.Name)
		}
	}
}
