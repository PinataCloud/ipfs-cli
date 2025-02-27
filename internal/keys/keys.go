package keys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pinata/internal/common"
	"pinata/internal/config"
	"pinata/internal/types"
	"strings"
)

func ListKeys(name string, revoked bool, limitedUse bool, exhausted bool, offset string) (types.KeyListResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.KeyListResponse{}, err
	}
	url := fmt.Sprintf("https://%s/v3/pinata/keys?", config.GetAPIHost())

	params := []string{}

	if name != "" {
		params = append(params, "name="+name)
	}

	if revoked {
		params = append(params, "revoked=true")
	}

	if limitedUse {
		params = append(params, "limitedUse=true")
	}

	if exhausted {
		params = append(params, "exhausted=true")
	}
	if offset != "" {
		params = append(params, "offset="+offset)
	}

	if len(params) > 0 {
		url += strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.KeyListResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.KeyListResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.KeyListResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.KeyListResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.KeyListResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return types.KeyListResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func CreateKey(name string, admin bool, uses int, endpoints []string) (types.CreateKeyResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.CreateKeyResponse{}, err
	}

	payload := types.CreatKeyBody{
		KeyName: name,
		Permissions: types.Permissions{
			Admin: admin,
		},
	}

	if uses > 0 {
		payload.MaxUses = uses
	}

	if !admin && len(endpoints) > 0 {
		dataEndpoints := types.DataEndpoints{}
		pinningEndpoints := types.PinningEndpoints{}

		for _, endpoint := range endpoints {
			switch endpoint {
			case "pinList":
				dataEndpoints.PinList = true
			case "userPinnedDataTotal":
				dataEndpoints.UserPinnedDataTotal = true

			case "hashMetadata":
				pinningEndpoints.HashMetadata = true
			case "hashPinPolicy":
				pinningEndpoints.HashPinPolicy = true
			case "pinByHash":
				pinningEndpoints.PinByHash = true
			case "pinFileToIPFS":
				pinningEndpoints.PinFileToIPFS = true
			case "pinJSONToIPFS":
				pinningEndpoints.PinJSONToIPFS = true
			case "pinJobs":
				pinningEndpoints.PinJobs = true
			case "unpin":
				pinningEndpoints.Unpin = true
			case "userPinPolicy":
				pinningEndpoints.UserPinPolicy = true
			}
		}

		payload.Permissions.Endpoints = types.Endpoints{
			Data:    dataEndpoints,
			Pinning: pinningEndpoints,
		}
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.CreateKeyResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}

	url := fmt.Sprintf("https://%s/v3/pinata/keys", config.GetAPIHost())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return types.CreateKeyResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.CreateKeyResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.CreateKeyResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.CreateKeyResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.CreateKeyResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return types.CreateKeyResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func RevokeKey(id string) error {
	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/v3/pinata/keys/%s", config.GetAPIHost(), id)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("server Returned an error %d, check CID", resp.StatusCode)
	}

	fmt.Println("Key Revoked")

	return nil

}
