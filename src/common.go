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
    "encoding/json"
)

type Report struct {
    Reason string `json:"reason"`
}

type Label struct {
    Name string `json:"name"`
}

type LabelMeEntry struct {
    Label string `json:"label"` 
    Sublabels []string `json:"sublabels"`
}


type ContributionsPerCountryRequest struct {
    CountryCode string `json:"country_code"`
    Type string `json:"type"`
}


type MetaLabelMapEntry struct {
    Description string  `json:"description"`
    Name string `json:"name"`
}

type LabelMapEntry struct {
    Description string  `json:"description"`
    LabelMapEntries map[string]LabelMapEntry  `json:"has"`
}

type LabelMap struct {
    LabelMapEntries map[string]LabelMapEntry `json:"labels"`
    MetaLabelMapEntries map[string]MetaLabelMapEntry  `json:"metalabels"`
}

type LabelValidationEntry struct {
    Label string  `json:"label"`
    Sublabel string `json:"sublabel"`
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

func getLabelMap(path string) (map[string]LabelMapEntry, []string, error) {
    var words []string
    var labelMap LabelMap

    data, err := ioutil.ReadFile(path)
    if err != nil {
        return labelMap.LabelMapEntries, words, err
    }

    err = json.Unmarshal(data, &labelMap)
    if err != nil {
        return labelMap.LabelMapEntries, words, err
    }

    words = make([]string, len(labelMap.LabelMapEntries))
    i := 0
    for key := range labelMap.LabelMapEntries {
        words[i] = key
        i++
    }

    return labelMap.LabelMapEntries, words, nil
}