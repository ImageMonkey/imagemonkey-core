package main

import (
	"math/rand"
	"time"
	"strings"
	"io/ioutil"
    "github.com/bbernhard/imghash"
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "net"
    "bytes"
    "net/http"
)

type Report struct {
    Reason string `json:"reason"`
}

type Label struct {
    Name string `json:"name"`
}

type ContributionsPerCountryRequest struct {
    CountryCode string `json:"country_code"`
    Type string `json:"type"`
}

func use(vals ...interface{}) {
    for _, val := range vals {
        _ = val
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func pick(args ...interface{}) []interface{} {
    return args
}

/*
 * Loads all data in memory.
 * If file gets too big, refactor it!
 */
func getWordLists(path string) ([]Label, error) {
    var lines []string
    var labels []Label
	data, err := ioutil.ReadFile(path)
    if(err != nil){
        return labels, err
    }
    lines = strings.Split(string(data), "\r\n")
    for _, v := range lines {
        var label Label
        label.Name = v
        labels = append(labels, label)
    }

    return labels, nil
}

func getStrWordLists(path string) ([]string, error) {
    var lines []string
    data, err := ioutil.ReadFile(path)
    if(err != nil){
        return lines, err
    }
    lines = strings.Split(string(data), "\r\n")
    return lines, nil
}

func hashImage(file io.Reader) (uint64, error){
    img, _, err := image.Decode(file)
    if err != nil {
        return 0, err
    }

    return imghash.Average(img), nil
}




//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
    start net.IP
    end net.IP
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
    // strcmp type byte comparison
    if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
        return true
    }
    return false
}

var privateRanges = []ipRange{
    ipRange{
        start: net.ParseIP("10.0.0.0"),
        end:   net.ParseIP("10.255.255.255"),
    },
    ipRange{
        start: net.ParseIP("100.64.0.0"),
        end:   net.ParseIP("100.127.255.255"),
    },
    ipRange{
        start: net.ParseIP("172.16.0.0"),
        end:   net.ParseIP("172.31.255.255"),
    },
    ipRange{
        start: net.ParseIP("192.0.0.0"),
        end:   net.ParseIP("192.0.0.255"),
    },
    ipRange{
        start: net.ParseIP("192.168.0.0"),
        end:   net.ParseIP("192.168.255.255"),
    },
    ipRange{
        start: net.ParseIP("198.18.0.0"),
        end:   net.ParseIP("198.19.255.255"),
    },
}


// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
    // my use case is only concerned with ipv4 atm
    if ipCheck := ipAddress.To4(); ipCheck != nil {
        // iterate over all our ranges
        for _, r := range privateRanges {
            // check if this ip is in a private range
            if inRange(r, ipAddress){
                return true
            }
        }
    }
    return false
}


func getIPAddress(r *http.Request) string {
    for _, h := range []string{"X-Forwarded-For", "X-Real-IP"} {
        addresses := strings.Split(r.Header.Get(h), ",")
        // march from right to left until we get a public address
        // that will be the address right before our proxy.
        for i := len(addresses) -1 ; i >= 0; i-- {
            ip := strings.TrimSpace(addresses[i])

            // header can contain spaces too, strip those out.
            realIP := net.ParseIP(ip)
            if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
                // bad address, go to next
                continue
            }
            return ip
        }
    }
    return ""
}

/*func prettyPrintJSON(b []byte) ([]byte, error) {
    var out bytes.Buffer
    err := JSON.Indent(&out, b, "", "    ")
    return out.Bytes(), err
}*/