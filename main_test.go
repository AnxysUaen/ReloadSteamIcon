package main

import (
	"fmt"
	"testing"
)

var ScanSteamInstallation = scanSteamInstallation
var ScanSteamGameId = scanSteamGameId

func TestScanSteamInstallation(t *testing.T) {
	steamInstalledFolder = "./test_files"
	ScanSteamInstallation()
}

func TestScanSteamGameId(t *testing.T) {
	steamInstalledFolder = "./test_files"
	ScanSteamGameId()
	fmt.Println(installedGameIDList)
}
