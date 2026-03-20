// Package agents provides type definitions for the Pinata Agents API.
package agents

// AgentStatus represents the possible statuses of an agent.
type AgentStatus string

const (
	AgentStatusStarting   AgentStatus = "starting"
	AgentStatusRunning    AgentStatus = "running"
	AgentStatusNotRunning AgentStatus = "not_running"
)

// Agent represents an AI agent in the system.
type Agent struct {
	AgentID            string  `json:"agentId"`
	UserID             string  `json:"userId"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	Vibe               *string `json:"vibe"`
	Emoji              *string `json:"emoji"`
	GatewayToken       string  `json:"gatewayToken"`
	CreatedAt          string  `json:"createdAt"`
	Status             string  `json:"status"`
	LastSync           *string `json:"lastSync"`
	SnapshotCid        *string `json:"snapshotCid"`
	FileManifest       *string `json:"fileManifest"`
	SnapshotsUpdatedAt *string `json:"snapshotsUpdatedAt"`
}

// CreateAgentBody is the request body for creating a new agent.
type CreateAgentBody struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Vibe        string   `json:"vibe,omitempty"`
	Emoji       string   `json:"emoji,omitempty"`
	SkillCids   []string `json:"skillCids,omitempty"`
	SecretIds   []string `json:"secretIds,omitempty"`
	TemplateID  string   `json:"templateId,omitempty"`
}

// CreateAgentResponse is the response from creating a new agent.
type CreateAgentResponse struct {
	Success bool   `json:"success"`
	Agent   Agent  `json:"agent"`
	Error   string `json:"error,omitempty"`
}

// AgentListResponse is the response from listing agents.
type AgentListResponse struct {
	Agents []Agent `json:"agents"`
}

// AgentDetailResponse is the response from getting agent details.
type AgentDetailResponse struct {
	Agent          Agent            `json:"agent"`
	ProcessStatus  string           `json:"processStatus"`
	ProcessID      *string          `json:"processId"`
	Skills         []AgentSkill     `json:"skills"`
	Secrets        []AgentSecret    `json:"secrets"`
	Snapshots      []AgentSnapshot  `json:"snapshots,omitempty"`
	PortForwarding []PortForwarding `json:"portForwarding,omitempty"`
}

// DeleteAgentResponse is the response from deleting an agent.
type DeleteAgentResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	WasInRegistry    bool   `json:"wasInRegistry"`
	R2ObjectsDeleted int    `json:"r2ObjectsDeleted"`
}

// --- Skills ---

// EnvVarDef defines an environment variable required by a skill.
type EnvVarDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Skill represents a skill in the library.
type Skill struct {
	SkillID     string      `json:"skillId"`
	SkillCid    string      `json:"skillCid"`
	Name        string      `json:"name"`
	Description *string     `json:"description"`
	CreatedAt   string      `json:"createdAt"`
	UserID      string      `json:"userId"`
	EnvVars     []EnvVarDef `json:"envVars"`
	FileID      *string     `json:"fileId"`
}

// CreateSkillBody is the request body for creating a skill.
type CreateSkillBody struct {
	SkillCid    string   `json:"skillCid"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	EnvVars     []string `json:"envVars,omitempty"`
	FileID      string   `json:"fileId,omitempty"`
}

// CreateSkillResponse is the response from creating a skill.
type CreateSkillResponse struct {
	Success bool  `json:"success"`
	Skill   Skill `json:"skill"`
}

// SkillListResponse is the response from listing skills.
type SkillListResponse struct {
	Skills        []Skill `json:"skills"`
	CurrentUserID string  `json:"currentUserId"`
}

// DeleteSkillResponse is the response from deleting a skill.
type DeleteSkillResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	WasInLibrary bool   `json:"wasInLibrary"`
}

