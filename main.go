package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"pinata/internal/agents"
	"pinata/internal/agents/chat"
	"pinata/internal/auth"
	"pinata/internal/config"
	"pinata/internal/files"
	"pinata/internal/gateways"
	"pinata/internal/groups"
	"pinata/internal/keys"
	uploads "pinata/internal/upload"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "pinata",
		Usage: "The official Pinata IPFS CLI! To get started make an API key at https://app.pinata.cloud/keys, then authorize the CLI with the auth command with your JWT",
		Commands: []*cli.Command{
			{
				Name:      "auth",
				Aliases:   []string{"a"},
				Usage:     "Authorize the CLI with your Pinata JWT",
				ArgsUsage: "[your Pinata JWT]",
				Action: func(ctx *cli.Context) error {
					err := auth.SaveJWT()
					return err
				},
			},
			{
				Name:      "upload",
				Aliases:   []string{"u"},
				Usage:     "Upload a file to Pinata",
				ArgsUsage: "[path to file]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "group",
						Aliases: []string{"g"},
						Value:   "",
						Usage:   "Upload a file to a specific group by passing in the groupId",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Value:   "nil",
						Usage:   "Add a name for the file you are uploading. By default it will use the filename on your system.",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Show upload progress",
					},
					&cli.StringFlag{
						Name:    "network",
						Aliases: []string{"net"},
						Usage:   "Specify the network (public or private). Uses default if not specified",
					},
				},
				Action: func(ctx *cli.Context) error {
					filePath := ctx.Args().First()
					groupId := ctx.String("group")
					name := ctx.String("name")
					verbose := ctx.Bool("verbose")
					network := ctx.String("network")
					if filePath == "" {
						return errors.New("no file path provided")
					}
					_, err := uploads.Upload(filePath, groupId, name, verbose, network)
					return err
				},
			},
			{
				Name:    "groups",
				Aliases: []string{"g"},
				Usage:   "Interact with file groups",
				Subcommands: []*cli.Command{
					{
						Name:      "create",
						Aliases:   []string{"c"},
						Usage:     "Create a new group",
						ArgsUsage: "[name of group]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
							network := ctx.String("network")
							if name == "" {
								return errors.New("Group name required")
							}
							_, err := groups.CreateGroup(name, network)
							return err
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "List groups on your account",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "amount",
								Aliases: []string{"a"},
								Value:   "10",
								Usage:   "The number of groups you would like to return",
							},
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Filter groups by name",
							},
							&cli.StringFlag{
								Name:    "token",
								Aliases: []string{"t"},
								Usage:   "Paginate through results using the pageToken",
							},
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							amount := ctx.String("amount")
							name := ctx.String("name")
							token := ctx.String("token")
							network := ctx.String("network")
							_, err := groups.ListGroups(amount, name, token, network)
							return err
						},
					},
					{
						Name:      "update",
						Aliases:   []string{"u"},
						Usage:     "Update a group",
						ArgsUsage: "[ID of group]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Update the name of a group",
							},
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							groupId := ctx.Args().First()
							name := ctx.String("name")
							network := ctx.String("network")
							if groupId == "" {
								return errors.New("no ID provided")
							}
							_, err := groups.UpdateGroup(groupId, name, network)
							return err
						},
					},
					{
						Name:      "delete",
						Aliases:   []string{"d"},
						Usage:     "Delete a group by ID",
						ArgsUsage: "[ID of group]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							groupId := ctx.Args().First()
							network := ctx.String("network")
							if groupId == "" {
								return errors.New("no ID provided")
							}
							err := groups.DeleteGroup(groupId, network)
							return err
						},
					},
					{
						Name:      "get",
						Aliases:   []string{"g"},
						Usage:     "Get group info by ID",
						ArgsUsage: "[ID of group]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							groupId := ctx.Args().First()
							network := ctx.String("network")
							if groupId == "" {
								return errors.New("no ID provided")
							}
							_, err := groups.GetGroup(groupId, network)
							return err
						},
					},
					{
						Name:      "add",
						Aliases:   []string{"a"},
						Usage:     "Add a file to a group",
						ArgsUsage: "[group id] [file id]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							groupId := ctx.Args().First()
							fileId := ctx.Args().Get(1)
							network := ctx.String("network")
							if groupId == "" {
								return errors.New("no group id provided")
							}
							if fileId == "" {
								return errors.New("no file id provided")
							}
							err := groups.AddFile(groupId, fileId, network)
							return err
						},
					},
					{
						Name:      "remove",
						Aliases:   []string{"r"},
						Usage:     "Remove a file from a group",
						ArgsUsage: "[group id] [file id]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							groupId := ctx.Args().First()
							fileId := ctx.Args().Get(1)
							network := ctx.String("network")
							if groupId == "" {
								return errors.New("no group id provided")
							}
							if fileId == "" {
								return errors.New("no file id provided")
							}
							err := groups.RemoveFile(groupId, fileId, network)
							return err
						},
					},
				},
			},
			{
				Name:    "files",
				Aliases: []string{"f"},
				Usage:   "Interact with your files on Pinata",
				Subcommands: []*cli.Command{
					{
						Name:      "delete",
						Aliases:   []string{"d"},
						Usage:     "Delete a file by ID",
						ArgsUsage: "[ID of file]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							fileId := ctx.Args().First()
							network := ctx.String("network")
							if fileId == "" {
								return errors.New("no file ID provided")
							}
							err := files.DeleteFile(fileId, network)
							return err
						},
					},
					{
						Name:      "get",
						Aliases:   []string{"g"},
						Usage:     "Get file info by ID",
						ArgsUsage: "[ID of file]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							fileId := ctx.Args().First()
							network := ctx.String("network")
							if fileId == "" {
								return errors.New("no CID provided")
							}
							_, err := files.GetFile(fileId, network)
							return err
						},
					},
					{
						Name:      "update",
						Aliases:   []string{"u"},
						Usage:     "Update a file by ID",
						ArgsUsage: "[ID of file]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Update the name of a file",
							},
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							fileId := ctx.Args().First()
							name := ctx.String("name")
							network := ctx.String("network")
							if fileId == "" {
								return errors.New("no ID provided")
							}
							_, err := files.UpdateFile(fileId, name, network)
							return err
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "List most recent files",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Filter by name of the target file",
							},
							&cli.StringFlag{
								Name:    "cid",
								Aliases: []string{"c"},
								Usage:   "Filter results by CID",
							},
							&cli.StringFlag{
								Name:    "group",
								Aliases: []string{"g"},
								Usage:   "Filter results by group ID",
							},
							&cli.StringFlag{
								Name:    "mime",
								Aliases: []string{"m"},
								Usage:   "Filter results by file mime type",
							},
							&cli.StringFlag{
								Name:    "amount",
								Aliases: []string{"a"},
								Usage:   "The number of files you would like to return",
							},
							&cli.StringFlag{
								Name:    "token",
								Aliases: []string{"t"},
								Usage:   "Paginate through file results using the pageToken",
							},
							&cli.BoolFlag{
								Name:  "cidPending",
								Value: false,
								Usage: "Filter results based on whether or not the CID is pending",
							},
							&cli.StringSliceFlag{
								Name:    "keyvalues",
								Aliases: []string{"kv"},
								Usage:   "Filter results by metadata keyvalues (format: key=value)",
							},
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							amount := ctx.String("amount")
							token := ctx.String("token")
							name := ctx.String("name")
							cid := ctx.String("cid")
							group := ctx.String("group")
							mime := ctx.String("mime")
							cidPending := ctx.Bool("cidPending")
							keyvaluesSlice := ctx.StringSlice("keyvalues")
							keyvalues := make(map[string]string)
							network := ctx.String("network")
							for _, kv := range keyvaluesSlice {
								parts := strings.SplitN(kv, "=", 2)
								if len(parts) == 2 {
									keyvalues[parts[0]] = parts[1]
								}
							}
							_, err := files.ListFiles(amount, token, cidPending, name, cid, group, mime, keyvalues, network)
							return err
						},
					},
				},
			},
			{
				Name:    "swaps",
				Aliases: []string{"s"},
				Usage:   "Interact and manage hot swaps on Pinata",
				Subcommands: []*cli.Command{
					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List swaps for a given gateway domain or for your config gateway domain",
						ArgsUsage: "[cid] [optional gateway domain]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							cid := ctx.Args().First()
							domain := ctx.Args().Get(1)
							network := ctx.String("network")
							if cid == "" {
								return errors.New("No CID provided")
							}
							_, err := files.GetSwapHistory(cid, domain, network)
							return err
						},
					},
					{
						Name:      "add",
						Aliases:   []string{"a"},
						Usage:     "Add a swap for a CID",
						ArgsUsage: "[cid] [swap cid]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							cid := ctx.Args().First()
							swapCid := ctx.Args().Get(1)
							network := ctx.String("network")
							if cid == "" {
								return errors.New("No CID provided")
							}
							if swapCid == "" {
								return errors.New("No swap CID provided")
							}
							_, err := files.AddSwap(cid, swapCid, network)
							return err
						},
					},
					{
						Name:      "delete",
						Aliases:   []string{"d"},
						Usage:     "Remeove a swap for a CID",
						ArgsUsage: "[cid]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							cid := ctx.Args().First()
							network := ctx.String("network")
							if cid == "" {
								return errors.New("No CID provided")
							}
							err := files.RemoveSwap(cid, network)
							return err
						},
					},
				},
			},
			{
				Name:    "gateways",
				Aliases: []string{"gw"},
				Usage:   "Interact with your gateways on Pinata",
				Subcommands: []*cli.Command{
					{
						Name:      "set",
						Aliases:   []string{"s"},
						Usage:     "Set your default gateway to be used by the CLI",
						ArgsUsage: "[domain of the gateway]",
						Action: func(ctx *cli.Context) error {
							domain := ctx.Args().First()
							err := gateways.SetGateway(domain)
							return err
						},
					},
					{
						Name:      "open",
						Aliases:   []string{"o"},
						Usage:     "Open a file in the browser",
						ArgsUsage: "[CID of the file]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							cid := ctx.Args().First()
							network := ctx.String("network")
							if cid == "" {
								return errors.New("No CID provided")
							}
							err := gateways.OpenCID(cid, network)
							return err
						},
					},
					{
						Name:      "link",
						Aliases:   []string{"l"},
						Usage:     "Get either an IPFS link for a public file or a temporary access link for a Private IPFS file",
						ArgsUsage: "[cid of the file, seconds the url is valid for]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "network",
								Aliases: []string{"net"},
								Usage:   "Specify the network (public or private). Uses default if not specified",
							},
						},
						Action: func(ctx *cli.Context) error {
							network := ctx.String("network")
							cid := ctx.Args().First()
							if cid == "" {
								return errors.New("No CID provided")
							}
							expires := ctx.Args().Get(1)

							if expires == "" {
								expires = "30"
							}

							expiresInt, err := strconv.Atoi(expires)
							if err != nil {
								return errors.New("Invalid expire time")
							}
							_, err = gateways.GetAccessLink(cid, expiresInt, network)
							return err
						},
					},
				},
			},
			{
				Name:    "keys",
				Aliases: []string{"k"},
				Usage:   "Create and manage generated API keys",
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "Create an API key with admin or scoped permissions",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "Name of the API key",
								Required: true,
							},
							&cli.BoolFlag{
								Name:    "admin",
								Aliases: []string{"a"},
								Usage:   "Set the key as Admin",
								Value:   false,
							},
							&cli.IntFlag{
								Name:    "uses",
								Aliases: []string{"u"},
								Usage:   "Max uses a key can use",
							},
							&cli.StringSliceFlag{
								Name:    "endpoints",
								Aliases: []string{"e"},
								Usage:   "Optional array of endpoints the key is allowed to use",
							},
						},
						Action: func(ctx *cli.Context) error {
							name := ctx.String("name")
							admin := ctx.Bool("admin")
							uses := ctx.Int("uses")
							endpoints := ctx.StringSlice("endpoints")
							_, err := keys.CreateKey(name, admin, uses, endpoints)
							return err
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "List and filter API key",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Name of the API key",
							},
							&cli.BoolFlag{
								Name:    "revoked",
								Aliases: []string{"r"},
								Usage:   "Set the key as Admin",
							},
							&cli.BoolFlag{
								Name:    "exhausted",
								Aliases: []string{"e"},
								Usage:   "Filter keys that are exhausted or not",
							},
							&cli.BoolFlag{
								Name:    "uses",
								Aliases: []string{"u"},
								Usage:   "Filter keys that do or don't have limited uses",
							},
							&cli.StringFlag{
								Name:    "offset",
								Aliases: []string{"o"},
								Usage:   "Offset the number of results to paginate",
							},
						},
						Action: func(ctx *cli.Context) error {
							name := ctx.String("name")
							offset := ctx.String("offset")
							revoked := ctx.Bool("revoked")
							uses := ctx.Bool("uses")
							exhausted := ctx.Bool("exhausted")
							_, err := keys.ListKeys(name, revoked, uses, exhausted, offset)
							return err
						},
					},
					{
						Name:      "revoke",
						Aliases:   []string{"r"},
						Usage:     "Revoke an API key",
						ArgsUsage: "[key]",
						Action: func(ctx *cli.Context) error {
							key := ctx.Args().First()
							if key == "" {
								return errors.New("No key provided")
							}
							err := keys.RevokeKey(key)
							return err
						},
					},
				},
			},
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "Configure Pinata CLI settings",
				Subcommands: []*cli.Command{
					{
						Name:      "network",
						Aliases:   []string{"net"},
						Usage:     "Set default network (public or private)",
						ArgsUsage: "[network]",
						Action: func(ctx *cli.Context) error {
							network := ctx.Args().First()
							if network == "" {
								// If no parameter, show current setting
								current, err := config.GetDefaultNetwork()
								if err != nil {
									return err
								}
								fmt.Printf("Current default network: %s\n", current)
								return nil
							}
							return config.SetDefaultNetwork(network)
						},
					},
				},
			},
			{
				Name:    "agents",
			Aliases: []string{"ag"},
			Usage:   "Interact with AI agents on Pinata",
			Subcommands: []*cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "List all agents",
					Action: func(ctx *cli.Context) error {
						_, err := agents.ListAgents()
						return err
					},
				},
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new agent",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "name",
							Aliases:  []string{"n"},
							Usage:    "Name of the agent (required)",
							Required: true,
						},
						&cli.StringFlag{
							Name:    "description",
							Aliases: []string{"d"},
							Usage:   "Agent personality description",
						},
						&cli.StringFlag{
							Name:  "vibe",
							Usage: "Agent vibe/tagline",
						},
						&cli.StringFlag{
							Name:  "emoji",
							Usage: "Agent emoji",
						},
						&cli.StringSliceFlag{
							Name:  "skill",
							Usage: "Skill CIDs to attach (can be specified multiple times)",
						},
						&cli.StringSliceFlag{
							Name:  "secret",
							Usage: "Secret IDs to attach (can be specified multiple times)",
						},
					},
					Action: func(ctx *cli.Context) error {
						name := ctx.String("name")
						description := ctx.String("description")
						vibe := ctx.String("vibe")
						emoji := ctx.String("emoji")
						skills := ctx.StringSlice("skill")
						secrets := ctx.StringSlice("secret")
						_, err := agents.CreateAgent(name, description, vibe, emoji, skills, secrets)
						return err
					},
				},
				{
					Name:      "get",
					Aliases:   []string{"g"},
					Usage:     "Get agent details",
					ArgsUsage: "[agent ID]",
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						_, err := agents.GetAgent(agentID)
						return err
					},
				},
				{
					Name:      "delete",
					Aliases:   []string{"d"},
					Usage:     "Delete an agent",
					ArgsUsage: "[agent ID]",
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						return agents.DeleteAgent(agentID)
					},
				},
				{
					Name:      "restart",
					Aliases:   []string{"r"},
					Usage:     "Restart an agent",
					ArgsUsage: "[agent ID]",
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						_, err := agents.RestartAgent(agentID)
						return err
					},
				},
				{
					Name:      "logs",
					Usage:     "Get agent logs",
					ArgsUsage: "[agent ID]",
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						_, err := agents.GetAgentLogs(agentID)
						return err
					},
				},
				{
					Name:      "chat",
					Aliases:   []string{"c"},
					Usage:     "Interactive chat with an agent",
					ArgsUsage: "[agent ID] [optional prompt]",
					Description: `Start an interactive chat session with an agent.

The gateway URL and token are automatically fetched from the agent's configuration.

Output modes:
  - TTY stdout:     Interactive TUI with markdown rendering
  - Non-TTY stdout: JSONL streaming (machine-readable, default for pipes)
  - --text:         Plain text streaming (simpler alternative to JSONL)
  - --conversation: Multi-turn mode (read messages from stdin line-by-line)

Examples:
  # Interactive TUI mode
  pinata agents chat <agent-id>

  # Single message with plain text response (for agents/scripts)
  echo "Hello" | pinata agents chat <agent-id> --text

  # JSONL output (machine-readable, default when piped)
  echo "Hello" | pinata agents chat <agent-id>

  # Multi-turn conversation (each line is a message)
  echo -e "Hello\nHow are you?" | pinata agents chat <id> -C --text

  # Interactive conversation from a file
  pinata agents chat <id> --conversation --text < messages.txt

  # Filter JSONL with jq
  echo "hi" | pinata agents chat <id> | jq -c 'select(.type=="content_delta")'`,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "model",
							Usage: "Model override",
						},
						&cli.BoolFlag{
							Name:  "json",
							Usage: "Force JSONL output (auto-enabled when stdout is not a TTY)",
						},
						&cli.BoolFlag{
							Name:  "text",
							Usage: "Force plain text output (simpler alternative to JSONL for pipes)",
						},
						&cli.BoolFlag{
							Name:    "conversation",
							Aliases: []string{"C"},
							Usage:   "Multi-turn conversation mode (read messages from stdin line-by-line)",
						},
						&cli.StringFlag{
							Name:  "session",
							Usage: "Session key for conversation context (default: agent:main:cli)",
						},
						&cli.BoolFlag{
							Name:    "yes",
							Aliases: []string{"y"},
							Usage:   "Auto-approve tool calls (default: true, tools run server-side)",
							Hidden:  true,
							Value:   true,
						},
						&cli.StringFlag{
							Name:   "gateway",
							Usage:  "Override gateway URL (auto-detected from agent)",
							Hidden: true,
						},
						&cli.StringFlag{
							Name:   "token",
							Usage:  "Override API token (auto-detected from agent)",
							Hidden: true,
						},
					},
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						gatewayURL := ctx.String("gateway")
						token := ctx.String("token")
						model := ctx.String("model")
						jsonOutput := ctx.Bool("json")
						textOutput := ctx.Bool("text")
						conversationMode := ctx.Bool("conversation")
						autoApprove := ctx.Bool("yes")
						session := ctx.String("session")

						// Get optional prompt from remaining args
						prompt := ""
						if ctx.Args().Len() > 1 {
							prompt = strings.Join(ctx.Args().Slice()[1:], " ")
						}

						return chat.StartChat(agentID, gatewayURL, token, model, jsonOutput, textOutput, conversationMode, autoApprove, prompt, session)
					},
				},
				{
					Name:      "exec",
					Usage:     "Execute a command in an agent container",
					ArgsUsage: "[agent ID] [command]",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "cwd",
							Usage: "Working directory for the command",
						},
					},
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						command := ctx.Args().Get(1)
						cwd := ctx.String("cwd")
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						if command == "" {
							return errors.New("no command provided")
						}
						_, err := agents.ExecCommand(agentID, command, cwd)
						return err
					},
				},
				{
					Name:    "skills",
					Aliases: []string{"sk"},
					Usage:   "Manage agent skills",
					Subcommands: []*cli.Command{
						{
							Name:    "list",
							Aliases: []string{"l"},
							Usage:   "List available skills in library",
							Action: func(ctx *cli.Context) error {
								_, err := agents.ListSkills()
								return err
							},
						},
						{
							Name:    "create",
							Aliases: []string{"c"},
							Usage:   "Create a new skill",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "cid",
									Usage:    "Content ID of the skill (required)",
									Required: true,
								},
								&cli.StringFlag{
									Name:     "name",
									Aliases:  []string{"n"},
									Usage:    "Skill name (required)",
									Required: true,
								},
								&cli.StringFlag{
									Name:    "description",
									Aliases: []string{"d"},
									Usage:   "Skill description",
								},
								&cli.StringSliceFlag{
									Name:  "env",
									Usage: "Required environment variable names",
								},
								&cli.StringFlag{
									Name:  "file-id",
									Usage: "Pinata v3 file ID",
								},
							},
							Action: func(ctx *cli.Context) error {
								cid := ctx.String("cid")
								name := ctx.String("name")
								description := ctx.String("description")
								envVars := ctx.StringSlice("env")
								fileId := ctx.String("file-id")
								_, err := agents.CreateSkill(cid, name, description, envVars, fileId)
								return err
							},
						},
						{
							Name:      "delete",
							Aliases:   []string{"d"},
							Usage:     "Delete a skill from library",
							ArgsUsage: "[skill CID]",
							Action: func(ctx *cli.Context) error {
								skillCid := ctx.Args().First()
								if skillCid == "" {
									return errors.New("no skill CID provided")
								}
								return agents.DeleteSkill(skillCid)
							},
						},
						{
							Name:      "attach",
							Aliases:   []string{"a"},
							Usage:     "Attach skills to an agent",
							ArgsUsage: "[agent ID] [skill CID...]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								skillCids := ctx.Args().Tail()
								if len(skillCids) == 0 {
									return errors.New("no skill CIDs provided")
								}
								return agents.AttachSkills(agentID, skillCids)
							},
						},
						{
							Name:      "detach",
							Usage:     "Detach a skill from an agent",
							ArgsUsage: "[agent ID] [skill ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								skillID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if skillID == "" {
									return errors.New("no skill ID provided")
								}
								return agents.DetachSkill(agentID, skillID)
							},
						},
					},
				},
				{
					Name:    "secrets",
					Aliases: []string{"sec"},
					Usage:   "Manage secrets",
					Subcommands: []*cli.Command{
						{
							Name:    "list",
							Aliases: []string{"l"},
							Usage:   "List all secrets",
							Action: func(ctx *cli.Context) error {
								_, err := agents.ListSecrets()
								return err
							},
						},
						{
							Name:    "create",
							Aliases: []string{"c"},
							Usage:   "Create a new secret",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "name",
									Aliases:  []string{"n"},
									Usage:    "Secret name (e.g. ANTHROPIC_API_KEY)",
									Required: true,
								},
								&cli.StringFlag{
									Name:     "value",
									Aliases:  []string{"v"},
									Usage:    "Secret value",
									Required: true,
								},
							},
							Action: func(ctx *cli.Context) error {
								name := ctx.String("name")
								value := ctx.String("value")
								_, err := agents.CreateSecret(name, value)
								return err
							},
						},
						{
							Name:      "update",
							Aliases:   []string{"u"},
							Usage:     "Update a secret value",
							ArgsUsage: "[secret ID]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "value",
									Aliases:  []string{"v"},
									Usage:    "New secret value",
									Required: true,
								},
							},
							Action: func(ctx *cli.Context) error {
								secretID := ctx.Args().First()
								value := ctx.String("value")
								if secretID == "" {
									return errors.New("no secret ID provided")
								}
								return agents.UpdateSecret(secretID, value)
							},
						},
						{
							Name:      "delete",
							Aliases:   []string{"d"},
							Usage:     "Delete a secret",
							ArgsUsage: "[secret ID]",
							Action: func(ctx *cli.Context) error {
								secretID := ctx.Args().First()
								if secretID == "" {
									return errors.New("no secret ID provided")
								}
								return agents.DeleteSecret(secretID)
							},
						},
						{
							Name:      "attach",
							Aliases:   []string{"a"},
							Usage:     "Attach secrets to an agent",
							ArgsUsage: "[agent ID] [secret ID...]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								secretIds := ctx.Args().Tail()
								if len(secretIds) == 0 {
									return errors.New("no secret IDs provided")
								}
								return agents.AttachSecrets(agentID, secretIds)
							},
						},
						{
							Name:      "detach",
							Usage:     "Detach a secret from an agent",
							ArgsUsage: "[agent ID] [secret ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								secretID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if secretID == "" {
									return errors.New("no secret ID provided")
								}
								return agents.DetachSecret(agentID, secretID)
							},
						},
					},
				},
				{
					Name:    "channels",
					Aliases: []string{"ch"},
					Usage:   "Manage agent channels",
					Subcommands: []*cli.Command{
						{
							Name:      "status",
							Aliases:   []string{"s"},
							Usage:     "Get channel configuration status",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.GetChannelStatus(agentID)
								return err
							},
						},
						{
							Name:      "configure",
							Aliases:   []string{"c"},
							Usage:     "Configure a channel (telegram, slack, discord, whatsapp)",
							ArgsUsage: "[agent ID] [channel]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "bot-token",
									Usage: "Bot token",
								},
								&cli.StringFlag{
									Name:  "app-token",
									Usage: "App token (Slack only)",
								},
								&cli.StringFlag{
									Name:  "dm-policy",
									Usage: "DM policy: open or pairing",
								},
								&cli.StringSliceFlag{
									Name:  "allow-from",
									Usage: "Allowed user IDs/phone numbers",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								channel := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if channel == "" {
									return errors.New("no channel provided (telegram, slack, discord, whatsapp)")
								}
								botToken := ctx.String("bot-token")
								appToken := ctx.String("app-token")
								dmPolicy := ctx.String("dm-policy")
								allowFrom := ctx.StringSlice("allow-from")
								return agents.ConfigureChannel(agentID, channel, botToken, appToken, dmPolicy, allowFrom)
							},
						},
						{
							Name:      "remove",
							Aliases:   []string{"r"},
							Usage:     "Remove a channel configuration",
							ArgsUsage: "[agent ID] [channel]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								channel := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if channel == "" {
									return errors.New("no channel provided")
								}
								return agents.RemoveChannel(agentID, channel)
							},
						},
					},
				},
				{
					Name:    "devices",
					Aliases: []string{"dev"},
					Usage:   "Manage agent devices",
					Subcommands: []*cli.Command{
						{
							Name:      "list",
							Aliases:   []string{"l"},
							Usage:     "List pending and paired devices",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ListDevices(agentID)
								return err
							},
						},
						{
							Name:      "approve",
							Aliases:   []string{"a"},
							Usage:     "Approve a device pairing request",
							ArgsUsage: "[agent ID] [request ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								requestID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if requestID == "" {
									return errors.New("no request ID provided")
								}
								return agents.ApproveDevice(agentID, requestID)
							},
						},
						{
							Name:      "approve-all",
							Usage:     "Approve all pending device requests",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ApproveAllDevices(agentID)
								return err
							},
						},
					},
				},
				{
					Name:    "snapshots",
					Aliases: []string{"snap"},
					Usage:   "Manage agent snapshots",
					Subcommands: []*cli.Command{
						{
							Name:      "list",
							Aliases:   []string{"l"},
							Usage:     "List agent snapshots",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ListSnapshots(agentID)
								return err
							},
						},
						{
							Name:      "create",
							Aliases:   []string{"c"},
							Usage:     "Create a snapshot",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.CreateSnapshot(agentID)
								return err
							},
						},
						{
							Name:      "status",
							Aliases:   []string{"s"},
							Usage:     "Get sync status",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.GetSyncStatus(agentID)
								return err
							},
						},
						{
							Name:      "reset",
							Aliases:   []string{"r"},
							Usage:     "Reset to a snapshot",
							ArgsUsage: "[agent ID] [snapshot CID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								snapshotCid := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if snapshotCid == "" {
									return errors.New("no snapshot CID provided")
								}
								_, err := agents.ResetSnapshot(agentID, snapshotCid)
								return err
							},
						},
					},
				},
				{
					Name:    "tasks",
					Aliases: []string{"t"},
					Usage:   "Manage agent cron jobs/tasks",
					Subcommands: []*cli.Command{
						{
							Name:      "list",
							Aliases:   []string{"l"},
							Usage:     "List tasks",
							ArgsUsage: "[agent ID]",
							Flags: []cli.Flag{
								&cli.BoolFlag{
									Name:  "include-disabled",
									Usage: "Include disabled tasks",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								includeDisabled := ctx.Bool("include-disabled")
								_, err := agents.ListTasks(agentID, includeDisabled)
								return err
							},
						},
						{
							Name:      "delete",
							Aliases:   []string{"d"},
							Usage:     "Delete a task",
							ArgsUsage: "[agent ID] [job ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								jobID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if jobID == "" {
									return errors.New("no job ID provided")
								}
								return agents.DeleteTask(agentID, jobID)
							},
						},
						{
							Name:      "toggle",
							Usage:     "Enable or disable a task",
							ArgsUsage: "[agent ID] [job ID]",
							Flags: []cli.Flag{
								&cli.BoolFlag{
									Name:  "enable",
									Usage: "Enable the task",
								},
								&cli.BoolFlag{
									Name:  "disable",
									Usage: "Disable the task",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								jobID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if jobID == "" {
									return errors.New("no job ID provided")
								}
								enable := ctx.Bool("enable")
								disable := ctx.Bool("disable")
								if enable == disable {
									return errors.New("specify either --enable or --disable")
								}
								return agents.ToggleTask(agentID, jobID, enable)
							},
						},
						{
							Name:      "run",
							Usage:     "Run a task immediately",
							ArgsUsage: "[agent ID] [job ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								jobID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if jobID == "" {
									return errors.New("no job ID provided")
								}
								return agents.RunTask(agentID, jobID)
							},
						},
						{
							Name:      "history",
							Usage:     "View task run history",
							ArgsUsage: "[agent ID] [job ID]",
							Flags: []cli.Flag{
								&cli.IntFlag{
									Name:  "limit",
									Usage: "Number of runs to return",
									Value: 10,
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								jobID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if jobID == "" {
									return errors.New("no job ID provided")
								}
								limit := ctx.Int("limit")
								_, err := agents.GetTaskHistory(agentID, jobID, limit)
								return err
							},
						},
					},
				},
				{
					Name:    "ports",
					Aliases: []string{"p"},
					Usage:   "Manage agent port forwarding",
					Subcommands: []*cli.Command{
						{
							Name:      "list",
							Aliases:   []string{"l"},
							Usage:     "List port forwarding rules",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ListPorts(agentID)
								return err
							},
						},
					},
				},
				{
					Name:  "files",
					Usage: "Agent file operations",
					Subcommands: []*cli.Command{
						{
							Name:      "read",
							Aliases:   []string{"r"},
							Usage:     "Read a file from agent container",
							ArgsUsage: "[agent ID] [file path]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								filePath := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if filePath == "" {
									return errors.New("no file path provided")
								}
								_, err := agents.ReadFile(agentID, filePath)
								return err
							},
						},
					},
				},
				{
					Name:      "feedback",
					Usage:     "Submit feedback or feature request",
					ArgsUsage: "[message]",
					Action: func(ctx *cli.Context) error {
						message := strings.Join(ctx.Args().Slice(), " ")
						if message == "" {
							return errors.New("no feedback message provided")
						}
						return agents.SubmitFeedback(message)
					},
				},
			},
		},
	},
}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
