package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

func SetupCommonHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 在这里设置你的统一Header
		c.Response().Header().Set("Content-Type", "application/json")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// 调用下一个中间件或最终的处理程序
		return next(c)
	}
}

func Options(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	return c.String(http.StatusNonAuthoritativeInfo, "")
}

func getServerTime(c echo.Context) error {
	url := "https://www.binance.com/dapi/v1/time"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error parsing JSON from Binance")
	}
	return c.JSON(http.StatusOK, result)
}

func getNumbers(c echo.Context) error {
	startTimeStr := c.QueryParam("startTime")
	if startTimeStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "startTime parameter is required")
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid startTime")
	}
	endTimeStr := c.QueryParam("endTime")
	if endTimeStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "endTime parameter is required")
	}
	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid endTime")
	}
	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?startTime=%d&endTime=%d&limit=1&symbol=BTCUSDT&interval=1s", startTime, endTime)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error parsing JSON from Binance")
	}
	return c.JSON(http.StatusOK, result)
}

func getVolNum(c echo.Context) error {
	startTimeStr := c.QueryParam("startTime")
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid startTime")
	}
	endTimeStr := c.QueryParam("endTime")
	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid endTime")
	}
	url := fmt.Sprintf("https://api.binance.com/api/v3/aggTrades?symbol=BTCUSDT&limit=1000&startTime=%d&endTime=%d", startTime, endTime)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error parsing JSON from Binance")
	}
	return c.JSON(http.StatusOK, result)
}

func main() {
	e := echo.New()
	e.Use(SetupCommonHeaders)
	//e.OPTIONS("/proxy/*", model.Options)
	e.GET("/GetServerTime", getServerTime)
	e.GET("/GetNumbers", getNumbers)
	e.GET("/GetVolNum", getVolNum)
	e.Start("0.0.0.0:9600")
}

//func main() {
//	// 后端服务器的URL，这里假设是http://example.com
//	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
//		Scheme: "https",
//		Host:   "alegrialoteria.com:443", // 目标服务器地址
//	})
//	http.Handle("/api_igt/pronosticos/reporte/tris", proxy)
//
//	log.Fatal(http.ListenAndServe(":18084", nil))
//}
