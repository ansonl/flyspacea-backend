package main

import (
	"fmt"
	"log"
	"strconv"
	linuxproc "github.com/c9s/goprocinfo/linux"
)

var validTerminals int
var terminalsWith72HRAlbum int
var noErrorTerminals int
var foundFlightsTerminals int

var photosFound int
var photosFoundDateHeader int
var photosProcessed int

var noMatchDateHeaderInputs []string

//live stats
var liveTotalTerminals int
var liveTerminalsUpdated int

func resetStatistics() {
	validTerminals = 0
	terminalsWith72HRAlbum = 0
	noErrorTerminals = 0
	foundFlightsTerminals = 0

	photosFound = 0
	photosFoundDateHeader = 0
	photosProcessed = 0
}

func incrementValidTerminals() {
	validTerminals++
}

func incrementTerminalsWith72HRAlbum() {
	terminalsWith72HRAlbum++
}

func incrementNoErrorTerminals() {
	noErrorTerminals++
}

func incrementFoundFlightsTerminals() {
	foundFlightsTerminals++
}

func incrementPhotosFound() {
	photosFound++
}

func incrementPhotosProcessed() {
	photosProcessed++
}

func incrementPhotosFoundDateHeader() {
	photosFoundDateHeader++
}

func displayStatistics() {
	log.Printf(`Valid Terminals online%v
		Terminals w/ found 72HR Album %v
		Terminals w/ no errors (incl any old photos) %v
		Terminals w/ found flights %v
		Photos Found %v
		Photos Found Date Header %v
		Photos Processed %v`,
		validTerminals,
		terminalsWith72HRAlbum,
		noErrorTerminals,
		foundFlightsTerminals,
		photosFound,
		photosFoundDateHeader,
		photosProcessed)
	//log.Printf("No match date header inputs %v\n", noMatchDateHeaderInputs)
}

func statisticsString() string {
	return fmt.Sprintf(`Valid Terminals online %v
Terminals w/ found 72HR Album %v
Terminals w/ no errors (incl any old photos) %v
Terminals w/ found flights %v
Photos Found %v
Photos Found Date Header %v
Photos Processed %v`,
		validTerminals,
		terminalsWith72HRAlbum,
		noErrorTerminals,
		foundFlightsTerminals,
		photosFound,
		photosFoundDateHeader,
		photosProcessed)
}

//live stats
func setLiveTotalTerminals(total int) {
	liveTotalTerminals = total
}

func incrementLiveTerminalsUpdated() {
	liveTerminalsUpdated++
}

func liveStatisticsString() (live string) {

	//Print progress bar
	live += fmt.Sprintf("%v/%v terminals update progress.\n", strconv.Itoa(liveTerminalsUpdated), strconv.Itoa(liveTotalTerminals))

	live += "⌜"

	for i := 0; i < liveTotalTerminals; i++ {
		live += "-"
	}

	live += "⌝\n"
	live += "|"

	for i := 0; i < liveTerminalsUpdated; i++ {
		live += "█"
	}

	for i := 0; i < liveTotalTerminals-liveTerminalsUpdated; i++ {
		live += " "
	}

	live += "|\n"
	live += "⌞"

	for i := 0; i < liveTotalTerminals; i++ {
		live += "-"
	}
	live += "⌟\n"

	live += "CPU Info\n"
	/*
	if cpusInfo, err := linuxproc.ReadCPUInfo("/proc/stat"); err != nil {
		live += "stat read failed " + err.Error() + "\n"
	} else {
		live += "NumCPU\t" + strconv.Itoa(int(cpusInfo.NumCPU())) + "\n"
		live += "NumCore\t" + strconv.Itoa(int(cpusInfo.NumCore())) + "\n"
		live += "NumPhysicalCPU\t" + strconv.Itoa(int(cpusInfo.NumPhysicalCPU())) + "\n"
	}
	*/

	if cpusStat, err := linuxproc.ReadStat("/proc/stat"); err != nil {
		live += "stat read failed " + err.Error() + "\n"
	} else {
		live += "User\t" + strconv.Itoa(int(cpusStat.CPUStatAll.User)) + "\n"
		live += "Nice\t" + strconv.Itoa(int(cpusStat.CPUStatAll.Nice)) + "\n"
		live += "System\t" + strconv.Itoa(int(cpusStat.CPUStatAll.System)) + "\n"
		live += "Idle\t" + strconv.Itoa(int(cpusStat.CPUStatAll.Idle)) + "\n"
	}

	live += "Mem Info\n"
	if memsInfo, err := linuxproc.ReadMemInfo("/proc/stat"); err != nil {
		live += "stat read failed " + err.Error() + "\n"
	} else {
		live += "MemTotal\t" + strconv.Itoa(int(memsInfo.MemTotal)) + "\n"
		live += "MemFree\t" + strconv.Itoa(int(memsInfo.MemFree)) + "\n"
		live += "MemAvailable\t" + strconv.Itoa(int(memsInfo.MemAvailable)) + "\n"
		live += "Cached\t" + strconv.Itoa(int(memsInfo.Cached)) + "\n"

		live += "SwapTotal\t" + strconv.Itoa(int(memsInfo.SwapTotal)) + "\n"
		live += "SwapFree\t" + strconv.Itoa(int(memsInfo.SwapFree)) + "\n"
		live += "SwapCached\t" + strconv.Itoa(int(memsInfo.SwapCached)) + "\n"
	}
	

	return
}
