package main

import "testing"

func TestPartCountKeepsCampaignReportPartsBounded(t *testing.T) {
	if got := partCount(10*1024*1024+1, 5*1024*1024); got != 3 {
		t.Fatalf("partCount = %d, want 3", got)
	}
	if got := partCount(5*1024*1024, 5*1024*1024); got != 1 {
		t.Fatalf("partCount = %d, want 1", got)
	}
}
