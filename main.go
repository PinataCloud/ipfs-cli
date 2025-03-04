package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