// AgentSkill represents a skill attached to an agent.
type AgentSkill struct {
	AgentID     string  `json:"agentId"`
	SkillID     string  `json:"skillId"`
	SkillCid    string  `json:"skillCid"`
	AttachedAt  string  `json:"attachedAt"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// AddSkillsBody is the request body for adding skills to an agent.
type AddSkillsBody struct {
	SkillCids []string `json:"skillCids"`
}

// AddSkillsResponse is the response from adding skills to an agent.
type AddSkillsResponse struct {
	Success  bool `json:"success"`
	Attached int  `json:"attached"`
}

// --- Secrets ---

// Secret represents a user secret.
type Secret struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// SecretAgentRef is a reference to an agent using a secret.
type SecretAgentRef struct {
	AgentID string `json:"agentId"`
	Name    string `json:"name"`
}

// SecretWithAgents extends Secret with the list of agents using it.
type SecretWithAgents struct {
	Secret
	Agents []SecretAgentRef `json:"agents"`
}

// SecretListResponse is the response from listing secrets.
type SecretListResponse struct {
	Secrets []SecretWithAgents `json:"secrets"`
}

// CreateSecretBody is the request body for creating a secret.
type CreateSecretBody struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CreateSecretResponse is the response from creating a secret.
type CreateSecretResponse struct {
	Success bool   `json:"success"`
	Secret  Secret `json:"secret"`
}

// UpdateSecretBody is the request body for updating a secret.
type UpdateSecretBody struct {
	Value string `json:"value"`
}

// AgentSecret represents a secret attached to an agent.
type AgentSecret struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Synced    bool   `json:"synced"`
}

// AddSecretsBody is the request body for adding secrets to an agent.
type AddSecretsBody struct {
	SecretIds []string `json:"secretIds"`
}

// AddSecretsResponse is the response from adding secrets to an agent.
type AddSecretsResponse struct {
	Success  bool `json:"success"`
	Attached int  `json:"attached"`
}

// --- Channels ---

// TelegramStatus represents the Telegram channel configuration status.
type TelegramStatus struct {
	Configured   bool        `json:"configured"`
	Enabled      bool        `json:"enabled"`
	DmPolicy     string      `json:"dmPolicy"`
	AllowFrom    interface{} `json:"allowFrom"` // Can be array of strings or numbers
	BotTokenSet  bool        `json:"botTokenSet"`
	BotTokenHint *string     `json:"botTokenHint"`
}

// SlackStatus represents the Slack channel configuration status.
type SlackStatus struct {
	Configured   bool    `json:"configured"`
	Enabled      bool    `json:"enabled"`
	BotTokenSet  bool    `json:"botTokenSet"`
	BotTokenHint *string `json:"botTokenHint"`
	AppTokenSet  bool    `json:"appTokenSet"`
	AppTokenHint *string `json:"appTokenHint"`
}

// DiscordStatus represents the Discord channel configuration status.
type DiscordStatus struct {
	Configured   bool    `json:"configured"`
	Enabled      bool    `json:"enabled"`
	BotTokenSet  bool    `json:"botTokenSet"`
	BotTokenHint *string `json:"botTokenHint"`
}

// WhatsAppStatus represents the WhatsApp channel configuration status.
type WhatsAppStatus struct {
	Configured bool        `json:"configured"`
	Enabled    bool        `json:"enabled"`
	DmPolicy   string      `json:"dmPolicy"`
	AllowFrom  interface{} `json:"allowFrom"` // Can be array of strings or numbers
	Linked     bool        `json:"linked"`
}

// ChannelStatusResponse is the response from getting channel statuses.
type ChannelStatusResponse struct {
	Telegram *TelegramStatus `json:"telegram"`
	Slack    *SlackStatus    `json:"slack"`
	Discord  *DiscordStatus  `json:"discord"`
	WhatsApp *WhatsAppStatus `json:"whatsapp"`
}

// ConfigureChannelBody is the request body for configuring a channel.
type ConfigureChannelBody struct {
	BotToken  string   `json:"botToken,omitempty"`
	AppToken  string   `json:"appToken,omitempty"`
	DmPolicy  string   `json:"dmPolicy,omitempty"`
	AllowFrom []string `json:"allowFrom,omitempty"`
}

// ConfigureChannelResponse is the response from configuring a channel.
type ConfigureChannelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Devices ---

// PendingDevice represents a pending device pairing request.
type PendingDevice struct {
	RequestID string `json:"requestId"`
}

// DeviceListResponse is the response from listing devices.
type DeviceListResponse struct {
	Pending    []PendingDevice  `json:"pending"`
	Paired     []map[string]any `json:"paired"`
	Raw        string           `json:"raw,omitempty"`
	Stderr     string           `json:"stderr,omitempty"`
	ParseError string           `json:"parseError,omitempty"`
}

// ApproveDeviceResponse is the response from approving a single device.
type ApproveDeviceResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"requestId"`
	Message   string `json:"message"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
}

// ApproveFailure represents a failed device approval.
type ApproveFailure struct {
	RequestID string `json:"requestId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// ApproveAllResponse is the response from approving all pending devices.
type ApproveAllResponse struct {
	Approved []string         `json:"approved"`
	Failed   []ApproveFailure `json:"failed,omitempty"`
	Message  string           `json:"message"`
}

// --- Snapshots ---

// AgentSnapshot represents a workspace snapshot.
type AgentSnapshot struct {
	AgentID       string  `json:"agentId"`
	SnapshotCid   string  `json:"snapshotCid"`
	UserID        string  `json:"userId"`
	CommitSha     *string `json:"commitSha"`
	ChangeSummary *string `json:"changeSummary"`
	ContentDiff   *string `json:"contentDiff"`
	CreatedAt     string  `json:"createdAt"`
}

// AgentSnapshotsResponse is the response from getting agent snapshots.
type AgentSnapshotsResponse struct {
	AgentID   string          `json:"agentId"`
	Snapshots []AgentSnapshot `json:"snapshots"`
}

// StorageStatusResponse is the response from getting storage sync status.
type StorageStatusResponse struct {
	AgentID    string   `json:"agentId"`
	Configured bool     `json:"configured"`
	Missing    []string `json:"missing,omitempty"`
	LastSync   *string  `json:"lastSync"`
	Message    string   `json:"message"`
}

// StorageSyncResponse is the response from creating a snapshot.
type StorageSyncResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message,omitempty"`
	LastSync      string  `json:"lastSync,omitempty"`
	SnapshotCid   *string `json:"snapshotCid"`
	CommitSha     *string `json:"commitSha"`
	ChangeSummary *string `json:"changeSummary"`
	Error         string  `json:"error,omitempty"`
	Details       string  `json:"details,omitempty"`
}

