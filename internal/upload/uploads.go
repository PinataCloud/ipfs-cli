package uploads

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"pinata/internal/common"
	"pinata/internal/config"
	cliConfig "pinata/internal/config"
	"pinata/internal/types"
	"runtime"
	"strings"
	"time"

	"github.com/eventials/go-tus"
	"github.com/schollz/progressbar/v3"
)

const (
	MAX_SIZE_REGULAR_UPLOAD = 100 * 1024 * 1024 // Uploead threshold
	CHUNK_SIZE              = 10 * 1024 * 1024  // Chunk size
)

func Upload(filePath string, groupId string, name string, verbose bool, network string) (types.UploadResponse, error) {

	stats, err := os.Stat(filePath)
	if err != nil {
		return types.UploadResponse{}, err
	}

	if stats.IsDir() {
		// Check if network is private and return error if so
		networkParam, err := config.GetNetworkParam(network)
		if err != nil {
			return types.UploadResponse{}, err
		}

		if networkParam == "private" {
			return types.UploadResponse{}, errors.New("folders are not supported on the private network")
		}

		// For folders, we use a different API endpoint
		return folderUpload(filePath, groupId, name, verbose)
	}

	if stats.Size() > MAX_SIZE_REGULAR_UPLOAD {
		return uploadWithTUS(filePath, groupId, name, verbose, stats, network)
	}

	return regularUpload(filePath, groupId, name, verbose, network)
}

type progressReader struct {
	r   io.Reader
	bar *progressbar.ProgressBar
}

func regularUpload(filePath string, groupId string, name string, verbose bool, network string) (types.UploadResponse, error) {

	jwt, err := common.FindToken()
	if err != nil {
		return types.UploadResponse{}, err
	}
	stats, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Println("File or folder does not exist")
		return types.UploadResponse{}, errors.Join(err, errors.New("file or folder does not exist"))
	}
	files, err := pathsFinder(filePath, stats)
	if err != nil {
		return types.UploadResponse{}, err
	}
	body := &bytes.Buffer{}
	contentType, err := createMultipartRequest(filePath, files, body, stats, groupId, name, network)
	if err != nil {
		return types.UploadResponse{}, err
	}

	var requestBody io.Reader
	if !verbose {
		requestBody = body
	} else {
		totalSize := int64(body.Len())
		fmt.Printf("Uploading %s (%s)\n", stats.Name(), formatSize(int(totalSize)))
		requestBody = newProgressReader(body, totalSize)
	}

	url := fmt.Sprintf("https://%s/v3/files", cliConfig.GetUploadsHost())
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return types.UploadResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.UploadResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	if resp.StatusCode != 200 {
		return types.UploadResponse{}, fmt.Errorf("server Returned an error %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var response types.UploadResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return types.UploadResponse{}, err
	}

	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.UploadResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil
}

func cmpl() {
	fmt.Println()
}

func newProgressReader(r io.Reader, size int64) *progressReader {
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("Uploading..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
		progressbar.OptionOnCompletion(cmpl),
	)
	return &progressReader{r: r, bar: bar}
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	if err != nil {
		return 0, err
	}
	err = pr.bar.Add(n)
	if err != nil {
		return 0, err
	}
	return
}

func formatSize(bytes int) string {
	const (
		KB = 1000
		MB = KB * KB
		GB = MB * KB
	)

	var formattedSize string

	switch {
	case bytes < KB:
		formattedSize = fmt.Sprintf("%d bytes", bytes)
	case bytes < MB:
		formattedSize = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	case bytes < GB:
		formattedSize = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	default:
		formattedSize = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	}

	return formattedSize
}

