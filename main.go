package main

import (
	"fmt"
	"log"
)

func main() {
	report := CampaignReport{
		Campaign:               "spring-food-drive",
		DonorReceiptCount:      1842,
		VolunteerReminderCount: 97,
		Report:                 []byte("campaign,spring-food-drive\ndonor_receipts,1842\nvolunteer_reminders,97\n"),
	}
	if err := uploadReport(report, 5*1024*1024); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("uploaded report for %s (%d bytes)\n", report.Campaign, len(report.Report))
}
