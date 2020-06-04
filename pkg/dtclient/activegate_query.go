package dtclient

import (
	"encoding/json"
	"fmt"
)

type ActiveGateQuery struct {
	Hostname       string
	NetworkAddress string
	NetworkZone    string
	UpdateStatus   string
}

type ActiveGate struct {
	NetworkAddresses []string
	AutoUpdateStatus string
	OfflineSince     int64
	Version          string
	Hostname         string
	NetworkZone      string
}

func (dtc *dynatraceClient) QueryActiveGates(query ActiveGateQuery) ([]ActiveGate, error) {
	url := fmt.Sprintf("%s/v2/activeGates?%s", dtc.url, buildQueryParams(query))
	logger.Info("querying from url", "url", url)
	response, err := dtc.makeRequest(url, dynatraceApiToken)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logger.Error(err, "error closing response body")
		}
	}()

	data, err := dtc.getServerResponseData(response)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, err
	}

	result := []ActiveGate{}
	activegates, err := dtc.readResponseForActiveGates(data)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, err
	}

	for _, activegate := range activegates {
		if activegate.OfflineSince <= 0 {
			result = append(result, activegate)
		}
	}

	return result, nil
}

func (dtc *dynatraceClient) QueryOutdatedActiveGates(query ActiveGateQuery) ([]ActiveGate, error) {
	query.UpdateStatus = STATUS_OUTDATED
	return dtc.QueryActiveGates(query)
}

func buildQueryParams(query ActiveGateQuery) string {
	params := ""
	if query.Hostname != "" {
		params += "hostname=" + query.Hostname + "&"
	}
	if query.NetworkZone != "" {
		params += "networkZone=" + query.NetworkZone + "&"
	}
	if query.NetworkAddress != "" {
		params += "networkAddress=" + query.NetworkAddress + "&"
	}
	if query.UpdateStatus != "" {
		params += "updateStatus=" + query.UpdateStatus + "&"
	}

	params += "osType=" + OsLinux + "&" +
		"type=ENVIRONMENT"

	logger.Info("built parameters to query activegates", "params", params)
	return params
}

func (dtc *dynatraceClient) readResponseForActiveGates(data []byte) ([]ActiveGate, error) {
	type jsonResponse struct {
		ActiveGates []ActiveGate
	}

	jr := &jsonResponse{}
	err := json.Unmarshal(data, jr)
	if err != nil {
		logger.Error(err, "cannot unmarshal activegate response")
		return nil, err
	}
	return jr.ActiveGates, nil
}

const (
	OsLinux         = "LINUX"
	STATUS_OUTDATED = "OUTDATED"
)
