package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	httpClient = http.DefaultClient
)

func sendWatchCenterMessage(watchCenterID, message string) error {
	resp, err := httpClient.PostForm(fmt.Sprintf("%s", watchCenterURL), url.Values{
		"to":  []string{watchCenterID},
		"msg": []string{message},
	})
	if err != nil {
		log.Printf("sendind a message has failed, error = %s\n", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("sendind a message has failed, status code of sending a message is not 200, got status code %d\n", resp.StatusCode)
		return fmt.Errorf("not expected status code")
	}

	return nil
}
