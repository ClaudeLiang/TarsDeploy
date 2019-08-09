package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Logger interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type L struct {}
func (l L)Infof(format string, v ...interface{}) {fmt.Printf(format + "\n", v...)}
func (l L)Debugf(format string, v ...interface{}) {fmt.Printf(format + "\n", v...)}
func (l L)Errorf(format string, v ...interface{}) {fmt.Printf(format + "\n", v...)}

func HttpRequest(method string, url string, body *[]byte, headers *map[string]string) (string, error) {
	client := &http.Client{}
	request, _ := http.NewRequest(method, url, bytes.NewReader(*body))
	request.Close = true
	for k, v := range *headers {
		request.Header.Set(k, v)
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", errors.New("HTTP statusCode:" + strconv.Itoa(response.StatusCode))
	}
	rsp, _ := ioutil.ReadAll(response.Body)
	rspBody := string(rsp)
	return rspBody, nil
}
