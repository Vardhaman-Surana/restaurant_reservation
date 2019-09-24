package models

type Logger struct{
	ServiceName string
	TimeStamp int
	RequestUrl string
	RequestBody string
	ResponseCode int
	ResponseBody string
	Information string
}
