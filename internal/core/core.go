package core

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/galaco/bsp"
	"github.com/pkg/errors"
)

func extractRadarImage(f *zip.File, targetDir string) error {
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

func ExtractRadarImages(f *bsp.Bsp, targetDir string) error {
	err := os.MkdirAll(targetDir, 0744)
	if err != nil {
		return errors.Wrapf(err, "failed to create target dir %q", targetDir)
	}

	b := f.RawLump(bsp.LumpPakfile).RawContents()
	r := bytes.NewReader(b)
	zipR, err := zip.NewReader(r, int64(len(b)))
	if err != nil {
		return errors.Wrapf(err, "failed to open Pakfile lump")
	}

	for _, pakF := range zipR.File {
		if filepath.Ext(pakF.Name) == ".dds" {
			err := extractRadarImage(pakF, targetDir)
			if err != nil {
				return errors.Wrapf(err, "failed to extract radar image %q", pakF.Name)
			}
		}
	}

	return nil
}

type GetPublishedFileDetailsResponse struct {
	Response struct {
		Result               int `json:"result"`
		Resultcount          int `json:"resultcount"`
		Publishedfiledetails []struct {
			Publishedfileid       string `json:"publishedfileid"`
			Result                int    `json:"result"`
			Creator               string `json:"creator"`
			CreatorAppID          int    `json:"creator_app_id"`
			ConsumerAppID         int    `json:"consumer_app_id"`
			Filename              string `json:"filename"`
			FileSize              int    `json:"file_size"`
			FileURL               string `json:"file_url"`
			HcontentFile          string `json:"hcontent_file"`
			PreviewURL            string `json:"preview_url"`
			HcontentPreview       string `json:"hcontent_preview"`
			Title                 string `json:"title"`
			Description           string `json:"description"`
			TimeCreated           int    `json:"time_created"`
			TimeUpdated           int    `json:"time_updated"`
			Visibility            int    `json:"visibility"`
			Banned                int    `json:"banned"`
			BanReason             string `json:"ban_reason"`
			Subscriptions         int    `json:"subscriptions"`
			Favorited             int    `json:"favorited"`
			LifetimeSubscriptions int    `json:"lifetime_subscriptions"`
			LifetimeFavorited     int    `json:"lifetime_favorited"`
			Views                 int    `json:"views"`
			Tags                  []struct {
				Tag string `json:"tag"`
			} `json:"tags"`
		} `json:"publishedfiledetails"`
	} `json:"response"`
}

func GetWorkshopFileDetails(workshopID int) (GetPublishedFileDetailsResponse, error) {
	payload := url.Values{
		"itemcount":           []string{"1"},
		"publishedfileids[0]": []string{fmt.Sprint(workshopID)},
	}

	resp, err := http.PostForm("http://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1", payload)
	if err != nil {
		return GetPublishedFileDetailsResponse{}, err
	}

	defer resp.Body.Close()

	var respData GetPublishedFileDetailsResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&respData)
	if err != nil {
		return GetPublishedFileDetailsResponse{}, err
	}

	return respData, nil
}

func DownloadWorkshopItem(workshopID int, w io.Writer) error {
	details, err := GetWorkshopFileDetails(workshopID)
	if err != nil {
		return err
	}

	resp, err := http.Get(details.Response.Publishedfiledetails[0].FileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)

	return err
}
