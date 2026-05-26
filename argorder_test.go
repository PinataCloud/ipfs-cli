package main

import (
	"reflect"
	"testing"

	"github.com/urfave/cli/v2"
)

func testApp() *cli.App {
	return &cli.App{
		Name: "pinata",
		Commands: []*cli.Command{
			{
				Name:    "agents",
				Aliases: []string{"ag"},
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
							&cli.StringFlag{Name: "engine"},
							&cli.StringFlag{Name: "user-name"},
						},
					},
					{
						Name:    "templates",
						Aliases: []string{"tpl"},
						Subcommands: []*cli.Command{
							{
								Name:    "validate",
								Aliases: []string{"v"},
								Flags: []cli.Flag{
									&cli.StringFlag{Name: "ref", Aliases: []string{"r", "branch", "b"}},
									&cli.StringFlag{Name: "path", Aliases: []string{"p"}},
								},
							},
							{
								Name: "update",
								Flags: []cli.Flag{
									&cli.StringFlag{Name: "git-url"},
									&cli.StringFlag{Name: "ref", Aliases: []string{"r", "branch", "b"}},
									&cli.StringFlag{Name: "path", Aliases: []string{"p"}},
								},
							},
							{Name: "search-refs"},
						},
					},
					{
						Name: "channels",
						Subcommands: []*cli.Command{
							{
								Name: "configure",
								Flags: []cli.Flag{
									&cli.StringFlag{Name: "bot-token"},
									&cli.BoolFlag{Name: "enabled"},
									&cli.BoolFlag{Name: "skip-restart"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestReorderArgs(t *testing.T) {
	app := testApp()
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "flags after positional get moved before it",
			in:   []string{"pinata", "agents", "templates", "validate", "https://x", "--ref", "main", "--path", "sub"},
			want: []string{"pinata", "agents", "templates", "validate", "--ref", "main", "--path", "sub", "https://x"},
		},
		{
			name: "flags already before positional unchanged",
			in:   []string{"pinata", "agents", "templates", "validate", "--ref", "main", "https://x"},
			want: []string{"pinata", "agents", "templates", "validate", "--ref", "main", "https://x"},
		},
		{
			name: "multiple flags before positional unchanged",
			in:   []string{"pinata", "agents", "templates", "validate", "--ref", "main", "--path", "sub", "https://x"},
			want: []string{"pinata", "agents", "templates", "validate", "--ref", "main", "--path", "sub", "https://x"},
		},
		{
			name: "command with only flags, no positional (regression: do not orphan a flag value)",
			in:   []string{"pinata", "agents", "create", "--name", "zz", "--engine", "hermes", "--user-name", "Smoke Test"},
			want: []string{"pinata", "agents", "create", "--name", "zz", "--engine", "hermes", "--user-name", "Smoke Test"},
		},
		{
			name: "alias branch after positional",
			in:   []string{"pinata", "ag", "tpl", "update", "tid123", "--branch", "develop", "--path", "pkg/a"},
			want: []string{"pinata", "ag", "tpl", "update", "--branch", "develop", "--path", "pkg/a", "tid123"},
		},
		{
			name: "bool flag does not consume following positional",
			in:   []string{"pinata", "agents", "channels", "configure", "aid", "tg", "--enabled", "--bot-token", "xyz"},
			want: []string{"pinata", "agents", "channels", "configure", "--enabled", "--bot-token", "xyz", "aid", "tg"},
		},
		{
			name: "equals form is self-contained",
			in:   []string{"pinata", "agents", "templates", "validate", "https://x", "--ref=main"},
			want: []string{"pinata", "agents", "templates", "validate", "--ref=main", "https://x"},
		},
		{
			name: "two positionals no flags",
			in:   []string{"pinata", "agents", "templates", "search-refs", "https://x", "query"},
			want: []string{"pinata", "agents", "templates", "search-refs", "https://x", "query"},
		},
		{
			name: "double dash ends flag region",
			in:   []string{"pinata", "agents", "templates", "validate", "https://x", "--", "--ref"},
			want: []string{"pinata", "agents", "templates", "validate", "https://x", "--", "--ref"},
		},
		{
			name: "no args",
			in:   []string{"pinata"},
			want: []string{"pinata"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reorderArgs(app, tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("reorderArgs()\n got=%v\nwant=%v", got, tc.want)
			}
		})
	}
}
