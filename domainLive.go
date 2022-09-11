package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxGoRoutines         = 30
	Timeout               = 10 * time.Second
	ResponseHeaderTimeout = 10 * time.Second
	TLSHandshakeTimeout   = 10 * time.Second
	MaxIdleConns          = 100
	MaxConnsPerHost       = 100
	MaxIdleConnsPerHost   = 100
	InsecureSkipVerify    = true
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter a domain file")
		return
	}
	domainFile := os.Args[1]
	if _, err := os.Stat(domainFile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("File does not exist")
		return
	}
	f, err := os.Open(domainFile)
	check(err)
	defer func(f *os.File) {
		err := f.Close()
		check(err)
	}(f)
	scanner := bufio.NewScanner(f)

	var urls []string

	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	fmt.Println("Total domains: ", len(urls))
	uniqueUrls := unique(urls)
	fmt.Println("Total uniq domains: ", len(uniqueUrls))

	guard := make(chan struct{}, maxGoRoutines)

	var wg sync.WaitGroup
	for _, u := range uniqueUrls {
		guard <- struct{}{}
		wg.Add(1)
		go func(url string) {

			defer wg.Done()

			httpCheckHost(url)
			<-guard
		}(u)
	}
	wg.Wait()

}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func writeFile(str string) {
	f, err := os.OpenFile("domains-live",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	if _, err := f.WriteString(str + "\n"); err != nil {
		log.Println(err)
	}
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func TCPCheckHost(domain string) bool {
	timeout := 10 * time.Second
	_, httpError := net.DialTimeout("tcp", domain+":80", timeout)
	_, httpsError := net.DialTimeout("tcp", domain+":443", timeout)

	if httpError != nil || httpsError != nil {
		if httpError != nil {
			log.Println("Error: ", httpError)
			return false
		} else {
			log.Println("Error: ", httpsError)
			return false
		}
	}
	return true
}

func LookupIP(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return ips[0].String()
}

func httpCheckHost(domain string) {
	var wg sync.WaitGroup
	wg.Add(2)
	result := make(chan int)
	go makeRequest(&wg, domain, "http", result)
	go makeRequest(&wg, domain, "https", result)

	HttpStatusCode, HttpsStatusCode := <-result, <-result
	wg.Wait()

	if HttpStatusCode == 0 && HttpsStatusCode == 0 {
		return
	}
	if HttpStatusCode == 200 || HttpStatusCode == 301 || HttpStatusCode == 302 {
		writeFile(domain)
	} else if HttpsStatusCode == 200 || HttpsStatusCode == 301 || HttpsStatusCode == 302 {
		writeFile(domain)
	}
}

func makeRequest(wg *sync.WaitGroup, domain string, protocol string, result chan int) {
	defer wg.Done()

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = MaxIdleConns
	t.MaxConnsPerHost = MaxConnsPerHost
	t.MaxIdleConnsPerHost = MaxIdleConnsPerHost
	t.ResponseHeaderTimeout = ResponseHeaderTimeout
	t.TLSClientConfig.InsecureSkipVerify = InsecureSkipVerify
	t.TLSHandshakeTimeout = TLSHandshakeTimeout

	client := &http.Client{
		Timeout:   Timeout,
		Transport: t,
	}

	if protocol == "http" && !strings.HasPrefix(domain, "http://") {
		domain = "http://" + domain
	} else if protocol == "https" && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}

	req, err := http.NewRequest("GET", domain, nil)
	if err != nil {
		fmt.Println(err)
		result <- 0
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		result <- 0
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)
	req.Close = true

	result <- resp.StatusCode
}
