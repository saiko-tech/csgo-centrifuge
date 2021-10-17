package steamapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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
