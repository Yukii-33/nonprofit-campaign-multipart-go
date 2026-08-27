package main

import (
	"bytes"
	"fmt"
	"net/http"
)

const campaignBucket = "nonprofit-campaign-media"

type CampaignReport struct {
	Campaign               string
	DonorReceiptCount      int
	VolunteerReminderCount int
	Report                 []byte
}

func partCount(size, partSize int) int {
	if size <= 0 {
		return 0
	}
	return (size + partSize - 1) / partSize
}

func uploadReport(report CampaignReport, partSize int) error {
	if len(report.Report) == 0 || partSize <= 0 {
		return fmt.Errorf("report and part size are required")
	}
	key := "campaigns/" + report.Campaign + "/report.csv"
	created, err := infrai.storage.multipart.create(campaignBucket, map[string]any{"key": key})
	if err != nil {
		return err
	}
	uploadID, ok := created["upload_id"].(string)
	if !ok || uploadID == "" {
		return fmt.Errorf("response did not include an upload ID")
	}
	parts := make([]map[string]any, 0, partCount(len(report.Report), partSize))
	for start, number := 0, 1; start < len(report.Report); start, number = start+partSize, number+1 {
		end := start + partSize
		if end > len(report.Report) {
			end = len(report.Report)
		}
		presigned, err := infrai.storage.multipart.presign_part(uploadID, number)
		if err != nil {
			return err
		}
		url, err := signedURL(presigned)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("PUT", url, bytes.NewReader(report.Report[start:end]))
		if err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
		if res.StatusCode >= 300 {
			return fmt.Errorf("part %d upload returned %s", number, res.Status)
		}
		parts = append(parts, map[string]any{"part_number": number, "etag": res.Header.Get("ETag")})
	}
	_, err = infrai.storage.multipart.complete(uploadID, map[string]any{"parts": parts})
	return err
}
