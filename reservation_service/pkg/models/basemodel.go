package models

type BaseModel struct{
	ID int `json:"id"`
	Created int64 `json:"created"`
	Updated int64 `json:"updated"`
	Deleted bool `json:"deleted"`
}

