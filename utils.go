package smartapigo

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type Time struct {
	time.Time
}

type Exchange string

type TimeInterval string

const (
	// TimeFormatLayout for History API as described in the NOTE section of https://smartapi.angelbroking.com/docs/Historical
	TimeFormatLayout string = "2006-01-02 15:04"

	// NSE Exchange constant for NSE Equity as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	NSE Exchange = "NSE"

	// NFO Exchange constant for Only available for NSE Futures as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	NFO Exchange = "NFO"

	// ONE_MINUTE interval constant for "1 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	ONE_MINUTE TimeInterval = "ONE_MINUTE"

	// THREE_MINUTE interval constant for "3 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	THREE_MINUTE TimeInterval = "THREE_MINUTE"

	// FIVE_MINUTE interval constant for "5 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	FIVE_MINUTE TimeInterval = "FIVE_MINUTE"

	// TEN_MINUTE interval constant for "10 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	TEN_MINUTE TimeInterval = "TEN_MINUTE"

	// FIFTEEN_MINUTE interval constant for "15 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	FIFTEEN_MINUTE TimeInterval = "FIFTEEN_MINUTE"

	// THIRTY_MINUTE interval constant for "30 Minute" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	THIRTY_MINUTE TimeInterval = "THIRTY_MINUTE"

	// ONE_HOUR interval constant for "1 Hour" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	ONE_HOUR TimeInterval = "ONE_HOUR"

	// ONE_DAY interval constant for "1 Day" as defined in documentation at: https://smartapi.angelbroking.com/docs/Historical
	ONE_DAY TimeInterval = "ONE_DAY"
)

// API endpoints
const (
	URILogin                string = "rest/auth/angelbroking/user/v1/loginByPassword"
	URIUserSessionRenew     string = "rest/auth/angelbroking/jwt/v1/generateTokens"
	URIUserProfile          string = "rest/secure/angelbroking/user/v1/getProfile"
	URILogout               string = "rest/secure/angelbroking/user/v1/logout"
	URIGetOrderBook         string = "rest/secure/angelbroking/order/v1/getOrderBook"
	URIPlaceOrder           string = "rest/secure/angelbroking/order/v1/placeOrder"
	URIModifyOrder          string = "rest/secure/angelbroking/order/v1/modifyOrder"
	URICancelOrder          string = "rest/secure/angelbroking/order/v1/cancelOrder"
	URIGetHoldings          string = "rest/secure/angelbroking/portfolio/v1/getHolding"
	URIGetPositions         string = "rest/secure/angelbroking/order/v1/getPosition"
	URIGetTradeBook         string = "rest/secure/angelbroking/order/v1/getTradeBook"
	URILTP                  string = "rest/secure/angelbroking/order/v1/getLtpData"
	URIRMS                  string = "rest/secure/angelbroking/user/v1/getRMS"
	URIConvertPosition      string = "rest/secure/angelbroking/order/v1/convertPosition"
	URIGetCandleData        string = "rest/secure/angelbroking/historical/v1/getCandleData"
	URLHistoryDocumentation string = "https://smartapi.angelbroking.com/docs/Historical"
)

func structToMap(obj interface{}, tagName string) map[string]interface{} {
	var values reflect.Value
	switch obj.(type) {
	case OrderParams:
		{
			con := obj.(OrderParams)
			values = reflect.ValueOf(&con).Elem()
		}

	case ModifyOrderParams:
		{
			con := obj.(ModifyOrderParams)
			values = reflect.ValueOf(&con).Elem()
		}
	case LTPParams:
		{
			con := obj.(LTPParams)
			values = reflect.ValueOf(&con).Elem()
		}
	case ConvertPositionParams:
		{
			con := obj.(ConvertPositionParams)
			values = reflect.ValueOf(&con).Elem()
		}
	}

	tags := reflect.TypeOf(obj)
	params := make(map[string]interface{})
	for i := 0; i < values.NumField(); i++ {
		params[tags.Field(i).Tag.Get(tagName)] = values.Field(i).Interface()
	}

	return params
}

func getIpAndMac() (string, string, string, error) {

	//----------------------
	// Get the local machine IP address
	//----------------------

	var localIp, currentNetworkHardwareName string

	localIp, err := getLocalIP()

	if err != nil {
		return "", "", "", err
	}

	// get all the system's or local machine's network interfaces

	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {

		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {

				// only interested in the name with current IP address
				if strings.Contains(addr.String(), localIp) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}

	// extract the hardware information base on the interface name
	// capture above
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)

	if err != nil {
		return "", "", "", err
	}

	macAddress := netInterface.HardwareAddr

	// verify if the MAC address can be parsed properly
	_, err = net.ParseMAC(macAddress.String())

	if err != nil {
		return "", "", "", err
	}

	publicIp, err := getPublicIp()
	if err != nil {
		return "", "", "", err
	}

	return localIp, publicIp, macAddress.String(), nil

}

func getLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("please check your network connection")
}

func getPublicIp() (string, error) {
	resp, err := http.Get("https://myexternalip.com/raw")
	if err != nil {
		return "", err
	}

	content, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func secondsToDays(seconds int64) int64 {
	return int64(((seconds / 60) / 60) / 24)
}
