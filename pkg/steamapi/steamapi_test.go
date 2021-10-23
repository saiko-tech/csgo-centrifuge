package steamapi_test

import (
	"fmt"
	"testing"

	"github.com/saiko-tech/csgo-centrifuge/pkg/steamapi"
	"github.com/stretchr/testify/assert"
)

func TestX(t *testing.T) {
	resp, err := steamapi.GetWorkshopFileDetails(472138951)
	assert.NoError(t, err)

	fmt.Println(resp)
}
