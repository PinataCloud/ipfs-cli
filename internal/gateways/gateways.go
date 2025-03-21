package gateways

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"pinata/internal/common"
	"pinata/internal/config"
	"pinata/internal/types"
	"pinata/internal/utils"
	"runtime"
	"strings"
	"time"
)

func FindGatewayDomain() ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dotFilePath := filepath.Join(homeDir, ".pinata-files-cli-gateway")
	Domain, err := os.ReadFile(dotFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("JWT not found. Please authorize first using the 'auth' command")
		} else {
			return nil, err
		}
	}
	return Domain, err
}

func SetGateway(domain string) error {
	if domain == "" {
		jwt, err := common.FindToken()
		if err != nil {
			return err
		}
		url := fmt.Sprintf("https://%s/v3/ipfs/gateways", config.GetAPIHost())

		req, err := http.NewRequest("GET", url, nil)
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
			return fmt.Errorf("server Returned an error %d", resp.StatusCode)
		}
		var response types.GetGatewaysResponse

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return err
		}

		options := make([]string, len(response.Data.Rows))
		for i, item := range response.Data.Rows {
			options[i] = item.Domain + ".mypinata.cloud"
		}
		domain, err := utils.MultiSelect(options)
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		p := filepath.Join(home, ".pinata-files-cli-gateway")
		err = os.WriteFile(p, []byte(domain), 0600)
		if err != nil {
			return err
		}
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, ".pinata-files-cli-gateway")
	err = os.WriteFile(p, []byte(domain), 0600)
	if err != nil {
		return err
	}

	fmt.Println("Gateway Saved!")
	return nil
}

func GetAccessLink(cid string, expires int, network string) (types.GetSignedURLResponse, error) {

	jwt, err := common.FindToken()
	if err != nil {
		return types.GetSignedURLResponse{}, err
	}

	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return types.GetSignedURLResponse{}, err
	}

	domain, err := FindGatewayDomain()
	if err != nil {
		return types.GetSignedURLResponse{}, err
	}

	if networkParam == "public" {
		url := fmt.Sprintf("https://%s/ipfs/%s", domain, cid)
		fmt.Println(url)
		return types.GetSignedURLResponse{}, err
	}

	domainUrl := fmt.Sprintf("https://%s/files/%s", domain, cid)

	currentTime := time.Now().Unix()

	payload := types.GetSignedURLBody{
		URL:     domainUrl,
		Expires: expires,
		Date:    currentTime,
		Method:  "GET",
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return types.GetSignedURLResponse{}, errors.Join(err, errors.New("Failed to marshal paylod"))
	}

	url := fmt.Sprintf("https://%s/v3/files/private/download_link", config.GetAPIHost())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return types.GetSignedURLResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GetSignedURLResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.GetSignedURLResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	var response types.GetSignedURLResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.GetSignedURLResponse{}, err
	}

	unescapedURL := strings.ReplaceAll(response.Data, "\\u0026", "&")
	unescapedURL = strings.Trim(unescapedURL, "\"")

	fmt.Println(unescapedURL)

	return response, nil
}

func OpenCID(cid string, network string) error {
	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return fmt.Errorf("problem getting network parameter: %w", err)
	}

	var url string
	if networkParam == "public" {
		domain, err := FindGatewayDomain()
		if err != nil {
			return fmt.Errorf("problem finding gateway domain: %w", err)
		}
		url = fmt.Sprintf("https://%s/ipfs/%s", string(domain), cid)
	} else {
		resp, err := GetAccessLink(cid, 30, networkParam)
		if err != nil {
			return fmt.Errorf("problem creating URL: %w", err)
		}

		url = resp.Data
		url = strings.ReplaceAll(url, "\\u0026", "&")
		url = strings.Trim(url, "\"")
	}

	fmt.Printf("Opening URL: %s\n", url)

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		return fmt.Errorf("problem opening URL: %w", err)
	}

	return nil
}
