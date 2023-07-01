package smartapigo

import (
	"fmt"
	"net/http"
	"time"
)

type HistoryParams struct {
	Exchange    string `json:"exchange"`
	SymbolToken string `json:"symboltoken"`
	Interval    string `json:"interval"`
	FromDate    string `json:"fromdate"`
	ToDate      string `json:"todate"`
}

func (h *HistoryParams) GetParams() map[string]interface{} {
	params := make(map[string]interface{})
	params["exchange"] = h.Exchange
	params["symboltoken"] = h.SymbolToken
	params["interval"] = h.Interval
	params["fromdate"] = h.FromDate
	params["todate"] = h.ToDate
	return params
}

type HistoryDatum struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type HistoryResponse struct {
	Status    bool           `json:"status"`
	Message   string         `json:"message"`
	Errorcode string         `json:"errorcode"`
	Data      []HistoryDatum `json:"data"`
}

func (h HistoryResponse) String() string {
	retVal := fmt.Sprintf("Status: %t\nMessage: %s\nErrorcode: %s\n", h.Status, h.Message, h.Errorcode)
	for _, datum := range h.Data {
		retVal += fmt.Sprintf("\tTimeStamp: %s\n\tOpen: %f, High: %f, Low: %f, Close: %f, Volume: %f", datum.Timestamp, datum.Open, datum.High, datum.Low, datum.Close, datum.Volume)
	}
	return retVal
}

// GetCandleData gets history of the specified symbol between a defined time-range
func (c *Client) GetCandleData(params HistoryParams) (HistoryResponse, error) {
	var history HistoryResponse

	err := c.doEnvelope(http.MethodGet, URIGetCandleData, params.GetParams(), nil, &history, true)
	return history, err
}
