package domain

type MapResponse struct {
	Status int   `json:"status"`
	Data   []Map `json:"data"`
}