// ResetSnapshotBody is the request body for resetting to a snapshot.
type ResetSnapshotBody struct {
	SnapshotCid string `json:"snapshotCid"`
}

// ResetSnapshotResponse is the response from resetting to a snapshot.
type ResetSnapshotResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	SnapshotCid string  `json:"snapshotCid"`
	CommitSha   *string `json:"commitSha"`
}

// --- Port Forwarding ---

// PortForwarding represents a port forwarding rule.
type PortForwarding struct {
	Port       int    `json:"port"`
	PathPrefix string `json:"pathPrefix"`
	Protected  bool   `json:"protected,omitempty"`
}

// PortForwardingResponse is the response from getting port forwarding rules.
type PortForwardingResponse struct {
	Mappings []PortForwarding `json:"mappings"`
}

// UpdatePortForwardingBody is the request body for updating port forwarding.
type UpdatePortForwardingBody struct {
	Mappings []PortForwarding `json:"mappings"`
}

// UpdatePortForwardingResponse is the response from updating port forwarding.
type UpdatePortForwardingResponse struct {
	Success  bool             `json:"success"`
	Mappings []PortForwarding `json:"mappings"`
}

// --- Console Exec ---

// ExecRequest is the request body for executing a command.
type ExecRequest struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}

// ExecResponse is the response from executing a command.
type ExecResponse struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exitCode"`
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}

// --- Logs and Restart ---

// LogsResponse is the response from getting agent logs.
type LogsResponse struct {
	Logs string `json:"logs"`
}

// RestartResponse is the response from restarting an agent.
type RestartResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	PreviousProcessID *string `json:"previousProcessId"`
}

// --- Tasks/Cron ---

// ScheduleKind represents the type of cron schedule.
type ScheduleKind string

const (
	ScheduleKindAt    ScheduleKind = "at"
	ScheduleKindEvery ScheduleKind = "every"
	ScheduleKindCron  ScheduleKind = "cron"
)

// TaskSchedule represents a cron job schedule.
type TaskSchedule struct {
	Kind      ScheduleKind `json:"kind"`
	At        string       `json:"at,omitempty"`
	EveryMs   int          `json:"everyMs,omitempty"`
	Expr      string       `json:"expr,omitempty"`
	Tz        string       `json:"tz,omitempty"`
	StaggerMs int          `json:"staggerMs,omitempty"`
}

// PayloadKind represents the type of task payload.
type PayloadKind string

const (
	PayloadKindSystemEvent PayloadKind = "systemEvent"
	PayloadKindAgentTurn   PayloadKind = "agentTurn"
)

// TaskPayload represents the payload of a cron job.
type TaskPayload struct {
	Kind           PayloadKind `json:"kind"`
	Text           string      `json:"text,omitempty"`
	Message        string      `json:"message,omitempty"`
	Model          string      `json:"model,omitempty"`
	Thinking       string      `json:"thinking,omitempty"`
	TimeoutSeconds int         `json:"timeoutSeconds,omitempty"`
}

