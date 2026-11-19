package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

type phoneLink struct{ network, host string }

func watchPhoneLinks(ctx context.Context, bound *net.TCPAddr, token string) {
	if bound.IP.IsLoopback() {
		fmt.Println("Сервер доступен только на ноутбуке:", bound)
		return
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	monitorLinks(ctx, ticker.C, func() []phoneLink { return phoneLinks(bound.IP) }, func(links []phoneLink) {
		if len(links) == 0 {
			fmt.Println("Ожидание подключения к Wi-Fi или Ethernet…")
			return
		}
		for _, link := range links {
			printPhoneLink(link.network, link.host, bound.Port, token)
		}
	})
}

func monitorLinks(ctx context.Context, ticks <-chan time.Time, scan func() []phoneLink, display func([]phoneLink)) {
	var previous map[phoneLink]bool
	for {
		links := scan()
		current := make(map[phoneLink]bool, len(links))
		var added []phoneLink
		for _, link := range links {
			if !previous[link] && !current[link] {
				added = append(added, link)
			}
			current[link] = true
		}
		if len(added) > 0 || (len(current) == 0 && (previous == nil || len(previous) > 0)) {
			display(added)
		}
		previous = current
		select {
		case <-ctx.Done():
			return
		case <-ticks:
		}
	}
}

func phoneLinks(bound net.IP) []phoneLink {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var links []phoneLink
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if bound.IsUnspecified() && !physicalLAN(iface.Index) {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil || !ip.IsGlobalUnicast() || (bound.IsUnspecified() && ip.To4() == nil) || (!bound.IsUnspecified() && !bound.Equal(ip)) {
				continue
			}
			links = append(links, phoneLink{iface.Name, ip.String()})
		}
	}
	return links
}

func printPhoneLink(network, host string, port int, token string) {
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + "/"
	fmt.Printf("\n%s  %s\n", network, url)
	code, err := qrcode.New(url+"#pair="+token, qrcode.Medium)
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
