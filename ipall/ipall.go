/*
    REVISION HISTORY
     2 Nov 20 -- I started Networking in Go video from Packtpub, and these little programs were part of its code.
                   I'm using it to combine ipinfo with this code to make one utility.  It will do what ipinfo does
                   if a dotted quartet is input, and it will report the domain name doing a reverse lookup.
                   If a domain name is provided, it will do what nslookup does now, on both linux and windows.
		   Since it's possible for a domain name to begin with a number, I retained the option of specifically setting the
		   ip or host string
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"unicode"
	"unicode/utf8"
)

const LastAltered = "2 Nov 2020"

var ip string
var host string
var ns bool
var mx bool
var txt bool
var cname bool

const (
	ipMode = iota
	dnsNameMode
)

var ipDNSstate int

func init() {
	flag.StringVar(&ip, "ip", "", "IP address for DNS operation")
	flag.StringVar(&host, "host", "", "Host address for DNS operation")
	flag.BoolVar(&ns, "ns", false, "Host name server lookup")
	flag.BoolVar(&mx, "mx", false, "Host domain mail server lookup")
	flag.BoolVar(&txt, "txt", false, "Host domain TXT lookup")
	flag.BoolVar(&cname, "cname", false, "Host CNAME lookup")
}

type lsdns struct {
	resolver *net.Resolver
}

func main() {
	flag.Parse()
	ls := newLsdns()

	var arg string
	if ip == "" && host == "" {
		if flag.NArg() == 0 {
			fmt.Println("Usage: ipall [ip dotted quarted] | [domain name string]")
			os.Exit(1)
		}

		arg = flag.Arg(0)
		r, _ := utf8.DecodeRuneInString(arg)
		if unicode.IsDigit(r) {
			ip = arg
		} else {
			host = arg
		}
	}

	switch {
	case ip != "":
		ls.reverseLkp(ip)

		URL := "http://ipinfo.io/" + arg
		data, err := http.Get(URL)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			defer data.Body.Close()
			_, err := io.Copy(os.Stdout, data.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println()
		}

	case host != "":
		ls.nsLkp(host)
		ls.mxLkp(host)
		if txt {              // very long and not very helpful.
			ls.txtLkp(host)  // I'll not display it by default.
		}
		ls.cnameLkp(host)
		ls.hostLkp(host)

	default:
		fmt.Println("flag ip or host must be provided")
		os.Exit(1)
	}
} // end main

func newLsdns() *lsdns {
	return &lsdns{net.DefaultResolver}
}

func (ls *lsdns) reverseLkp(ip string) error {
	names, err := ls.resolver.LookupAddr(context.Background(), ip)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("Reverse lookup")
	fmt.Println("--------------")
	for _, name := range names {
		fmt.Println(name)
	}
	fmt.Println()
	return nil
}

func (ls *lsdns) hostLkp(host string) error {
	addrs, err := ls.resolver.LookupHost(context.Background(), host)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("Host lookup")
	fmt.Println("-----------")
	for _, addr := range addrs {
		fmt.Printf("%-30s%-20s\n", host, addr)
	}
	fmt.Println()
	return nil
}

func (ls *lsdns) nsLkp(host string) error {
	nses, err := ls.resolver.LookupNS(context.Background(), host)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("NS lookup")
	fmt.Println("---------")
	for _, ns := range nses {
		fmt.Printf("%-25s%-20s\n", host, ns.Host)
	}
	fmt.Println()
	return nil
}

func (ls *lsdns) mxLkp(host string) error {
	mxes, err := ls.resolver.LookupMX(context.Background(), host)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("MX lookup")
	fmt.Println("---------")
	for _, mx := range mxes {
		fmt.Printf("%-17s%-11s\n", host, mx.Host)
	}
	fmt.Println()
	return nil
}

func (ls *lsdns) txtLkp(host string) error {
	txts, err := ls.resolver.LookupTXT(context.Background(), host)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("TXT lookup")
	fmt.Println("---------")
	for _, txt := range txts {
		fmt.Printf("%-17s%-11s\n", host, txt)
	}
	fmt.Println()
	return nil
}

func (ls *lsdns) cnameLkp(host string) error {
	name, err := ls.resolver.LookupCNAME(context.Background(), host)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("CNAME lookup")
	fmt.Println("------------")
	fmt.Printf("%s: %s\n", host, name)
	fmt.Println()
	return nil
}

/*
func main() {
	flag.StringVar(&host, "host", "localhost", "host name to resolve")
	flag.Parse()

	addrs, err := net.LookupHost(host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(addrs)
}
func main() {
	flag.StringVar(&addr, "addr", "127.0.0.1", "host address to lookup")
	flag.Parse()

	names, err := net.LookupAddr(addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(names)
}

*/
