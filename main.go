package main

import (
	"encoding/json"
	"fetchTest/server"
	_ "fetchTest/server"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/PuerkitoBio/goquery"
	"github.com/mitchellh/mapstructure"
	_ "github.com/mitchellh/mapstructure"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	fetchTime      = 6
	waitTime       = 0
	numOfThread    = 5
	stuckTimeK     = 5
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.50"
	requestTest    = "https://leetcode.com/contest/api/ranking/weekly-contest-318/?pagination=1&region=global"
	requestRank    = "https://leetcode.com/contest/api/ranking/"
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

func main() {
	server.GinRun()
	//ChannelStart("weekly-contest-333")
	//Predict()
}

type Contestant struct {
	Contest_id           int
	Username             string
	User_slug            string
	Rank                 int
	Finish_time          int64
	Data_region          string
	Attend               bool
	AttendedContestCount int
	Country_code         interface{}
	Country_name         interface{}
	Score                int
	Rating               float64
	PredictedRating      float64
	//Global_ranking int
}

var contestant = make(map[int]Contestant)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

var start time.Time

func FactorF(k int) float64 {
	sum := 1.0
	x := 1.0
	for i := 0; i <= k; i++ {
		sum += x
		x *= (5.0 / 7.0)
	}
	return (1.0) / sum
}

var pageMap [1200]map[string]interface{}
var ratingMap [30001]map[string]interface{}
var cnt int = 0

func Predict() {
	for k, v := range contestant {
		eRank := 0.5
		for _, v1 := range contestant {
			eRank += 1.0 / (1.0 + math.Pow(10.0, (v.Rating-v1.Rating)/400))
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
		v.PredictedRating = left
		contestant[k] = v
	}
}
func ChannelStart(contestName string) {
	start = time.Now()
	client := http.Client{}
	req, _ := http.NewRequest("GET", requestRank+contestName, nil)
	req.Header.Set("Cookie", myCookie)
	req.Header.Set("Accept", "*/*")
	resp, _ := client.Do(req)
	docDetail, _ := goquery.NewDocumentFromReader(resp.Body)
	firstPageMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(docDetail.Text()), &firstPageMap)
	contestantNum := int(firstPageMap["user_num"].(float64))
	pageNum := (contestantNum-1)/25 + 1
	pagesPerThread := (pageNum / numOfThread)
	ch := make(chan bool)
	for i := 1; i <= pageNum; i += pagesPerThread {
		go fetchRank("weekly-contest-333", false, ch, i, min(pageNum, i+pagesPerThread-1))
	}
	for i := 1; i <= pageNum; i += pagesPerThread {
		<-ch
	}
	elapsed := time.Since(start)
	fmt.Printf("ChannelStart Time %s \n", elapsed)
}

func getrand() float64 {
	return 0.1 + rand.Float64()/5
}
func sleeprand() float64 {
	return 0.05 + rand.Float64()*0.15
}
func fetchRank(contestName string, isBiweek bool, ch chan bool, startPage int, endPage int) {

	client := http.Client{}
	var tempContestant Contestant
	for i := startPage; i <= endPage; i++ {
		/*
			For each page,request the rank
		*/
		reqUrl := requestRank + contestName + urlPrefix + strconv.Itoa(i) + urlSuffix
		flg1 := false
		var unerr2 error
		/*

		 */
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
			unerr2 = json.Unmarshal([]byte(docDetail.Text()), &pageMap[i])
			if unerr2 != nil {
				randtime := getrand() * stuckTimeK
				time.Sleep(time.Duration(randtime) * time.Second)
				fmt.Printf("Page %d Requset Error , bout to sleep for %v second\n\n\n", i, randtime)
			}
			//sleep process after each request
			time.Sleep(time.Duration(sleeprand()) * time.Second)
		}
		println(i)
		contestantInfo := pageMap[i]["total_rank"]
		for _, v := range contestantInfo.([]interface{}) {
			r := v.(map[string]interface{})
			_ = mapstructure.Decode(r, &tempContestant)
			if tempContestant.Score > 0 {
				tempContestant.Attend = true
			} else {
				tempContestant.Attend = false
			}
			if !tempContestant.Attend {
				continue
			}
			if !isBiweek {
				var reqBody string
				var req *http.Request
				flg := false
				var unerr3 error
				for !flg || unerr3 != nil {
					flg = true
					if tempContestant.Data_region == "US" {
						reqBody = usRatingPrefix + tempContestant.Username + usRatingSuffix
						req, _ = http.NewRequest("POST", usReqUrl, strings.NewReader(reqBody))
					} else {
						reqBody = cnRatingPrefix + tempContestant.Username + cnRatingSuffix
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
					if unerr3 := json.Unmarshal([]byte(docDetail.Text()), &ratingMap[tempContestant.Rank]); unerr3 != nil {
						randtime := getrand() * stuckTimeK
						println("User Rating Request Stuck at Rank ", tempContestant.Rank, "\tbout to sleep for", randtime, "second")
						time.Sleep(time.Duration(randtime) * time.Second)
						println("Over Stuck at Rank ", tempContestant.Rank, "for", randtime, "second")
					}
				}
				cnt++
				nowTime := int(time.Since(start).Seconds())
				vv := (cnt) / nowTime
				fmt.Printf("Success Request at Rank %5d ,total requests are now %5d ,time elapse: %4d speed:%2d\n", tempContestant.Rank,
					cnt, nowTime, vv)
				//sleep process after each request
				time.Sleep(time.Duration(sleeprand()) * time.Second)
				if ratingMap[tempContestant.Rank]["data"] == nil {
					tempContestant.Rating = 1500.00
				} else {
					userContestRankingMap := ratingMap[tempContestant.Rank]["data"].(map[string]interface{})
					if userContestRankingMap["userContestRanking"] != nil {
						lastMap := userContestRankingMap["userContestRanking"].(map[string]interface{})
						tempContestant.Rating = (float64)(lastMap["rating"].(float64))
					} else {
						tempContestant.Rating = 1500.00
					}
				}
			} else {

			}
			contestant[tempContestant.Rank] = tempContestant
		}
	}
	testOrNot := false
	if testOrNot {
		for k, v := range contestant {
			println(k)
			fmt.Printf("%v\n", v)
		}
	}
	if ch != nil {
		ch <- true
	}
}