func uploadWithTUS(filePath string, groupId string, name string, verbose bool, stats os.FileInfo, network string) (types.UploadResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.UploadResponse{}, err
	}

	// Create the TUS client with config
	config := &tus.Config{
		ChunkSize:  CHUNK_SIZE, // 50MB chunks
		Resume:     false,
		Header:     http.Header{"Authorization": {fmt.Sprintf("Bearer %s", jwt)}},
		HttpClient: http.DefaultClient,
	}

	uploadHost := cliConfig.GetUploadsHost()
	url := fmt.Sprintf("https://%s/v3/files", uploadHost)
	client, err := tus.NewClient(url, config)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to create TUS client: %w", err)
	}

	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	networkParam, err := cliConfig.GetNetworkParam(network)
	if err != nil {
		return types.UploadResponse{}, err
	}

	// Create metadata
	metadata := map[string]string{
		"filename": filepath.Base(filePath),
		"network":  networkParam,
	}
	if groupId != "" {
		metadata["group_id"] = groupId
	}
	if name != "nil" {
		metadata["filename"] = name
	}

	// Create the upload
	upload := tus.NewUpload(f, stats.Size(), metadata, "")

	// Create and configure the uploader
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to create upload: %w", err)
	}

	var bar *progressbar.ProgressBar
	if verbose {
		fmt.Printf("Starting upload of %s (%s)\n", stats.Name(), formatSize(int(stats.Size())))
		bar = progressbar.NewOptions64(
			stats.Size(),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetDescription("Uploading..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerPadding: " ",
				BarStart:      "|",
				BarEnd:        "|",
			}),
			progressbar.OptionOnCompletion(cmpl),
		)

		go func() {
			for {
				offset := uploader.Offset()
				if offset >= stats.Size() {
					return
				}
				bar.Set64(offset)
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	err = uploader.Upload()
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed during upload: %w", err)
	}

	if verbose {
		fmt.Println("\nUpload completed!")
	}

	uploadURL := uploader.Url()
	urlParts := strings.Split(uploadURL, "/")
	fileId := urlParts[len(urlParts)-2]

	apiURL := fmt.Sprintf("https://%s/v3/files/%s", cliConfig.GetAPIHost(), fileId)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to create response request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to fetch upload response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response types.UploadResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.UploadResponse{}, fmt.Errorf("failed to format response: %w", err)
	}
	fmt.Println(string(formattedJSON))

	return response, nil
}

