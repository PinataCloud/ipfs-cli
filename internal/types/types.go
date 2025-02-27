package types

type UploadResponse struct {
	Data struct {
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
	} `json:"data"`
}

type Options struct {
	GroupId string `json:"group_id"`
}

type Metadata struct {
	Name string `json:"name"`
}

type File struct {
	Id            string                 `json:"id"`
	Name          string                 `json:"name"`
	Cid           string                 `json:"cid"`
	Size          int                    `json:"size"`
	NumberOfFiles int                    `json:"number_of_files"`
	MimeType      string                 `json:"mime_type"`
	KeyValues     map[string]interface{} `json:"keyvalues"`
	GroupId       *string                `json:"group_id,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

type FileUpdateBody struct {
	Name string `json:"name"`
}

type GetFileResponse struct {
	Data File `json:"data"`
}

type ListFilesData struct {
	Files         []File `json:"files"`
	NextPageToken string `json:"next_page_token"`
}

type ListResponse struct {
	Data ListFilesData `json:"data"`
}

type GroupResponseItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type GroupListResponse struct {
	Data struct {
		Groups        []GroupResponseItem `json:"groups"`
		NextPageToken string              `json:"next_page_token"`
	} `json:"data"`
}

type GroupCreateResponse struct {
	Data struct {
		GroupResponseItem
	} `json:"data"`
}

type GroupCreateBody struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type GetSignedURLBody struct {
	URL     string `json:"url"`
	Expires int    `json:"expires"`
	Date    int64  `json:"date"`
	Method  string `json:"method"`
}

type GetSignedURLResponse struct {
	Data string `json:"data"`
}

type GetGatewayItem struct {
	Domain string `json:"domain"`
}

type GetGatewaysResponse struct {
	Data struct {
		Rows []GetGatewayItem
	} `json:"data"`
}

type GetSwapHistoryResponse struct {
	Data []struct {
		MappedCid string `json:"mapped_cid"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

type AddSwapBody struct {
	SwapCid string `json:"swap_cid"`
}

type AddSwapResponse struct {
	Data struct {
		MappedCid string `json:"mapped_cid"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

type CreateKeyResponse struct {
	JWT             string `json:"JWT"`
	PinataAPIKey    string `json:"pinata_api_key"`
	PinataAPISecret string `json:"pinata_api_secret"`
}

type CreatKeyBody struct {
	KeyName     string      `json:"keyName"`
	Permissions Permissions `json:"permissions"`
	MaxUses     int         `json:"maxUses,omitempty"`
}

type Permissions struct {
	Admin     bool      `json:"admin,omitempty"`
	Endpoints Endpoints `json:"endpoints,omitempty"`
}

type Endpoints struct {
	Data    DataEndpoints    `json:"data,omitempty"`
	Pinning PinningEndpoints `json:"pinning,omitempty"`
}

type DataEndpoints struct {
	PinList             bool `json:"pinList,omitempty"`
	UserPinnedDataTotal bool `json:"userPinnedDataTotal,omitempty"`
}

type PinningEndpoints struct {
	HashMetadata  bool `json:"hashMetadata,omitempty"`
	HashPinPolicy bool `json:"hashPinPolicy,omitempty"`
	PinByHash     bool `json:"pinByHash,omitempty"`
	PinFileToIPFS bool `json:"pinFileToIPFS,omitempty"`
	PinJSONToIPFS bool `json:"pinJSONToIPFS,omitempty"`
	PinJobs       bool `json:"pinJobs,omitempty"`
	Unpin         bool `json:"unpin,omitempty"`
	UserPinPolicy bool `json:"userPinPolicy,omitempty"`
}

type KeyListResponse struct {
	Keys  []KeyItem `json:"keys"`
	Count int       `json:"count"`
}

type KeyItem struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Key       string      `json:"key"`
	Secret    string      `json:"secret"`
	MaxUses   int         `json:"max_uses"`
	Uses      int         `json:"uses"`
	UserID    string      `json:"user_id"`
	Scopes    Permissions `json:"scopes"`
	Revoked   bool        `json:"revoked"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updatedAt"`
}

type PinataOptions struct {
	CidVersion int    `json:"cidVersion"`
	GroupId    string `json:"groupId,omitempty"`
}

type PinataMetadata struct {
	Name      string            `json:"name"`
	KeyValues map[string]string `json:"keyvalues,omitempty"`
}
