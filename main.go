package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
						&cli.StringFlag{
							Name:    "template",
							Aliases: []string{"t"},
							Usage:   "Template ID to deploy from (uses template snapshot, skills, and defaults)",
						},
					},
					Action: func(ctx *cli.Context) error {
						name := ctx.String("name")
						description := ctx.String("description")
						vibe := ctx.String("vibe")
						emoji := ctx.String("emoji")
						skills := ctx.StringSlice("skill")
						secrets := ctx.StringSlice("secret")
						template := ctx.String("template")
						_, err := agents.CreateAgent(name, description, vibe, emoji, template, skills, secrets)
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
							Name:      "create",
							Aliases:   []string{"c"},
							Usage:     "Create a new task",
							ArgsUsage: "[agent ID]",
							Description: `Create a new cron job for an agent.

Schedule types:
  --at       Run once at a specific time (ISO 8601 format)
  --every    Run at intervals (e.g., "1h", "30m", "24h")
  --cron     Run on a cron schedule (e.g., "0 9 * * *")

Payload types (choose one):
  --system-event  System event text (triggers heartbeat-style execution)
  --agent-turn    Agent turn message (triggers conversational response)

Examples:
  # Run every hour with a system event
  pinata agents tasks create <agent-id> --name "hourly-check" --every 1h --system-event "Check for updates"

  # Run daily at 9am UTC with agent turn
  pinata agents tasks create <agent-id> --name "daily-report" --cron "0 9 * * *" --agent-turn "Generate daily report"

  # Run once at a specific time
  pinata agents tasks create <agent-id> --name "one-time" --at "2026-04-01T12:00:00Z" --system-event "Do task"`,
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "name",
									Aliases:  []string{"n"},
									Usage:    "Task name (required)",
									Required: true,
								},
								&cli.StringFlag{
									Name:  "description",
									Usage: "Task description",
								},
								&cli.BoolFlag{
									Name:  "disabled",
									Usage: "Create task in disabled state",
								},
								&cli.StringFlag{
									Name:  "at",
									Usage: "Run once at this time (ISO 8601)",
								},
								&cli.StringFlag{
									Name:  "every",
									Usage: "Run every interval (e.g., 1h, 30m)",
								},
								&cli.StringFlag{
									Name:  "cron",
									Usage: "Cron expression (e.g., '0 9 * * *')",
								},
								&cli.StringFlag{
									Name:  "tz",
									Usage: "Timezone for cron schedule",
								},
								&cli.StringFlag{
									Name:  "system-event",
									Usage: "System event payload text",
								},
								&cli.StringFlag{
									Name:  "agent-turn",
									Usage: "Agent turn message",
								},
								&cli.StringFlag{
									Name:  "model",
									Usage: "Model override for agent turn",
								},
								&cli.IntFlag{
									Name:  "timeout",
									Usage: "Timeout in seconds",
								},
								&cli.StringFlag{
									Name:  "session",
									Usage: "Session target: main or isolated",
									Value: "main",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}

								// Parse schedule
								var schedule agents.TaskSchedule
								atTime := ctx.String("at")
								every := ctx.String("every")
								cronExpr := ctx.String("cron")

								scheduleCount := 0
								if atTime != "" {
									scheduleCount++
								}
								if every != "" {
									scheduleCount++
								}
								if cronExpr != "" {
									scheduleCount++
								}

								if scheduleCount == 0 {
									return errors.New("specify one of --at, --every, or --cron")
								}
								if scheduleCount > 1 {
									return errors.New("specify only one of --at, --every, or --cron")
								}

								if atTime != "" {
									schedule.Kind = agents.ScheduleKindAt
									schedule.At = atTime
								} else if every != "" {
									schedule.Kind = agents.ScheduleKindEvery
									// Parse duration string to milliseconds
									dur, err := time.ParseDuration(every)
									if err != nil {
										return fmt.Errorf("invalid duration: %w", err)
									}
									schedule.EveryMs = int(dur.Milliseconds())
								} else if cronExpr != "" {
									schedule.Kind = agents.ScheduleKindCron
									schedule.Expr = cronExpr
								}

								if tz := ctx.String("tz"); tz != "" {
									schedule.Tz = tz
								}

								// Parse payload
								var payload agents.TaskPayload
								systemEvent := ctx.String("system-event")
								agentTurn := ctx.String("agent-turn")

								if systemEvent == "" && agentTurn == "" {
									return errors.New("specify either --system-event or --agent-turn")
								}
								if systemEvent != "" && agentTurn != "" {
									return errors.New("specify only one of --system-event or --agent-turn")
								}

								if systemEvent != "" {
									payload.Kind = agents.PayloadKindSystemEvent
									payload.Text = systemEvent
								} else {
									payload.Kind = agents.PayloadKindAgentTurn
									payload.Message = agentTurn
								}

								if model := ctx.String("model"); model != "" {
									payload.Model = model
								}
								if timeout := ctx.Int("timeout"); timeout > 0 {
									payload.TimeoutSeconds = timeout
								}

								body := agents.CreateTaskBody{
									Name:        ctx.String("name"),
									Description: ctx.String("description"),
									Enabled:     !ctx.Bool("disabled"),
									Schedule:    schedule,
									Payload:     payload,
								}

								if session := ctx.String("session"); session != "" {
									if session == "main" {
										body.SessionTarget = agents.SessionTargetMain
									} else if session == "isolated" {
										body.SessionTarget = agents.SessionTargetIsolated
									}
								}

								_, err := agents.CreateTask(agentID, body)
								return err
							},
						},
						{
							Name:      "update",
							Aliases:   []string{"u"},
							Usage:     "Update an existing task",
							ArgsUsage: "[agent ID] [job ID]",
							Description: `Update an existing cron job. Only specified fields are changed.

Examples:
  # Change task name
  pinata agents tasks update <agent-id> <job-id> --name "new-name"

  # Update schedule to run every 2 hours
  pinata agents tasks update <agent-id> <job-id> --every 2h

  # Update payload message
  pinata agents tasks update <agent-id> <job-id> --system-event "New message"`,
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "name",
									Aliases: []string{"n"},
									Usage:   "New task name",
								},
								&cli.StringFlag{
									Name:  "description",
									Usage: "New task description",
								},
								&cli.StringFlag{
									Name:  "at",
									Usage: "Run once at this time (ISO 8601)",
								},
								&cli.StringFlag{
									Name:  "every",
									Usage: "Run every interval (e.g., 1h, 30m)",
								},
								&cli.StringFlag{
									Name:  "cron",
									Usage: "Cron expression (e.g., '0 9 * * *')",
								},
								&cli.StringFlag{
									Name:  "tz",
									Usage: "Timezone for cron schedule",
								},
								&cli.StringFlag{
									Name:  "system-event",
									Usage: "System event payload text",
								},
								&cli.StringFlag{
									Name:  "agent-turn",
									Usage: "Agent turn message",
								},
								&cli.StringFlag{
									Name:  "model",
									Usage: "Model override for agent turn",
								},
								&cli.IntFlag{
									Name:  "timeout",
									Usage: "Timeout in seconds",
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

								body := agents.UpdateTaskBody{}

								if name := ctx.String("name"); name != "" {
									body.Name = name
								}
								if desc := ctx.String("description"); desc != "" {
									body.Description = desc
								}

								// Parse schedule if any schedule flag is set
								atTime := ctx.String("at")
								every := ctx.String("every")
								cronExpr := ctx.String("cron")

								if atTime != "" || every != "" || cronExpr != "" {
									schedule := &agents.TaskSchedule{}

									if atTime != "" {
										schedule.Kind = agents.ScheduleKindAt
										schedule.At = atTime
									} else if every != "" {
										schedule.Kind = agents.ScheduleKindEvery
										dur, err := time.ParseDuration(every)
										if err != nil {
											return fmt.Errorf("invalid duration: %w", err)
										}
										schedule.EveryMs = int(dur.Milliseconds())
									} else if cronExpr != "" {
										schedule.Kind = agents.ScheduleKindCron
										schedule.Expr = cronExpr
									}

									if tz := ctx.String("tz"); tz != "" {
										schedule.Tz = tz
									}

									body.Schedule = schedule
								}

								// Parse payload if any payload flag is set
								systemEvent := ctx.String("system-event")
								agentTurn := ctx.String("agent-turn")

								if systemEvent != "" || agentTurn != "" {
									payload := &agents.TaskPayload{}

									if systemEvent != "" {
										payload.Kind = agents.PayloadKindSystemEvent
										payload.Text = systemEvent
									} else {
										payload.Kind = agents.PayloadKindAgentTurn
										payload.Message = agentTurn
									}

									if model := ctx.String("model"); model != "" {
										payload.Model = model
									}
									if timeout := ctx.Int("timeout"); timeout > 0 {
										payload.TimeoutSeconds = timeout
									}

									body.Payload = payload
								}

								_, err := agents.UpdateTask(agentID, jobID, body)
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
						{
							Name:      "set",
							Aliases:   []string{"s"},
							Usage:     "Set port forwarding rules",
							ArgsUsage: "[agent ID] [port:pathPrefix] [port:pathPrefix] ...",
							Description: `Replace all port forwarding rules for this agent.

Each mapping is specified as port:pathPrefix (e.g., 8080:/api).
Up to 10 rules can be configured.

Examples:
  # Forward port 8080 to /api path
  pinata agents ports set <agent-id> 8080:/api

  # Forward multiple ports
  pinata agents ports set <agent-id> 8080:/api 3000:/app

  # Clear all port forwarding rules
  pinata agents ports set <agent-id>`,
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								var mappings []agents.PortForwarding
								for _, arg := range ctx.Args().Tail() {
									parts := strings.SplitN(arg, ":", 2)
									if len(parts) != 2 {
										return fmt.Errorf("invalid port mapping: %s (expected port:pathPrefix)", arg)
									}
									port, err := strconv.Atoi(parts[0])
									if err != nil {
										return fmt.Errorf("invalid port number: %s", parts[0])
									}
									mappings = append(mappings, agents.PortForwarding{
										Port:       port,
										PathPrefix: parts[1],
									})
								}
								_, err := agents.UpdatePorts(agentID, mappings)
								return err
							},
						},
					},
				},
				{
					Name:    "domains",
					Aliases: []string{"dom"},
					Usage:   "Manage custom domains (beta)",
					Subcommands: []*cli.Command{
						{
							Name:      "list",
							Aliases:   []string{"l"},
							Usage:     "List custom domains",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ListDomains(agentID)
								return err
							},
						},
						{
							Name:      "add",
							Aliases:   []string{"a"},
							Usage:     "Register a custom domain",
							ArgsUsage: "[agent ID]",
							Description: `Register a subdomain or custom domain for this agent.

Use --subdomain for a *.apps.pinata.cloud subdomain, or --domain for your own domain.
Max 5 domains per agent. Port 18789 is reserved.

Examples:
  # Register a subdomain (myapp.apps.pinata.cloud)
  pinata agents domains add <agent-id> --subdomain myapp --port 8080

  # Register a custom domain
  pinata agents domains add <agent-id> --domain api.example.com --port 3000

  # Add authentication protection
  pinata agents domains add <agent-id> --subdomain myapp --port 8080 --protected`,
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "subdomain",
									Usage: "Subdomain name (e.g., 'myapp' for myapp.apps.pinata.cloud)",
								},
								&cli.StringFlag{
									Name:  "domain",
									Usage: "Custom domain (e.g., 'api.example.com')",
								},
								&cli.IntFlag{
									Name:     "port",
									Usage:    "Target container port",
									Required: true,
								},
								&cli.BoolFlag{
									Name:  "protected",
									Usage: "Require authentication",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								subdomain := ctx.String("subdomain")
								customDomain := ctx.String("domain")
								port := ctx.Int("port")
								protected := ctx.Bool("protected")

								if subdomain == "" && customDomain == "" {
									return errors.New("specify either --subdomain or --domain")
								}
								if subdomain != "" && customDomain != "" {
									return errors.New("specify only one of --subdomain or --domain")
								}

								_, err := agents.CreateDomain(agentID, subdomain, customDomain, port, protected)
								return err
							},
						},
						{
							Name:      "update",
							Aliases:   []string{"u"},
							Usage:     "Update a custom domain",
							ArgsUsage: "[agent ID] [domain ID]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "subdomain",
									Usage: "New subdomain name",
								},
								&cli.StringFlag{
									Name:  "domain",
									Usage: "New custom domain",
								},
								&cli.IntFlag{
									Name:  "port",
									Usage: "New target port",
								},
								&cli.BoolFlag{
									Name:  "protected",
									Usage: "Enable authentication",
								},
								&cli.BoolFlag{
									Name:  "no-protected",
									Usage: "Disable authentication",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								domainID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if domainID == "" {
									return errors.New("no domain ID provided")
								}

								subdomain := ctx.String("subdomain")
								customDomain := ctx.String("domain")

								var targetPort *int
								if ctx.IsSet("port") {
									p := ctx.Int("port")
									targetPort = &p
								}

								var protected *bool
								if ctx.IsSet("protected") {
									p := true
									protected = &p
								} else if ctx.IsSet("no-protected") {
									p := false
									protected = &p
								}

								_, err := agents.UpdateDomain(agentID, domainID, subdomain, customDomain, targetPort, protected)
								return err
							},
						},
						{
							Name:      "delete",
							Aliases:   []string{"d"},
							Usage:     "Remove a custom domain",
							ArgsUsage: "[agent ID] [domain ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								domainID := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if domainID == "" {
									return errors.New("no domain ID provided")
								}
								_, err := agents.DeleteDomain(agentID, domainID)
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
					Name:    "templates",
					Aliases: []string{"tpl"},
					Usage:   "Browse and manage agent templates",
					Subcommands: []*cli.Command{
						{
							Name:    "list",
							Aliases: []string{"l"},
							Usage:   "List available templates",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "category",
									Aliases: []string{"c"},
									Usage:   "Filter by category",
								},
								&cli.BoolFlag{
									Name:  "featured",
									Usage: "Show only featured templates",
								},
							},
							Action: func(ctx *cli.Context) error {
								category := ctx.String("category")
								featured := ctx.Bool("featured")
								_, err := agents.ListTemplates(category, featured)
								return err
							},
						},
						{
							Name:      "get",
							Aliases:   []string{"g"},
							Usage:     "Get template details by slug",
							ArgsUsage: "[slug]",
							Action: func(ctx *cli.Context) error {
								slug := ctx.Args().First()
								if slug == "" {
									return errors.New("no template slug provided")
								}
								_, err := agents.GetTemplate(slug)
								return err
							},
						},
						{
							Name:  "mine",
							Usage: "List templates you have submitted",
							Action: func(ctx *cli.Context) error {
								_, err := agents.ListTemplatesBySubmitter()
								return err
							},
						},
						{
							Name:      "validate",
							Aliases:   []string{"v"},
							Usage:     "Validate a git repo for template submission",
							ArgsUsage: "[git URL]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "branch",
									Aliases: []string{"b"},
									Usage:   "Branch to validate (default: main)",
								},
							},
							Action: func(ctx *cli.Context) error {
								gitURL := ctx.Args().First()
								if gitURL == "" {
									return errors.New("no git URL provided")
								}
								branch := ctx.String("branch")
								_, err := agents.ValidateTemplate(gitURL, branch)
								return err
							},
						},
						{
							Name:      "submit",
							Aliases:   []string{"s"},
							Usage:     "Submit a new template from a git repo",
							ArgsUsage: "[git URL]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "branch",
									Aliases: []string{"b"},
									Usage:   "Branch to submit from (default: main)",
								},
							},
							Action: func(ctx *cli.Context) error {
								gitURL := ctx.Args().First()
								if gitURL == "" {
									return errors.New("no git URL provided")
								}
								branch := ctx.String("branch")
								_, err := agents.SubmitTemplate(gitURL, branch)
								return err
							},
						},
						{
							Name:      "update",
							Aliases:   []string{"u"},
							Usage:     "Update an existing template submission (re-pull from repo)",
							ArgsUsage: "[template ID]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "git-url",
									Usage: "New git URL (optional, uses existing if omitted)",
								},
								&cli.StringFlag{
									Name:    "branch",
									Aliases: []string{"b"},
									Usage:   "Branch to pull from (default: main)",
								},
							},
							Action: func(ctx *cli.Context) error {
								templateID := ctx.Args().First()
								if templateID == "" {
									return errors.New("no template ID provided")
								}
								gitURL := ctx.String("git-url")
								branch := ctx.String("branch")
								_, err := agents.UpdateTemplate(templateID, gitURL, branch)
								return err
							},
						},
						{
							Name:      "delete",
							Aliases:   []string{"d"},
							Usage:     "Archive a template submission",
							ArgsUsage: "[template ID]",
							Action: func(ctx *cli.Context) error {
								templateID := ctx.Args().First()
								if templateID == "" {
									return errors.New("no template ID provided")
								}
								_, err := agents.DeleteTemplate(templateID)
								return err
							},
						},
						{
							Name:      "branches",
							Usage:     "List branches for a git repository",
							ArgsUsage: "[git URL]",
							Action: func(ctx *cli.Context) error {
								gitURL := ctx.Args().First()
								if gitURL == "" {
									return errors.New("no git URL provided")
								}
								_, err := agents.ListBranches(gitURL)
								return err
							},
						},
					},
				},
				{
					Name:    "clawhub",
					Aliases: []string{"hub"},
					Usage:   "Browse and install skills from ClawHub",
					Subcommands: []*cli.Command{
						{
							Name:    "list",
							Aliases: []string{"l"},
							Usage:   "Browse ClawHub skills",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "category",
									Aliases: []string{"c"},
									Usage:   "Filter by category",
								},
								&cli.StringFlag{
									Name:    "sort",
									Aliases: []string{"s"},
									Usage:   "Sort by: popular, newest, name",
								},
								&cli.BoolFlag{
									Name:  "featured",
									Usage: "Show only featured skills",
								},
								&cli.StringFlag{
									Name:  "cursor",
									Usage: "Pagination cursor",
								},
							},
							Action: func(ctx *cli.Context) error {
								category := ctx.String("category")
								sort := ctx.String("sort")
								featured := ctx.Bool("featured")
								cursor := ctx.String("cursor")
								_, err := agents.ListHubSkills(category, sort, featured, cursor)
								return err
							},
						},
						{
							Name:      "get",
							Aliases:   []string{"g"},
							Usage:     "Get ClawHub skill details by slug",
							ArgsUsage: "[slug]",
							Action: func(ctx *cli.Context) error {
								slug := ctx.Args().First()
								if slug == "" {
									return errors.New("no skill slug provided")
								}
								_, err := agents.GetHubSkill(slug)
								return err
							},
						},
						{
							Name:      "install",
							Aliases:   []string{"i"},
							Usage:     "Install a ClawHub skill to your library",
							ArgsUsage: "[hub-skill-id]",
							Action: func(ctx *cli.Context) error {
								hubSkillID := ctx.Args().First()
								if hubSkillID == "" {
									return errors.New("no hub skill ID provided")
								}
								_, err := agents.InstallHubSkill(hubSkillID)
								return err
							},
						},
					},
				},
				{
					Name:    "config",
					Aliases: []string{"cfg"},
					Usage:   "Manage agent configuration",
					Subcommands: []*cli.Command{
						{
							Name:      "get",
							Aliases:   []string{"g"},
							Usage:     "Get agent openclaw config",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.GetConfig(agentID)
								return err
							},
						},
						{
							Name:      "set",
							Aliases:   []string{"s"},
							Usage:     "Set agent openclaw config (JSON)",
							ArgsUsage: "[agent ID] [json config]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								configJSON := ctx.Args().Get(1)
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								if configJSON == "" {
									return errors.New("no config JSON provided")
								}
								var configData interface{}
								if err := json.Unmarshal([]byte(configJSON), &configData); err != nil {
									return fmt.Errorf("invalid JSON: %w", err)
								}
								return agents.SetConfig(agentID, configData)
							},
						},
						{
							Name:      "validate",
							Aliases:   []string{"v"},
							Usage:     "Validate agent openclaw config",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.ValidateConfig(agentID)
								return err
							},
						},
					},
				},
				{
					Name:    "update",
					Aliases: []string{"up"},
					Usage:   "Manage agent openclaw updates",
					Subcommands: []*cli.Command{
						{
							Name:      "check",
							Aliases:   []string{"c"},
							Usage:     "Check for openclaw updates",
							ArgsUsage: "[agent ID]",
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								_, err := agents.CheckUpdate(agentID)
								return err
							},
						},
						{
							Name:      "apply",
							Aliases:   []string{"a"},
							Usage:     "Apply openclaw update",
							ArgsUsage: "[agent ID]",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "tag",
									Usage: "Specific version or tag to install (e.g., 'latest', '0.3.0', 'beta')",
								},
							},
							Action: func(ctx *cli.Context) error {
								agentID := ctx.Args().First()
								if agentID == "" {
									return errors.New("no agent ID provided")
								}
								tag := ctx.String("tag")
								_, err := agents.ApplyUpdate(agentID, tag)
								return err
							},
						},
					},
				},
				{
					Name:      "versions",
					Aliases:   []string{"ver"},
					Usage:     "List available agent versions",
					ArgsUsage: "[agent ID]",
					Action: func(ctx *cli.Context) error {
						agentID := ctx.Args().First()
						if agentID == "" {
							return errors.New("no agent ID provided")
						}
						_, err := agents.GetAvailableVersions(agentID)
						return err
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
