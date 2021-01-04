package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/opreader/zoro/spinner"
)

// http://www.nmc.cn/rest/province/AGD
const Guangzhou = 59287

func main() {
	s := spinner.New(spinner.CharSets[4], 800*time.Millisecond)
	s.Start()
	//defer s.Stop()

	msg := weather(Guangzhou)
	sendMsg(msg...)

	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			msg := weather(Guangzhou)
			sendMsg(msg...)
		default:
		}
	}
}

func sendMsg(msg ...string) {
	log.Println("start..")
	pids, err := robotgo.FindIds("WeChat")
	if err != nil {
		panic(err)
	}
	log.Printf("found WeChat! pids: %v", pids)
	_ = robotgo.ActivePID(pids[0])
	time.Sleep(2000 * time.Millisecond)

	for _, s := range msg {
		robotgo.TypeStr(s)
		time.Sleep(time.Second)
		robotgo.KeyTap("enter")
	}
}

func weather(stationId int) []string {
	resp, err := http.Get(fmt.Sprintf("http://www.nmc.cn/rest/weather?stationid=%d", stationId))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}

	data := res.Data
	city := data.Real.Station.City
	temperature := data.Real.Weather.Temperature
	feelst := data.Real.Weather.Feelst
	humidity := data.Real.Weather.Humidity
	rain := data.Real.Weather.Rain
	airPressure := data.Real.Weather.Airpressure
	direct := data.Real.Wind.Direct
	power := data.Real.Wind.Power
	publishTime, err := time.ParseInLocation("2006-01-02 15:04", data.Real.PublishTime, time.Local)
	if err != nil {
		log.Fatal(publishTime)
	}
	today := fmt.Sprintf(`%s现在温度%0.1f℃，相对湿度%0.0f%%，体感温度%0.1f℃，空气质量%s，%s(%s)，降水量%0.0fmm，气压%0.0fhPa`,
		city, temperature, humidity, feelst, data.Air.Text, power, direct, rain, airPressure) + "\n" +
		fmt.Sprintf(`中央气象台(%s发布)`, publishTime.Format("15:04"))

	var tomorrow string
	var maxTemp, minTemp float64
	index := -1
	for i, t := range data.Tempchart {
		if index >= 0 && i == index+1 { //tomorrow
			maxTemp, minTemp = t.MaxTemp, t.MinTemp
			break
		}
		if time.Now().Format("2006/01/02") == t.Time { //today
			index = i
		}
	}
	index = -1
	for i, d := range data.Predict.Detail {
		if index >= 0 && i == index+1 {
			tomorrow = fmt.Sprintf(`预计明天白天%s，%s(%s)，最高温度%0.1f℃，最低温度%0.1f℃，夜晚%s，%s(%s)`,
				d.Day.Weather.Info, d.Day.Wind.Power, d.Day.Wind.Direct, maxTemp, minTemp,
				d.Night.Weather.Info, d.Night.Wind.Power, d.Night.Wind.Direct)
			break
		}
		if time.Now().Format("2006-01-02") == d.Date {
			index = i
		}
	}
	return []string{today, tomorrow}
}

type Response struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		Real struct {
			Station struct {
				Code     string `json:"code"`
				Province string `json:"province"`
				City     string `json:"city"`
				URL      string `json:"url"`
			} `json:"station"`
			PublishTime string `json:"publish_time"`
			Weather     struct {
				Temperature     float64 `json:"temperature"`
				TemperatureDiff float64 `json:"temperatureDiff"`
				Airpressure     float64 `json:"airpressure"`
				Humidity        float64 `json:"humidity"`
				Rain            float64 `json:"rain"`
				Rcomfort        int     `json:"rcomfort"`
				Icomfort        int     `json:"icomfort"`
				Info            string  `json:"info"`
				Img             string  `json:"img"`
				Feelst          float64 `json:"feelst"`
			} `json:"weather"`
			Wind struct {
				Direct string `json:"dect"`
				Power  string `json:"power"`
				Speed  string `json:"speed"`
			} `json:"wind"`
			Warn struct {
				Alert        string `json:"alert"`
				Pic          string `json:"pic"`
				Province     string `json:"province"`
				City         string `json:"city"`
				URL          string `json:"url"`
				Issuecontent string `json:"issuecontent"`
				Fmeans       string `json:"fmeans"`
				Signaltype   string `json:"signaltype"`
				Signallevel  string `json:"signallevel"`
				Pic2         string `json:"pic2"`
			} `json:"warn"`
		} `json:"real"`
		Predict struct {
			Station struct {
				Code     string `json:"code"`
				Province string `json:"province"`
				City     string `json:"city"`
				URL      string `json:"url"`
			} `json:"station"`
			PublishTime string `json:"publish_time"`
			Detail      []struct {
				Date string `json:"date"`
				Pt   string `json:"pt"`
				Day  struct {
					Weather struct {
						Info        string `json:"info"`
						Img         string `json:"img"`
						Temperature string `json:"temperature"`
					} `json:"weather"`
					Wind struct {
						Direct string `json:"direct"`
						Power  string `json:"power"`
					} `json:"wind"`
				} `json:"day"`
				Night struct {
					Weather struct {
						Info        string `json:"info"`
						Img         string `json:"img"`
						Temperature string `json:"temperature"`
					} `json:"weather"`
					Wind struct {
						Direct string `json:"direct"`
						Power  string `json:"power"`
					} `json:"wind"`
				} `json:"night"`
			} `json:"detail"`
		} `json:"predict"`
		Air struct {
			Forecasttime string `json:"forecasttime"`
			Aqi          int    `json:"aqi"`
			Aq           int    `json:"aq"`
			Text         string `json:"text"`
			AqiCode      string `json:"aqiCode"`
		} `json:"air"`
		Tempchart []struct {
			Time      string  `json:"time"`
			MaxTemp   float64 `json:"max_temp"`
			MinTemp   float64 `json:"min_temp"`
			DayImg    string  `json:"day_img"`
			DayText   string  `json:"day_text"`
			NightImg  string  `json:"night_img"`
			NightText string  `json:"night_text"`
		} `json:"tempchart"`
		Passedchart []struct {
			Rain1H        float64 `json:"rain1h"`
			Rain24H       float64 `json:"rain24h"`
			Rain12H       float64 `json:"rain12h"`
			Rain6H        float64 `json:"rain6h"`
			Temperature   float64 `json:"temperature"`
			TempDiff      string  `json:"tempDiff"`
			Humidity      float64 `json:"humidity"`
			Pressure      float64 `json:"pressure"`
			WindDirection float64 `json:"windDirection"`
			WindSpeed     float64 `json:"windSpeed"`
			Time          string  `json:"time"`
		} `json:"passedchart"`
		Climate struct {
			Time  string `json:"time"`
			Month []struct {
				Month         int     `json:"month"`
				MaxTemp       float64 `json:"maxTemp"`
				MinTemp       float64 `json:"minTemp"`
				Precipitation float64 `json:"precipitation"`
			} `json:"month"`
		} `json:"climate"`
		Radar struct {
			Title string `json:"title"`
			Image string `json:"image"`
			URL   string `json:"url"`
		} `json:"radar"`
	} `json:"data"`
}