// DeliveryMode represents the delivery mode for task results.
type DeliveryMode string

const (
	DeliveryModeNone     DeliveryMode = "none"
	DeliveryModeAnnounce DeliveryMode = "announce"
	DeliveryModeWebhook  DeliveryMode = "webhook"
)

// TaskDelivery represents delivery settings for a cron job.
type TaskDelivery struct {
	Mode       DeliveryMode `json:"mode,omitempty"`
	Channel    string       `json:"channel,omitempty"`
	To         string       `json:"to,omitempty"`
	BestEffort bool         `json:"bestEffort,omitempty"`
}

// SessionTarget represents the session target for a task.
type SessionTarget string

const (
	SessionTargetMain     SessionTarget = "main"
	SessionTargetIsolated SessionTarget = "isolated"
)

// WakeMode represents when to wake the agent.
type WakeMode string

const (
	WakeModeNow           WakeMode = "now"
	WakeModeNextHeartbeat WakeMode = "next-heartbeat"
)

// CreateTaskBody is the request body for creating a cron job.
type CreateTaskBody struct {
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	Enabled       bool          `json:"enabled,omitempty"`
	Schedule      TaskSchedule  `json:"schedule"`
	SessionTarget SessionTarget `json:"sessionTarget,omitempty"`
	WakeMode      WakeMode      `json:"wakeMode,omitempty"`
	Payload       TaskPayload   `json:"payload"`
	Delivery      *TaskDelivery `json:"delivery,omitempty"`
}

// UpdateTaskBody is the request body for updating a cron job.
type UpdateTaskBody struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Enabled     *bool         `json:"enabled,omitempty"`
	Schedule    *TaskSchedule `json:"schedule,omitempty"`
	Payload     *TaskPayload  `json:"payload,omitempty"`
	Delivery    *TaskDelivery `json:"delivery,omitempty"`
}

// ToggleTaskBody is the request body for toggling a cron job.
type ToggleTaskBody struct {
	Enabled bool `json:"enabled"`
}

// --- Common Responses ---

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a simple success response.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// FeedbackResponse is the response from submitting feedback.
type FeedbackResponse struct {
	Success bool `json:"success"`
}

// FeedbackBody is the request body for submitting feedback.
type FeedbackBody struct {
	Message string `json:"message"`
}

// --- Templates ---

// TemplateRequiredSecret defines a secret required by a template.
type TemplateRequiredSecret struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	GuideURL    string   `json:"guideUrl,omitempty"`
	GuideSteps  []string `json:"guideSteps,omitempty"`
	Required    bool     `json:"required"`
}

// Template represents a pre-built agent template.
type Template struct {
	TemplateID        string                   `json:"templateId"`
	Name              string                   `json:"name"`
	Slug              string                   `json:"slug"`
	Description       string                   `json:"description"`
	LongDescription   *string                  `json:"longDescription"`
	PartnerName       string                   `json:"partnerName"`
	PartnerLogoURL    *string                  `json:"partnerLogoUrl"`
	PartnerURL        *string                  `json:"partnerUrl"`
	Category          string                   `json:"category"`
	Tags              []string                 `json:"tags"`
	SnapshotCid       string                   `json:"snapshotCid"`
	OpenclawVersion   string                   `json:"openclawVersion"`
	RequiredSecrets   []TemplateRequiredSecret `json:"requiredSecrets"`
	IncludedSkillCids []string                 `json:"includedSkillCids"`
	DefaultVibe       *string                  `json:"defaultVibe"`
	DefaultEmoji      *string                  `json:"defaultEmoji"`
	Featured          bool                     `json:"featured"`
	SortOrder         int                      `json:"sortOrder"`
	Status            string                   `json:"status"`
	Price             *string                  `json:"price"`
	PriceAsset        string                   `json:"priceAsset"`
	PriceNetwork      string                   `json:"priceNetwork"`
	PayToAddress      *string                  `json:"payToAddress"`
	IsFree            bool                     `json:"isFree"`
}

// TemplateListResponse is the response from listing templates.
type TemplateListResponse struct {
	Templates []Template `json:"templates"`
}

// TemplateDetailResponse is the response from getting a template.
type TemplateDetailResponse struct {
	Template Template `json:"template"`
}

// --- Config ---

// ConfigResponse is the response from getting agent config.
type ConfigResponse struct {
	Config interface{} `json:"config"`
}

