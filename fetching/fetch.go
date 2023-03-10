package fetching

import (
	"encoding/json"
	"fetchTest/common"
	"fetchTest/model"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fetchTime      = 6
	waitTime       = 0
	numOfThread    = 5
	stuckTimeK     = 5
	testPage       = 1
	testPageNum    = 5
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.50"
	requestTest    = "https://leetcode.com/contest/api/ranking/weekly-contest-318/?pagination=1&region=global"
	requestRank    = "https://leetcode.cn/contest/api/ranking/"
	urlPrefix      = "/?pagination="
	urlSuffix      = "&region=global"
	myCookie       = "csrftoken=pJ3Vi3l17nuhHN0x0icfdVtmCvkg2nDG2ubcDQUuRO0PmMCQKrmZCA8PK0oJXN0E"
	usReqUrl       = "https://leetcode.com/graphql/"
	cnReqUrl       = "https://leetcode.cn/graphql/noj-go/"
	usRatingPrefix = "{\"query\":\"\\n    query userContestRankingInfo($username: String!) {\\n  userContestRanking(username: $username) {\\n rating\\n   attendedContestsCount\\n }\\n \\n}\\n\",\"variables\":{\"username\":\""
	cnRatingPrefix = "{\"query\":\"\\n    query userContestRankingInfo($userSlug: String!) {\\n  userContestRanking(userSlug: $userSlug) {\\n     rating\\n    attendedContestsCount\\n  }\\n  \\n}\\n    \",\"variables\":{\"userSlug\":\""
	usRatingSuffix = "\"}}"
	cnRatingSuffix = "\"}}"

	usSubmitPrefix = "{\"query\":\"\\n    query recentSubmissionList($username: String!) {\\n  recentSubmissionList(username: $username) {\\n   status\\n  titleSlug\\n    timestamp\\n  }\\n}\\n    \",\"variables\":{\"username\":\""
	cnSubmitPrefix = "{\"operationName\":\"RecentSubmissions\",\"variables\":{\"userSlug\":\""
	usSubmitSuffix = "\"}}"
	cnSubmitSuffix = "\"},\"query\":\"query RecentSubmissions($userSlug: String!) { recentSubmissions(userSlug: $userSlug) { status submitTime id question { titleSlug } } }\"}"
)

var contestant = make(map[int]model.Contestant)
var lock sync.Mutex

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

var start time.Time

func FactorF(k int) float64 {
	if k > 19 {
		return 0.223
	}
	sum := 1.0
	x := 1.0
	for i := 0; i <= k; i++ {
		sum += x
		x *= (5.0 / 7.0)
	}
	return (1.0) / sum
}

var pageMap [1200]map[string]interface{}
var cnt int = 0

func ChannelStart(contestName string, isPreparation bool) {
	contestant = make(map[int]model.Contestant)
	start = time.Now()
	contestantNum := GetContestantNum(contestName)
	pageNum := (contestantNum-1)/25 + 1
	if testPage == 1 {
		pageNum = testPageNum
	}
	pagesPerThread := (pageNum / numOfThread)
	ch := make(chan bool)
	for i := 1; i <= pageNum; i += pagesPerThread {
		go fetchRank(contestName, false, ch, i, min(pageNum, i+pagesPerThread-1))
	}
	for i := 1; i <= pageNum; i += pagesPerThread {
		<-ch
	}
	elapsed := time.Since(start)
	fmt.Printf("ChannelStart Time %s \n", elapsed)
	if isPreparation {
		InsertIntoRedis()
		return
	}
	Predict()
	InsertIntoRedis()
	//TestRedis()
	InsertIntoDB()
	contestant = make(map[int]model.Contestant)
}
func Predict() {
	for k, v := range contestant {
		eRank := 0.5
		for _, v1 := range contestant {
			eRank += 1.0 / (1.0 + math.Pow(10.0, (v.Rating-v1.Rating)/400.0))
		}
		avgRank := math.Sqrt(eRank * float64(v.Rank))
		left := 0.0
		right := 4500.0
		for right-left > 0.1 {
			mid := (left + right) / 2
			newRank := 0.5
			for _, v1 := range contestant {
				newRank += 1.0 / (1.0 + math.Pow(10.0, (mid-v1.Rating)/400))
			}
			if newRank > avgRank {
				left = mid
			} else {
				right = mid
			}
		}
		v.PredictedRating = (left - v.Rating) * FactorF(v.AttendedContestCount)
		contestant[k] = v
	}
}
func GetContestantNum(contestName string) int {
	client := http.Client{}
	req, _ := http.NewRequest("GET", requestRank+contestName, nil)
	req.Header.Set("Cookie", myCookie)
	req.Header.Set("Accept", "*/*")
	resp, _ := client.Do(req)
	docDetail, _ := goquery.NewDocumentFromReader(resp.Body)
	firstPageMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(docDetail.Text()), &firstPageMap)
	return int(firstPageMap["user_num"].(float64))
}
func isContestantExisted(db *gorm.DB, con model.Contestant) bool {
	var c model.Contestant
	res := db.Where("contestname=?", con.Contestname)
	res = res.Where("username=?", con.Username)
	res.First(&c)
	return c.ID != 0
}
func ParesRedisKey(s string) (float64, int) {
	if s == "" {
		return 0.0, 0
	}
	pos := strings.IndexByte(s, '_')
	res1, _ := strconv.ParseFloat(s[:pos], 8)
	res2, _ := strconv.Atoi(s[(pos + 1):])
	return res1, res2
}

