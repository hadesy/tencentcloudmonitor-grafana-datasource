package model

type QueryModel struct {
	Service    string      `json:"service"`
	Region     string      `json:"region"`
	Metric     string      `json:"metric"`
	Period     uint64      `json:"period"`
	Dimensions []Dimension `json:"dimensions"`
	StartTime  string      `json:"startDate"`
	EndTime    string      `json:"endDate"`
	Hide       bool        `json:"hide"`
}

type Dimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
