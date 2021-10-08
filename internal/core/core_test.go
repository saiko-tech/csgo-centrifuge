package core_test

import (
	"testing"

	"github.com/galaco/bsp"
	"github.com/saiko-tech/csgo-overviews-all-versions/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestCore(t *testing.T) {
	const bspFilePath = "../../test/data/de_train.bsp"

	f, err := bsp.ReadFromFile(bspFilePath)
	assert.NoErrorf(t, err, "failed to open BSP file %q", bspFilePath)

	err = core.ExtractRadarImages(f)
	assert.NoError(t, err, "failed to extract radar images from BSP file %q", bspFilePath)
}
