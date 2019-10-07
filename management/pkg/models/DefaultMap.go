package models

import (
	"encoding/json"
	"net/http"
)

type DefaultMap map[string]interface{}


func (d *DefaultMap)ConvertToByteArray()[]byte{
	jsonData,_:=json.Marshal(d)
	return jsonData
}

func WriteToResponse(w http.ResponseWriter,statusCode int,model *DefaultMap){
	w.WriteHeader(statusCode)
	jsonData:=model.ConvertToByteArray()
	w.Write(jsonData)
}