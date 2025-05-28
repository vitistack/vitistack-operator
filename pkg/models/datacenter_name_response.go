package models

type Datacenter struct {
	Name     string `json:"name"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Provider string `json:"provider"`
}
