package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
)

var wg sync.WaitGroup

func main() {
	log.Printf("\t%v\t%v\t%v\t%v\t%v\t\n", "min", "max", "avg", "proto", "host")
	wg.Add(3)
	go bench("https://dns.google.com/resolve", &h2quic.RoundTripper{})
	go bench("https://dns.google.com/resolve", nil)
	go bench("https://cloudflare-dns.com/dns-query", nil)
	wg.Wait()
}

func bench(host string, rt http.RoundTripper) {
	quic := &http.Client{Transport: rt}
	query := host + "?ct=application/dns-json&name=apple.com.&type=A"
	r, _ := quic.Get(query)
	r.Body.Close()
	min := time.Hour
	max := time.Duration(0)
	total := time.Duration(0)
	samples := 500
	for i := 0; i < samples; i++ {
		start := time.Now()
		resp, err := quic.Get(query)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if resp.StatusCode != 200 {
			log.Println(resp.StatusCode)
		}
		io.Copy(ioutil.Discard, resp.Body)
		l := time.Now().Sub(start)
		resp.Body.Close()
		if l > max {
			max = l
		}
		if l < min {
			min = l
		}
		total += l
	}
	proto := "tcp"
	if rt != nil {
		proto = "quic"
	}
	log.Printf("\t%v\t%v\t%v\t%v\t%v\t\n", min.Truncate(time.Millisecond), max.Truncate(time.Millisecond), (total / time.Duration(samples)).Truncate(time.Millisecond), proto, host)
	wg.Done()
}