// UpdateConfigBody is the request body for updating agent config.
type UpdateConfigBody struct {
	Config interface{} `json:"config"`
}

// ValidateConfigResponse is the response from validating agent config.
type ValidateConfigResponse struct {
	Valid  bool   `json:"valid"`
	Output string `json:"output"`
}

// --- Updates ---

// UpdateCheckResponse is the response from checking for openclaw updates.
type UpdateCheckResponse struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// UpdateApplyBody is the request body for applying an openclaw update.
type UpdateApplyBody struct {
	Tag string `json:"tag,omitempty"`
}

// UpdateApplyResponse is the response from applying an openclaw update.
type UpdateApplyResponse struct {
	Success         bool   `json:"success"`
	PreviousVersion string `json:"previousVersion"`
	NewVersion      string `json:"newVersion"`
	Output          string `json:"output"`
}

// --- Agent Versions ---

// VersionsResponse is the response from getting available agent versions.
type VersionsResponse struct {
	CurrentVersion    string   `json:"currentVersion"`
	AvailableVersions []string `json:"availableVersions"`
}

// --- ClawHub (Skills Marketplace) ---

// HubEnvVarDef defines an environment variable for a hub skill.
type HubEnvVarDef struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	GuideURL    string   `json:"guideUrl,omitempty"`
	GuideSteps  []string `json:"guideSteps,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// HubSkill represents a skill on ClawHub.
type HubSkill struct {
	HubSkillID      string         `json:"hubSkillId"`
	SkillCid        string         `json:"skillCid"`
	Name            string         `json:"name"`
	Slug            string         `json:"slug"`
	Description     string         `json:"description"`
	LongDescription *string        `json:"longDescription"`
	AuthorName      string         `json:"authorName"`
	AuthorURL       *string        `json:"authorUrl"`
	AuthorLogoURL   *string        `json:"authorLogoUrl"`
	Category        string         `json:"category"`
	Tags            []string       `json:"tags"`
	EnvVars         []HubEnvVarDef `json:"envVars"`
	ReadmeURL       *string        `json:"readmeUrl"`
	Featured        bool           `json:"featured"`
	InstallCount    int            `json:"installCount"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
}

// HubSkillListResponse is the response from listing hub skills.
type HubSkillListResponse struct {
	Skills     []HubSkill `json:"skills"`
	NextCursor *string    `json:"nextCursor"`
}

// HubSkillDetailResponse is the response from getting a hub skill.
type HubSkillDetailResponse struct {
	Skill HubSkill `json:"skill"`
}

// InstallHubSkillResponse is the response from installing a hub skill.
type InstallHubSkillResponse struct {
	Success bool  `json:"success"`
	Skill   Skill `json:"skill"`
}

// --- Custom Domains ---

// CustomDomain represents a custom domain mapping for an agent.
type CustomDomain struct {
	ID           string  `json:"id"`
	AgentID      string  `json:"agentId"`
	UserID       string  `json:"userId"`
	Subdomain    *string `json:"subdomain"`
	CustomDomain *string `json:"customDomain"`
	TargetPort   int     `json:"targetPort"`
	Protected    bool    `json:"protected"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// CustomDomainListResponse is the response from listing custom domains.
type CustomDomainListResponse struct {
	Domains []CustomDomain `json:"domains"`
}

// CreateCustomDomainBody is the request body for creating a custom domain.
type CreateCustomDomainBody struct {
	Subdomain    string `json:"subdomain,omitempty"`
	CustomDomain string `json:"customDomain,omitempty"`
	TargetPort   int    `json:"targetPort"`
	Protected    bool   `json:"protected,omitempty"`
}

// CreateCustomDomainResponse is the response from creating a custom domain.
type CreateCustomDomainResponse struct {
	Success bool         `json:"success"`
	Domain  CustomDomain `json:"domain"`
}

// UpdateCustomDomainBody is the request body for updating a custom domain.
type UpdateCustomDomainBody struct {
	Subdomain    string `json:"subdomain,omitempty"`
	CustomDomain string `json:"customDomain,omitempty"`
	TargetPort   *int   `json:"targetPort,omitempty"`
	Protected    *bool  `json:"protected,omitempty"`
}

// UpdateCustomDomainResponse is the response from updating a custom domain.
type UpdateCustomDomainResponse struct {
	Success bool         `json:"success"`
	Domain  CustomDomain `json:"domain"`
}

// DeleteCustomDomainResponse is the response from deleting a custom domain.
type DeleteCustomDomainResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
