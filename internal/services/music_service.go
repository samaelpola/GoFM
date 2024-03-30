package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samaelpola/GoFM/internal/models"
	"github.com/tcolgate/mp3"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	GORAP  = "GO-RAP"
	GOPOP  = "GO-POP"
	GOROCK = "GO-ROCK"
	GOSLOW = "GO-SLOW"
	GOGEN  = "GO-GEN"
)

func GetListOfStation() [5]string {
	return [5]string{GOGEN, GOROCK, GORAP, GOPOP, GOSLOW}
}

func prepareRequest(method, uri string) (*http.Request, error) {
	request, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/musics/%s", os.Getenv("API_ENDPOINT"), uri),
		nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set(
		"Authorization",
		fmt.Sprintf("Bearer %s", os.Getenv("TOKEN")),
	)

	return request, nil
}

func GetMusics(musicType string) ([]models.Music, error) {
	request, err := prepareRequest("GET", musicType)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(content))
	}

	var musics []models.Music
	if err := json.Unmarshal(content, &musics); err != nil {
		return nil, err
	}

	return musics, nil
}

func GetAudio(musicID int) (*bytes.Reader, error) {
	request, err := prepareRequest(
		"GET",
		fmt.Sprintf("%d/audio", musicID),
	)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(content))
	}

	return bytes.NewReader(content), nil
}

func GetTrackDuration(fd io.Reader) time.Duration {
	var t float64
	var f mp3.Frame
	d := mp3.NewDecoder(fd)
	skipped := 0

	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}

		t = t + f.Duration().Seconds()
	}

	return time.Duration(t * float64(time.Second))
}

func GetCurrentMusicType() string {
	now := time.Now()
	hour := now.Hour()
	var musicType string

	switch {
	case hour >= 0 && hour < 6:
		musicType = GORAP
	case hour >= 6 && hour < 12:
		musicType = GOROCK
	case hour >= 12 && hour < 18:
		musicType = GOPOP
	case hour >= 18 && hour < 24:
		musicType = GOSLOW
	default:
		musicType = ""
	}

	return musicType
}
