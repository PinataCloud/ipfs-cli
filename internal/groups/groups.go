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

func GetGroup(id string) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}
	url := fmt.Sprintf("https://%s/v3/files/groups/%s", config.GetAPIHost(), id)

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

func ListGroups(amount string, isPublic bool, name string, token string) (types.GroupListResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupListResponse{}, err
	}
	url := fmt.Sprintf("https://%s/v3/files/groups?", config.GetAPIHost())

	params := []string{}

	if amount != "" {
		params = append(params, "limit="+amount)
	}

	if isPublic {
		params = append(params, "isPublic=true")
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

func CreateGroup(name string, isPublic bool) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	payload := types.GroupCreateBody{
		Name:     name,
		IsPublic: isPublic,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}

	url := fmt.Sprintf("https://%s/v3/files/groups", config.GetAPIHost())
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

func UpdateGroup(id string, name string, isPublic bool) (types.GroupCreateResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.GroupCreateResponse{}, err
	}

	payload := types.GroupCreateBody{
		Name:     name,
		IsPublic: isPublic,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.GroupCreateResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}

	url := fmt.Sprintf("https://%s/v3/files/groups/%s", config.GetAPIHost(), id)
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

func DeleteGroup(id string) error {
	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/v3/files/groups/%s", config.GetAPIHost(), id)

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

func AddFile(groupId string, fileId string) error {

	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/v3/files/groups/%s/ids/%s", config.GetAPIHost(), groupId, fileId)

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

func RemoveFile(groupId string, fileId string) error {

	jwt, err := common.FindToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/v3/files/groups/%s/ids/%s", config.GetAPIHost(), groupId, fileId)

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
