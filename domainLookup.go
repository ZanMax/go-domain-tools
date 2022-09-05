package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	url := os.Args[1]
	if url == "" {
		log.Fatal("Please provide a domain name")
	} else {
		fmt.Println("----------------")
		fmt.Println("-> Domain info")
		fmt.Println("----------------")
		ip := ipLookup(url)
		if ip != "" {
			fmt.Println("IP", ip)
			ns := nsLookup(url)
			fmt.Println("NS", ns)
			mx := mxLookup(url)
			fmt.Println("MX", mx)
			cname := cnameLookup(url)
			fmt.Println("CNAME", cname)
			txt := txtLookup(url)
			fmt.Println("TXT", txt)
		} else {
			fmt.Println("No IP found")
		}
	}
}

func nsLookup(domain string) []string {
	ns, err := net.LookupNS(domain)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	var nsList []string
	for _, n := range ns {
		nsList = append(nsList, n.Host)
	}
	return nsList
}

func ipLookup(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Println(err)
		return ""
	}
	return ips[0].String()
}

func mxLookup(domain string) []string {
	mx, err := net.LookupMX(domain)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	var mxList []string
	for _, m := range mx {
		mxList = append(mxList, m.Host)
	}
	return mxList
}

func cnameLookup(domain string) []string {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	return []string{cname}
}

func txtLookup(domain string) []string {
	txt, err := net.LookupTXT(domain)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	return txt
}
