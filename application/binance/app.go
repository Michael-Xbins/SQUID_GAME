package binance

import (
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"math"
	"net/http"
	"strconv"
	"time"
)

const Latency = 3000

func StartPullBinance() {
	utils.SafeGo(func() {
		doWork()
	})
}

func doWork() {
	fetchTime, err := redis.GetFetchTime()
	if err != nil {
		log.Panic(err)
	}
	st, e := GetServerTime()
	if e != nil {
		log.Panic(e)
	}
	var currentTime int64
	serverDate := time.UnixMilli(st)
	nearDate := GetNearEffectTime(serverDate)
	fetchTime = nearDate.UnixMilli()
	currentTime = fetchTime
	for i := 1; i < math.MaxInt64; i++ {
		if i%100 == 0 {
			if st, e = GetServerTime(); e != nil {
				log.Error(e)
				continue
			} else {
				currentTime = st
			}
		}
		start := time.Now()
		if currentTime >= fetchTime {
			nextTimestamp := fetchTime             // nextTimestamp要<=serverTime
			num, e := GetNumbers(fetchTime + 3000) // 临时 测试binance快3秒
			if e != nil {
				log.Error(e)
				time.Sleep(time.Second * 1)
				continue
			}
			if e = redis.Push(fetchTime, num, redis.SquidPrefix); e != nil {
				log.Error(e)
				time.Sleep(time.Second * 1)
				continue
			}

			vol, e := GetVolNum(fetchTime-1000, nextTimestamp)
			if e != nil {
				log.Error(e)
				time.Sleep(time.Second * 1)
				continue
			}
			if e = redis.Push(fetchTime, vol, redis.Prefix); e != nil {
				log.Error(e)
				time.Sleep(time.Second * 1)
				continue
			}

			//times, BinanceData, _ := redis.BPop(redis.SquidPrefix)
			//tt := time.UnixMilli(times)
			//if tt.Year() != 1970 {
			//	formattedTime := tt.Format("2006-01-02 15:04:05")
			//	tt.Year()
			//	fmt.Printf("fetchTime: %v, binanceData: %v\n", formattedTime, BinanceData)
			//}

			fetchTime += 1000
		} else {
			time.Sleep(30 * time.Millisecond)
		}
		currentTime += time.Now().Sub(start).Milliseconds()
	}
}

func GetServerTime() (int64, error) {
	for {
		var url string
		if viper.GetString("common.env") == utils.Test {
			//url = "http://43.156.159.52:9600/proxy?url_proxy=https://www.binance.com/dapi/v1/time"
			url = "http://43.156.159.52:9600/GetServerTime"
		} else {
			url = "https://www.binance.com/dapi/v1/time"
		}
		client := &http.Client{
			Timeout: time.Second * 3,
		}
		resp, err := client.Get(url)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return 0, err
		}
		stime := result["serverTime"]
		if stime == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		//serverTime := time.UnixMilli(int64(stime.(float64)))
		//currentTime := time.Now().UTC()
		//duration := currentTime.Sub(serverTime)
		//// 输出当前UTC时间和服务器时间
		//fmt.Println("Current UTC Time:", currentTime.Format("2006-01-02 15:04:05"))
		//fmt.Println("Server Time:", serverTime.Format("2006-01-02 15:04:05"))
		//// 输出时间差
		//fmt.Println("Difference:", duration)
		//// 如果需要输出时间差的绝对值（无论正负）
		//fmt.Println("Absolute Difference:", duration.Round(time.Millisecond))
		//fmt.Println()

		return int64(stime.(float64)) - Latency, nil //测试：延迟3秒
	}
}

func GetNearEffectTime(effectTime time.Time) time.Time {
	year, month, day := effectTime.Date()
	hour := effectTime.Hour()
	minute := effectTime.Minute()
	second := effectTime.Second()
	//if second < 30 {
	//	second = 30
	//} else {
	//	second = 0
	//	minute++
	//}
	return time.Date(year, month, day, hour, minute, second, 0, time.Local)
}

// http://10.226.60.7:9600/proxy?url_proxy=https://www.binance.com/api/v3/uiKlines?endTime=1728983568000%26limit=1%26symbol=BTCUSDT%26interval=1s
// 成交价
func GetNumbers(timestamp int64) (float64, error) {
	var url string
	if viper.GetString("common.env") == utils.Test {
		//url = "http://43.156.159.52:9600/proxy?url_proxy=https://www.binance.com/api/v3/klines?startTime=" + fmt.Sprintf("%d", timestamp) + "%26endTime=" + fmt.Sprintf("%d", timestamp) + "%26limit=1%26symbol=BTCUSDT%26interval=1s"
		url = fmt.Sprintf("http://43.156.159.52:9600/GetNumbers?startTime=%d&endTime=%d", timestamp, timestamp)
	} else {
		url = fmt.Sprintf("https://www.binance.com/api/v3/klines?startTime=%d&endTime=%d&limit=1&symbol=BTCUSDT&interval=1s", timestamp, timestamp)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result [][]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	//if err != nil {
	//	return 0, err
	//}
	//data, _ := json.Marshal(result)
	//fmt.Println(timestamp, string(data))
	if len(result) == 0 || len(result[0]) < 5 {
		return 0, errors.New("result error")
	}
	num, err := strconv.ParseFloat(result[0][4].(string), 64)
	if err != nil {
		return 0, err
	}
	return num, err
}

// 成交量
func GetVolNum(startTime int64, endTime int64) (float64, error) {
	var url string
	if viper.GetString("common.env") == utils.Test {
		//url = "http://43.156.159.52:9600/proxy?url_proxy=https://api.binance.com/api/v3/aggTrades?symbol=BTCUSDT%26limit=1000%26startTime=" + fmt.Sprintf("%d", startTime) + "%26endTime=" + fmt.Sprintf("%d", endTime)
		url = fmt.Sprintf("http://43.156.159.52:9600/GetVolNum?symbol=BTCUSDT&limit=1000&startTime=%d&endTime=%d", startTime, endTime)
	} else {
		url = fmt.Sprintf("https://api.binance.com/api/v3/aggTrades?symbol=BTCUSDT&limit=1000&startTime=%d&endTime=%d", startTime, endTime)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer resp.Body.Close()
	type Value struct {
		Q string `json:"q"`
	}
	var result []Value
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}
	var sum float64
	if len(result) > 0 {
		for _, value := range result {
			v, _ := strconv.ParseFloat(value.Q, 64)
			sum += v
		}
	}
	return sum, nil
}
