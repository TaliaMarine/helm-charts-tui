package helm

import "testing"

func TestParseRepoList(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    int
		wantErr bool
	}{
		{
			name: "two repos",
			json: `[{"name":"stable","url":"https://charts.helm.sh/stable"},{"name":"bitnami","url":"https://charts.bitnami.com/bitnami"}]`,
			want: 2,
		},
		{
			name: "empty list",
			json: `[]`,
			want: 0,
		},
		{
			name:    "invalid json",
			json:    `not json`,
			wantErr: true,
		},
		{
			name:    "null",
			json:    `null`,
			want:    0,
			wantErr: false,
		},
		{
			name: "single repo",
			json: `[{"name":"myrepo","url":"https://example.com/charts"}]`,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, err := ParseRepoList([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRepoList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(repos) != tt.want {
				t.Fatalf("ParseRepoList() got %d repos, want %d", len(repos), tt.want)
			}
		})
	}
}

func TestParseRepoListFields(t *testing.T) {
	data := `[{"name":"bitnami","url":"https://charts.bitnami.com/bitnami"}]`
	repos, err := ParseRepoList([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if repos[0].Name != "bitnami" {
		t.Errorf("Name = %q, want %q", repos[0].Name, "bitnami")
	}
	if repos[0].URL != "https://charts.bitnami.com/bitnami" {
		t.Errorf("URL = %q, want %q", repos[0].URL, "https://charts.bitnami.com/bitnami")
	}
}

func TestParseChartList(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    int
		wantErr bool
	}{
		{
			name: "two charts",
			json: `[{"name":"stable/nginx","version":"1.0.0","app_version":"1.19","description":"A web server"},{"name":"stable/redis","version":"2.0.0","app_version":"6.2","description":"A KV store"}]`,
			want: 2,
		},
		{
			name: "empty list",
			json: `[]`,
			want: 0,
		},
		{
			name:    "invalid json",
			json:    `{bad`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charts, err := ParseChartList([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseChartList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(charts) != tt.want {
				t.Fatalf("ParseChartList() got %d charts, want %d", len(charts), tt.want)
			}
		})
	}
}

func TestParseChartListFields(t *testing.T) {
	data := `[{"name":"stable/nginx","version":"1.2.3","app_version":"1.19.0","description":"Nginx web server"}]`
	charts, err := ParseChartList([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	c := charts[0]
	if c.Name != "stable/nginx" {
		t.Errorf("Name = %q, want %q", c.Name, "stable/nginx")
	}
	if c.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", c.Version, "1.2.3")
	}
	if c.AppVersion != "1.19.0" {
		t.Errorf("AppVersion = %q, want %q", c.AppVersion, "1.19.0")
	}
	if c.Description != "Nginx web server" {
		t.Errorf("Description = %q, want %q", c.Description, "Nginx web server")
	}
}
