package gocs

import (
	"encoding/json"
	"fmt"
	"strings"
)

type command struct {
	Name           string
	Isasync        bool
	ReturnsAnId    bool
	RequiredParams map[string]*apiParam
	OptionalParams map[string]*apiParam
}

type listApisResult struct {
	Listapis struct {
		Count int
		Api   []*apiCommand
	} `json:"listapisresponse"`
}

type apiCommand struct {
	Name     string
	Isasync  bool
	Params   []*apiParam
	Response []*apiResponse
}

type apiParam struct {
	Name     string
	Type     string
	Required bool
}

type apiResponse struct {
	Name  string
	Type  string
	Value string
}

type csError struct {
	Errorcode   int
	CsErrorcode int
	Errortext   string
}

type asyncResult struct {
	Id    string
	Jobid string
}

type asyncJobResult struct {
	Jobstatus     float64
	Jobresulttype string
	Jobresult     json.RawMessage
}

type responseID struct {
	Id string
}

// Generic function to get the first value of a response as json.RawMessage
func getRawValue(rawJSON json.RawMessage) (json.RawMessage, error) {
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(rawJSON, &rawData); err != nil {
		return nil, err
	}
	for _, rawValue := range rawData {
		return rawValue, nil
	}
	return nil, fmt.Errorf("Unable to extract the raw value from:\n\n%s\n\n", string(rawJSON))
}

func unmarshalCsError(rawJSON json.RawMessage) (*csError, error) {
	rawError, err := getRawValue(rawJSON)
	if err != nil {
		return nil, err
	}
	var errorInfo csError
	if err := json.Unmarshal(rawError, &errorInfo); err != nil {
		return nil, err
	}
	return &errorInfo, nil
}

func unmarshalListId(key string, rawJSON json.RawMessage) (string, error) {
	rawResponse, err := getRawValue(rawJSON)
	if err != nil {
		return "", err
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(rawResponse, &result); err != nil {
		return "", err
	}
	rawResponse, found := result[key]
	if !found {
		return "", fmt.Errorf("Unable to find id in result: %v", result)
	}
	var id []responseID
	if err := json.Unmarshal(rawResponse, &id); err != nil {
		return "", err
	}
	if len(id) > 1 {
		return "", fmt.Errorf("Multiple id's found in result: %s", string(rawResponse))
	}
	return id[0].Id, nil
}

func unmarshalId(key string, rawJSON json.RawMessage) (string, error) {
	rawResponse, err := getRawValue(rawJSON)
	if err != nil {
		return "", err
	}
	var id responseID
	if err := json.Unmarshal(rawResponse, &id); err != nil {
		return "", err
	}
	return id.Id, nil
}

func unmarshalAsyncResponse(rawJSON json.RawMessage) (*asyncResult, error) {
	rawResponse, err := getRawValue(rawJSON)
	if err != nil {
		return nil, err
	}
	var result asyncResult
	if err := json.Unmarshal(rawResponse, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func unmarshalAsyncJobResponse(rawJSON json.RawMessage) (*asyncJobResult, error) {
	rawResponse, err := getRawValue(rawJSON)
	if err != nil {
		return nil, err
	}
	var result asyncJobResult
	if err := json.Unmarshal(rawResponse, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func unmarshalApiCommands(rawJSON json.RawMessage) (commands, error) {
	var apisResult listApisResult
	if err := json.Unmarshal(rawJSON, &apisResult); err != nil {
		return nil, err
	}
	apiCommands := make(commands, apisResult.Listapis.Count)
	for _, apiCmd := range apisResult.Listapis.Api {
		cmd := command{apiCmd.Name, apiCmd.Isasync, false, map[string]*apiParam{}, map[string]*apiParam{}}
		for _, apiParam := range apiCmd.Params {
			if apiParam.Required {
				cmd.RequiredParams[strings.ToLower(apiParam.Name)] = apiParam
			} else {
				cmd.OptionalParams[strings.ToLower(apiParam.Name)] = apiParam
			}
		}
		if !strings.HasPrefix(apiCmd.Name, "list") {
			for _, apiResp := range apiCmd.Response {
				if strings.ToLower(apiResp.Name) == "id" {
					cmd.ReturnsAnId = true
					break
				}
			}
		}
		apiCommands[strings.ToLower(apiCmd.Name)] = &cmd
	}
	return apiCommands, nil
}