func folderUpload(filePath string, groupId string, name string, verbose bool) (types.UploadResponse, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return types.UploadResponse{}, err
	}

	stats, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return types.UploadResponse{}, errors.Join(err, errors.New("folder does not exist"))
	}

	files, err := pathsFinder(filePath, stats)
	if err != nil {
		return types.UploadResponse{}, err
	}

	body := &bytes.Buffer{}
	contentType, err := createPinataMultipartRequest(filePath, files, body, stats, groupId, name)
	if err != nil {
		return types.UploadResponse{}, err
	}

	var requestBody io.Reader
	if !verbose {
		requestBody = body
	} else {
		totalSize := int64(body.Len())
		fmt.Printf("Uploading folder %s (%s)\n", stats.Name(), formatSize(int(totalSize)))
		requestBody = newProgressReader(body, totalSize)
	}

	// Use the pinning endpoint for folders
	url := fmt.Sprintf("https://%s/pinning/pinFileToIPFS", cliConfig.GetAPIHost())
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return types.UploadResponse{}, errors.Join(err, errors.New("failed to create the request"))
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("content-type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.UploadResponse{}, errors.Join(err, errors.New("failed to send the request"))
	}
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return types.UploadResponse{}, fmt.Errorf("server returned an error %d: %s", resp.StatusCode, string(respBody))
	}

	defer resp.Body.Close()

	// Parse the pinning API response
	var pinningResponse struct {
		ID            string            `json:"ID"`
		Name          string            `json:"Name"`
		IpfsHash      string            `json:"IpfsHash"`
		PinSize       int               `json:"PinSize"`
		Timestamp     string            `json:"Timestamp"`
		NumberOfFiles int               `json:"NumberOfFiles"`
		MimeType      string            `json:"MimeType"`
		GroupId       *string           `json:"GroupId"`
		Keyvalues     map[string]string `json:"Keyvalues"`
		IsDuplicate   bool              `json:"isDuplicate"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pinningResponse)
	if err != nil {
		return types.UploadResponse{}, err
	}

	// Map the pinning API response to our UploadResponse format following the TypeScript mapping
	response := types.UploadResponse{
		Data: struct {
			Id            string            `json:"id"`
			Name          string            `json:"name"`
			Cid           string            `json:"cid"`
			Size          int               `json:"size"`
			CreatedAt     string            `json:"created_at"`
			NumberOfFiles int               `json:"number_of_files"`
			MimeType      string            `json:"mime_type"`
			GroupId       *string           `json:"group_id"`
			KeyValues     map[string]string `json:"keyvalues"`
			Vectorized    bool              `json:"vectorized"`
			Network       string            `json:"network"`
			IsDuplicate   bool              `json:"is_duplicate,omitempty"`
		}{
			Id:            pinningResponse.ID,
			Name:          pinningResponse.Name,
			Cid:           pinningResponse.IpfsHash,
			Size:          pinningResponse.PinSize,
			CreatedAt:     pinningResponse.Timestamp,
			NumberOfFiles: pinningResponse.NumberOfFiles,
			MimeType:      pinningResponse.MimeType,
			GroupId:       pinningResponse.GroupId,
			KeyValues:     pinningResponse.Keyvalues,
			Vectorized:    false,
			Network:       "public",
			IsDuplicate:   pinningResponse.IsDuplicate,
		},
	}

	// If groupId is specified, set it in the response
	if groupId != "" {
		response.Data.GroupId = &groupId
	}

	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return types.UploadResponse{}, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response, nil
}

func createMultipartRequest(filePath string, files []string, body io.Writer, stats os.FileInfo, groupId string, name string, network string) (string, error) {
	contentType := ""
	writer := multipart.NewWriter(body)

	fileIsASingleFile := !stats.IsDir()
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return contentType, err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatal("could not close file")
			}
		}(file)

		var part io.Writer
		if fileIsASingleFile {
			part, err = writer.CreateFormFile("file", filepath.Base(f))
		} else {
			relPath, _ := filepath.Rel(filePath, f)
			part, err = writer.CreateFormFile("file", filepath.Join(stats.Name(), relPath))
		}
		if err != nil {
			return contentType, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return contentType, err
		}
	}

	networkParam, err := config.GetNetworkParam(network)
	if err != nil {
		return "", err
	}

	err = writer.WriteField("network", networkParam)

	if groupId != "" {
		err := writer.WriteField("group_id", groupId)
		if err != nil {
			return contentType, err
		}
	}

	nameToUse := stats.Name()
	if name != "nil" {
		nameToUse = name
	}
	err = writer.WriteField("name", nameToUse)
	if err != nil {
		return contentType, err
	}

	err = writer.Close()
	if err != nil {
		return contentType, err
	}

	contentType = writer.FormDataContentType()

	return contentType, nil
}

func createPinataMultipartRequest(filePath string, files []string, body io.Writer, stats os.FileInfo, groupId string, name string) (string, error) {
	contentType := ""
	writer := multipart.NewWriter(body)

	// Add files to the multipart request
	fileIsASingleFile := !stats.IsDir()
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return contentType, err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatal("could not close file")
			}
		}(file)

		var part io.Writer
		if fileIsASingleFile {
			part, err = writer.CreateFormFile("file", filepath.Base(f))
		} else {
			relPath, _ := filepath.Rel(filePath, f)
			if runtime.GOOS == "windows" {
				relPathForward := strings.ReplaceAll(relPath, "\\", "/")
				folderName := stats.Name()
				folderNameForward := strings.ReplaceAll(folderName, "\\", "/")
				fullPath := folderNameForward
				if relPathForward != "" {
					fullPath = folderNameForward + "/" + relPathForward
				}
				part, err = writer.CreateFormFile("file", fullPath)
			} else {
				part, err = writer.CreateFormFile("file", filepath.Join(stats.Name(), relPath))
			}
		}
		if err != nil {
			return contentType, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return contentType, err
		}
	}

	// Create and add PinataOptions
	pinataOptions := types.PinataOptions{
		CidVersion: 1, // Default CID version
	}

	// Add groupId to options if provided
	if groupId != "" {
		pinataOptions.GroupId = groupId
	}

	optionsBytes, err := json.Marshal(pinataOptions)
	if err != nil {
		return contentType, err
	}

	err = writer.WriteField("pinataOptions", string(optionsBytes))
	if err != nil {
		return contentType, err
	}

	// Create and add PinataMetadata
	nameToUse := stats.Name()
	if name != "nil" {
		nameToUse = name
	}

	pinataMetadata := types.PinataMetadata{
		Name:      nameToUse,
		KeyValues: make(map[string]string), // Empty keyvalues for now
	}

	metadataBytes, err := json.Marshal(pinataMetadata)
	if err != nil {
		return contentType, err
	}

	err = writer.WriteField("pinataMetadata", string(metadataBytes))
	if err != nil {
		return contentType, err
	}

	err = writer.Close()
	if err != nil {
		return contentType, err
	}

	contentType = writer.FormDataContentType()
	return contentType, nil
}

func pathsFinder(filePath string, stats os.FileInfo) ([]string, error) {
	var err error
	files := make([]string, 0)
	fileIsASingleFile := !stats.IsDir()
	if fileIsASingleFile {
		files = append(files, filePath)
		return files, err
	}
	err = filepath.Walk(filePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return files, err
}
