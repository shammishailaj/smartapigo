package smartapigo

import (
	"fmt"
	"net/http"
	"time"
)

type HistoryParams struct {
	Exchange    Exchange     `json:"exchange"`
	SymbolToken string       `json:"symboltoken"`
	Interval    TimeInterval `json:"interval"`
	FromDate    time.Time    `json:"fromdate"`
	ToDate      time.Time    `json:"todate"`
}

func (h *HistoryParams) GetParams() map[string]interface{} {
	params := make(map[string]interface{})
	params["exchange"] = h.Exchange
	params["symboltoken"] = h.SymbolToken
	params["interval"] = h.Interval
	params["fromdate"] = h.FromDate.Format(TimeFormatLayout)
	params["todate"] = h.ToDate.Format(TimeFormatLayout)
	return params
}

func (h *HistoryParams) ValidDates() bool {
	if h.FromDate.Unix() < h.ToDate.Unix() {
		return true
	}
	return false
}

func (h *HistoryParams) IntervalDays() int64 {
	var days int64

	if h.ValidDates() {
		days = secondsToDays(h.ToDate.Unix() - h.FromDate.Unix())
	} else {
		days = secondsToDays(h.FromDate.Unix() - h.ToDate.Unix())
	}

	return days
}

func (h *HistoryParams) IsValidInterval() bool {
	days := h.IntervalDays()

	switch h.Interval {
	case ONE_MINUTE:
		if days <= 30 {
			return true
		}
	case THREE_MINUTE:
		if days <= 90 {
			return true
		}
	case FIVE_MINUTE:
		if days <= 90 {
			return true
		}
	case TEN_MINUTE:
		if days <= 90 {
			return true
		}
	case FIFTEEN_MINUTE:
		if days <= 180 {
			return true
		}
	case THIRTY_MINUTE:
		if days <= 180 {
			return true
		}
	case ONE_HOUR:
		if days <= 365 {
			return true
		}
	case ONE_DAY:
		if days <= 2000 {
			return true
		}
	default:
		return false
	}
	return false
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

type ABResponse struct {
	Status    bool    `json:"status"`
	Message   string  `json:"message"`
	Errorcode string  `json:"errorcode"`
	Data      [][]any `json:"data"`
}

func (a ABResponse) String() string {
	retVal := fmt.Sprintf("Status: %t\nMessage: %s\nErrorcode: %s\n", a.Status, a.Message, a.Errorcode)
	for _, datum := range a.Data {
		retVal += fmt.Sprintf("\tTimeStamp: %s\n\tOpen: %f, High: %f, Low: %f, Close: %f, Volume: %f\n", datum[0], datum[1], datum[2], datum[3], datum[4], datum[5])
	}
	return retVal
}

func (a ABResponse) Parse() HistoryResponse {
	var retVal HistoryResponse
	retVal.Status = a.Status
	retVal.Message = a.Message
	retVal.Errorcode = a.Errorcode
	data := make([]HistoryDatum, len(a.Data))
	for k, datum := range a.Data {
		ts, tsErr := time.Parse(TimeFormatLayout, datum[0].(string))
		if tsErr != nil {
			ts = time.Now()
		}
		data[k].Timestamp = ts
		data[k].Open = datum[1].(float64)
		data[k].High = datum[2].(float64)
		data[k].Low = datum[3].(float64)
		data[k].Close = datum[4].(float64)
		data[k].Volume = datum[5].(float64)
	}
	retVal.Data = data
	return retVal
}

// GetCandleData gets history of the specified symbol between a defined time-range
func (c *Client) GetCandleData(params *HistoryParams) (HistoryResponse, error) {
	var candleData ABResponse
	if !params.ValidDates() {
		return HistoryResponse{}, fmt.Errorf("history.GetCandleData: fromdate can not be greater than todate")
	}

	if !params.IsValidInterval() {
		return HistoryResponse{}, fmt.Errorf("history.GetCandleData: interval days can not be %d when interval is %s. Please see %s for details", params.IntervalDays(), params.Interval, URLHistoryDocumentation)
	}

	err := c.doEnvelope(http.MethodPost, URIGetCandleData, params.GetParams(), nil, &candleData, true)

	return candleData.Parse(), err
}