func InsertIntoRedis() {
	redisDB := common.GetRedisDB()
	for _, v := range contestant {
		instr := v.Username + v.Contestname
		val := fmt.Sprintf("%f_%d", v.PredictedRating, v.AttendedContestCount)
		fmt.Printf("%v %v\n", instr, val)
		_ = redisDB.Set(instr, val, 24*time.Hour).Err()
		instr2 := v.Username
		val2 := fmt.Sprintf("%f_%d", v.Rating, v.AttendedContestCount)
		_ = redisDB.Set(instr2, val2, 24*time.Hour).Err()
	}
}

func TestRedis() {
	redisDB := common.GetRedisDB()
	for _, v := range contestant {
		newstr := v.Username + v.Contestname
		val, _ := redisDB.Get(newstr).Result()
		//k, _ := strconv.Atoi(val)
		res1, res2 := ParesRedisKey(val)
		fmt.Printf("%v  %v  %v\n", newstr, res1, res2)
	}
}

func InsertIntoDB() {
	db := common.GetDB()
	var contestName string
	for _, v := range contestant {
		contestName = v.Contestname
		break
	}
	a := model.Contest{
		StartTime: time.Now().Unix(),
		TitleSlug: contestName,
	}
	db.Create(&a)
	for _, v := range contestant {
		if !isContestantExisted(db, v) {
			db.Create(&v)
		} else {
			db.Where("contestname=?", contestName).Where("username=?", v.Username).Updates(&v)
		}
	}
}
func getrand() float64 {
	return 0.1 + rand.Float64()/5
}
func sleeprand() float64 {
	return 0.05 + rand.Float64()*0.15
}
func GetLatestRatingFromRedis(username string) (float64, int) {
	redisDB := common.GetRedisDB()
	val, _ := redisDB.Get(username).Result()
	return ParesRedisKey(val)
}
func GetLatestRatingFromMysql(username string) (float64, int) {
	db := common.GetDB()
	var con model.Contestant
	yesterday := time.Now().AddDate(0, 0, -1)
	db.Where("username=?", username).Where("updated_at>?", yesterday).First(&con)
	if con.ID != 0 {
		return con.PredictedRating, con.AttendedContestCount
	}
	return 0.0, 0
}
func GetRating(username string, dataregion string) (float64, int) {
	if res1, res2 := GetLatestRatingFromRedis(username); res1 != 0 {
		return res1, res2
	}
	if res1, res2 := GetLatestRatingFromMysql(username); res1 != 0 {
		return res1, res2
	}
	client := http.Client{}
	var reqBody string
	var req *http.Request
	flg := false
	tempMap := make(map[string]interface{})
	for !flg {
		flg = true
		if dataregion == "US" {
			reqBody = usRatingPrefix + username + usRatingSuffix
			req, _ = http.NewRequest("POST", usReqUrl, strings.NewReader(reqBody))
		} else {
			reqBody = cnRatingPrefix + username + cnRatingSuffix
			req, _ = http.NewRequest("POST", cnReqUrl, strings.NewReader(reqBody))
		}
		req.Header.Set("user-agent", UserAgent)
		req.Header.Set("Cookie", myCookie)
		req.Header.Set("Content-Type", "application/json")
		now := time.Since(start).Seconds()
		/*
			The waiting process while high concurrent
		*/
		if int(now)%(fetchTime+waitTime) >= fetchTime {
			pre := int(now) / (fetchTime + waitTime)
			var preT float64 = float64(pre * (fetchTime + waitTime))
			time.Sleep(time.Duration(fetchTime+waitTime-(now-preT)) * time.Second)
		}
		resp, _ := client.Do(req)
		docDetail, _ := goquery.NewDocumentFromReader(resp.Body)
		if unerr3 := json.Unmarshal([]byte(docDetail.Text()), &tempMap); unerr3 != nil || tempMap["data"] == nil {
			flg = false
			randtime := getrand() * stuckTimeK
			println("User Rating Request Stuck at cnt", cnt, "\tbout to sleep for", randtime, "second")
			time.Sleep(time.Duration(randtime) * time.Second)
			println("Over Stuck at cnt ", cnt, "for", randtime, "second")
		}
	}
	cnt++
	nowTime := int(time.Since(start).Seconds())
	if nowTime == 0 {
		nowTime++
	}
	vv := (cnt) / nowTime
	fmt.Printf("Success Request ,total requests are now %5d ,time elapse: %4d speed:%2d\n",
		cnt, nowTime, vv)
	//sleep process after each request
	time.Sleep(time.Duration(sleeprand()) * time.Second)
	userContestRankingMap := tempMap["data"].(map[string]interface{})
	if userContestRankingMap["userContestRanking"] != nil {
		lastMap := userContestRankingMap["userContestRanking"].(map[string]interface{})
		return lastMap["rating"].(float64), (int)(lastMap["attendedContestsCount"].(float64))
	}
	client.CloseIdleConnections()
	return 1500.002, 0
}
func fetchPage(contestName string, page int) {
	client := http.Client{}
	/*
		For each page,request the rank
	*/
	reqUrl := requestRank + contestName + urlPrefix + strconv.Itoa(page) + urlSuffix
	flg1 := false
	var unerr2 error
	for !flg1 || unerr2 != nil {
		flg1 = true
		req, _ := http.NewRequest("GET", reqUrl, nil)
		req.Header.Set("Cookie", myCookie)
		req.Header.Set("Accept", "*/*")
		now := time.Since(start).Seconds()
		/*
			The waiting process while high concurrent
		*/
		if int(now)%(fetchTime+waitTime) >= fetchTime {
			pre := int(now) / (fetchTime + waitTime)
			var preT = float64(pre * (fetchTime + waitTime))
			time.Sleep(time.Duration(fetchTime+waitTime-(now-preT)) * time.Second)
		}
		resp, _ := client.Do(req)
		docDetail, _ := goquery.NewDocumentFromReader(resp.Body)
		unerr2 = json.Unmarshal([]byte(docDetail.Text()), &pageMap[page])
		if unerr2 != nil {
			randtime := getrand() * stuckTimeK
			time.Sleep(time.Duration(randtime) * time.Second)
			fmt.Printf("Page %d Requset Error , bout to sleep for %v second\n\n\n", page, randtime)
		}
		//sleep process after each request
		time.Sleep(time.Duration(sleeprand()) * time.Second)
	}
}
func fetchRank(contestName string, isBiweek bool, ch chan bool, startPage int, endPage int) {
	var tempContestant model.Contestant
	for i := startPage; i <= endPage; i++ {
		fetchPage(contestName, i)
		contestantInfo := pageMap[i]["total_rank"]
		for _, v := range contestantInfo.([]interface{}) {
			r := v.(map[string]interface{})
			_ = mapstructure.Decode(r, &tempContestant)
			tempContestant.Contestname = contestName
			if tempContestant.Score > 0 {
				tempContestant.Attend = true
			} else {
				//tempContestant.Attend = false
				continue
			}
			tempContestant.Rating, tempContestant.AttendedContestCount = GetRating(tempContestant.Username, tempContestant.Data_region)
			lock.Lock()
			contestant[tempContestant.Rank] = tempContestant
			lock.Unlock()
		}
	}
	if ch != nil {
		ch <- true
	}
}
