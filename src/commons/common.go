package commons

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
    "github.com/garyburd/redigo/redis"
    log "github.com/Sirupsen/logrus"
    "net/url"
    "errors"
    "github.com/gin-gonic/gin"
    "../datastructures"
    "strconv"
)


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

func Random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func Pick(args ...interface{}) []interface{} {
    return args
}

func HashImage(file io.Reader) (uint64, error){
    img, _, err := image.Decode(file)
    if err != nil {
        return 0, err
    }

    return imghash.Average(img), nil
}

func GetImageInfo(file io.Reader) (datastructures.ImageInfo, error){
    var imageInfo datastructures.ImageInfo
    imageInfo.Hash = 0
    imageInfo.Width = 0
    imageInfo.Height = 0

    img, _, err := image.Decode(file)
    if err != nil {
        return imageInfo, err
    }

    bounds := img.Bounds()

    imageInfo.Hash = imghash.Average(img)
    imageInfo.Width = int32(bounds.Dx())
    imageInfo.Height = int32(bounds.Dy())

    return imageInfo, nil
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


func GetIPAddress(r *http.Request) string {
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

func GetLabelMap(path string) (map[string]datastructures.LabelMapEntry, []string, error) {
    var words []string
    var labelMap datastructures.LabelMap

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

func GetLabelRefinementsMap(path string) (map[string]datastructures.LabelMapRefinementEntry, error) {
    var labelMapRefinementEntries map[string]datastructures.LabelMapRefinementEntry

    data, err := ioutil.ReadFile(path)
    if err != nil {
        return labelMapRefinementEntries, err
    }

    err = json.Unmarshal(data, &labelMapRefinementEntries)
    if err != nil {
        return labelMapRefinementEntries, err
    }

    return labelMapRefinementEntries, nil
}

func GetSampleExportQueries() []string {
    var queries []string
    queries = append(queries, "dog | cat")
    queries = append(queries, "dog.has = 'mouth' | cat")
    queries = append(queries, "dog.size = 'big' | dog.size = 'small'")

    return queries
}

type RegisteredAppIdentifiersInterface interface {
    Load() error
    IsValid(key string) bool
    Get() (string, bool)
}

type RegisteredAppIdentifiers struct {
    identifiers map[string]string
}

func NewRegisteredAppIdentifiers() *RegisteredAppIdentifiers {
    return &RegisteredAppIdentifiers{} 
}

func (p *RegisteredAppIdentifiers) Load() error {
    p.identifiers = make(map[string]string)
    p.identifiers["edd77e5fb6fc0775a00d2499b59b75d"] = "ImageMonkey Website"
    p.identifiers["adf78e53bd6fc0875a00d2499c59b75"] = "ImageMonkey Browser Extension"
    return nil
}

func (p *RegisteredAppIdentifiers) IsValid(key string) bool {
    _, ok := p.identifiers[key]
    return ok
}

func (p *RegisteredAppIdentifiers) Get(key string) (string, bool) {
    val, ok := p.identifiers[key]
    return val, ok
}


type StatisticsPusherInterface interface {
    PushAppAction(appIdentifier string, actionType string)
    Load() error
}

type StatisticsPusher struct {
    registeredAppIdentifiers *RegisteredAppIdentifiers
    redisPool *redis.Pool
}

func NewStatisticsPusher(redisPool *redis.Pool) *StatisticsPusher {
    return &StatisticsPusher{
        redisPool: redisPool,
        registeredAppIdentifiers: NewRegisteredAppIdentifiers(),
    } 
}

func (p *StatisticsPusher) Load() error {
    return p.registeredAppIdentifiers.Load()
}

func (p *StatisticsPusher) PushAppAction(appIdentifier string, actionType string) {
    var contributionsPerAppRequest datastructures.ContributionsPerAppRequest
    contributionsPerAppRequest.Type = actionType
    val, ok := p.registeredAppIdentifiers.Get(appIdentifier)
    if ok {
        contributionsPerAppRequest.AppIdentifier = val
        serialized, err := json.Marshal(contributionsPerAppRequest)
        if err != nil {
            log.Debug("[Push Contributions per App to Redis] Couldn't create contributions-per-app request: ", err.Error())
            return
        }

        redisConn := p.redisPool.Get()
        defer redisConn.Close()

        _, err = redisConn.Do("RPUSH", "contributions-per-app", serialized)
        if err != nil { //just log error, but not abort (it's just some statistical information)
            log.Debug("[Push Contributions per App to Redis] Couldn't update contributions-per-app: ", err.Error())
            return
        }
    }
}


func IsAlphaNumeric(s string) bool {
    for _, c := range s {
        if (!(c > 47 && c < 58) && // numeric (0-9)
            !(c > 64 && c < 91) && // upper alpha (A-Z)
            !(c > 96 && c < 123)) { // lower alpha (a-z)
            return false
        }
    }
    return true
}

func IsLabelValid(labelsMap map[string]datastructures.LabelMapEntry, label string, sublabels []datastructures.Sublabel) bool {
    if val, ok := labelsMap[label]; ok {
        if len(sublabels) > 0 {
            availableSublabels := val.LabelMapEntries

            for _, value := range sublabels {
                _, ok := availableSublabels[value.Name]
                if !ok {
                    return false
                }
            }
            return true
        }
        return true
    }

    return false
}

func GetLabelIdFromUrlParams(params url.Values) (string, error) {
    var labelId string
    labelId = ""
    if temp, ok := params["label_id"]; ok {
        labelId = temp[0]
    }

    return labelId, nil
}

func GetValidationIdFromUrlParams(params url.Values) string {
    var validationId string
    validationId = ""
    if temp, ok := params["validation_id"]; ok {
        validationId = temp[0]
    }

    return validationId
}

func GetExploreUrlParams(c *gin.Context) (string, bool, error) {
    var query string
    var err error

    params := c.Request.URL.Query()

    annotationsOnly := false
    if temp, ok := params["annotations_only"]; ok {
        if temp[0] == "true" {
            annotationsOnly = true
        }
    }

    if temp, ok := params["query"]; ok {
        if temp[0] == "" {
            return "", annotationsOnly, errors.New("no query specified")
        }


        query, err = url.QueryUnescape(temp[0])
        if err != nil {
            return "", annotationsOnly, errors.New("invalid query")
        }
    } else {
        return "", annotationsOnly, errors.New("no query specified")
    }

    return query, annotationsOnly, nil 
}

func GetParamFromUrlParams(c *gin.Context, name string, defaultIfNotFound string) string {
    params := c.Request.URL.Query()

    param := defaultIfNotFound
    if temp, ok := params[name]; ok {
        param = temp[0]
    }

    return param
}

func GetParamsFromUrlParams(c *gin.Context, name string) []string {
    params := c.Request.URL.Query()

    if temp, ok := params[name]; ok {
        return temp
    }

    return []string{}
}

func GetImageUrlFromImageId(apiBaseUrl string, imageId string, unlocked bool) string {
    imageUrl := apiBaseUrl
    if unlocked {
        imageUrl += "v1/donation/" + imageId
    } else {
        imageUrl += "v1/unverified-donation/" + imageId
    }

    return imageUrl
}

func GetPublicBackups(path string) ([]datastructures.PublicBackup, error){
    var publicBackups []datastructures.PublicBackup

    data, err := ioutil.ReadFile(path)
    if err != nil {
        return publicBackups, err
    }

    err = json.Unmarshal(data, &publicBackups)
    if err != nil {
        return publicBackups, err
    }

    return publicBackups, nil
}

func GetImageRegionsFromUrlParams(c *gin.Context) ([]image.Rectangle, error) {
    regionsOfInterest := GetParamsFromUrlParams(c, "roi")
    imageRects := []image.Rectangle{}
    
    for _,regionOfInterest := range regionsOfInterest {
        regionOfInterestParams := strings.Split(regionOfInterest, ",")

        var err error
        x0 := 0
        y0 := 0
        x1 := 0
        y1 := 0

        if len(regionOfInterestParams) == 4 {
            x0, err = strconv.Atoi(regionOfInterestParams[0])
            if err != nil {
                return imageRects, err
            }
        } 
        if len(regionOfInterestParams) >= 2 {
            y0, err = strconv.Atoi(regionOfInterestParams[1])
            if err != nil {
                return imageRects, err
            }
        }
        if len(regionOfInterestParams) >= 3 {
            x1, err = strconv.Atoi(regionOfInterestParams[2])
            if err != nil {
                return imageRects, err
            }
        }
        if len(regionOfInterestParams) >= 4 {
            y1, err = strconv.Atoi(regionOfInterestParams[3])
            if err != nil {
                return imageRects, err
            }
        }

        imageRects = append(imageRects, image.Rect(x0, y0, x1, y1))
    }

    return imageRects, nil
}

func GetAvailableModels(s string) ([]json.RawMessage, error) {
    var models []json.RawMessage

    _, err := url.ParseRequestURI(s)
    if err == nil { //it's an URL
        resp, err := http.Get(s)
        if err != nil {
            return models, err
        }
        defer resp.Body.Close()

        data, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return models, err
        }

        err = json.Unmarshal(data, &models)
        if err != nil {
            return models, err
        }

    } else {
        data, err := ioutil.ReadFile(s)
        if err != nil {
            return models, err
        }

        err = json.Unmarshal(data, &models)
        if err != nil {
            return models, err
        }
    }

    return models, nil
}