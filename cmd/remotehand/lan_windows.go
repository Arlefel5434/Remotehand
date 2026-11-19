package main

import "golang.org/x/sys/windows"

func physicalLAN(index int) bool {
	row := windows.MibIfRow2{InterfaceIndex: uint32(index)}
	if windows.GetIfEntry2Ex(windows.MibIfEntryNormalWithoutStatistics, &row) != nil {
		return false
	}
	const hardwareInterface = 1
	return row.InterfaceAndOperStatusFlags&hardwareInterface != 0 && (row.Type == 6 || row.Type == 71)
}
