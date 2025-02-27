package groups

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

func GetGroup(id string, network string) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s/%s", config.GetAPIHost(), networkParam, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.GroupCreateResponse{}, fmt.Errorf("server Returned an error %d, check CID", resp.StatusCode)
	}
	var response types.GroupCreateResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.GroupCreateResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func ListGroups(amount string, name string, token string, network string) (types.GroupListResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupListResponse{}, err
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return types.GroupListResponse{}, err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s?", config.GetAPIHost(), networkParam)

	params := []string{}

	if amount != "" {
		params = append(params, "limit="+amount)
	}

	if name != "" {
		params = append(params, "name="+name)
	}

	if token != "" {
		params = append(params, "pageToken="+token)
	}

	if len(params) > 0 {
		url += strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.GroupListResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GroupListResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.GroupListResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.GroupListResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.GroupListResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.GroupListResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func CreateGroup(name string, network string) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	payload := types.GroupCreateBody{
		Name: name,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}

	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s", config.GetAPIHost(), networkParam)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.GroupCreateResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.GroupCreateResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.GroupCreateResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func UpdateGroup(id string, name string, network string) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	payload := types.GroupCreateBody{
		Name: name,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s/%s", config.GetAPIHost(), networkParam, id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.GroupCreateResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.GroupCreateResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.GroupCreateResponse{}, err
	}
	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.GroupCreateResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil

}

func DeleteGroup(id string, network string) error {
	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s/%s", config.GetAPIHost(), networkParam, id)

	req, err := http.NewRequest("DELETE", url, nil)
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

	fmt.Println("Group Deleted")

	return nil

}

func AddFile(groupId string, fileId string, network string) error {

	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s/%s/ids/%s", config.GetAPIHost(), networkParam, groupId, fileId)

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

	fmt.Println("File added to group")

	return nil
}

func RemoveFile(groupId string, fileId string, network string) error {

	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/v3/groups/%s/%s/ids/%s", config.GetAPIHost(), networkParam, groupId, fileId)

	req, err := http.NewRequest("DELETE", url, nil)
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

	fmt.Println("File removed from group")

	return nil
}
