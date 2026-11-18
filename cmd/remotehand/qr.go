package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

func showPhoneLinks(addr net.Addr) {
	bound := addr.(*net.TCPAddr)
	if !bound.IP.IsUnspecified() {
		if bound.IP.IsLoopback() {
			fmt.Println("Сервер доступен только на ноутбуке:", bound)
			return
		}
		printPhoneLink("Remotehand", bound.IP.String(), bound.Port)
		return
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Не удалось определить адрес сети:", err)
		return
	}
	found := false
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil || ip.To4() == nil || !ip.IsGlobalUnicast() {
				continue
			}
			printPhoneLink(iface.Name, ip.String(), bound.Port)
			found = true
		}
	}
	if !found {
		fmt.Println("Подключите ноутбук к Wi-Fi или Ethernet и перезапустите Remotehand.")
	}
}

func printPhoneLink(network, host string, port int) {
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + "/"
	fmt.Printf("\n%s  %s\n", network, url)
	code, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		fmt.Println(err)
		return
	}
	// A light background and quiet zone on a dark terminal. Each character
	// represents two vertical modules to keep the code within a console window.
	rows := code.Bitmap()
	var out strings.Builder
	for y := 0; y < len(rows); y += 2 {
		for x, black := range rows[y] {
			bottomBlack := y+1 < len(rows) && rows[y+1][x]
			switch {
			case black && bottomBlack:
				out.WriteByte(' ')
			case black:
				out.WriteRune('▄')
			case bottomBlack:
				out.WriteRune('▀')
			default:
				out.WriteRune('█')
			}
		}
		out.WriteByte('\n')
	}
	fmt.Print(out.String())
}
