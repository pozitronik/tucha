package httpapi

import "testing"

func TestParseUploadHome(t *testing.T) {
	tests := []struct {
		name    string
		urlPath string
		want    string
	}{
		{"plain upload", "/upload/", ""},
		{"with home", "/upload/home=/test.txt", "/test.txt"},
		{"nested path", "/upload/home=/folder/subfolder/file.bin", "/folder/subfolder/file.bin"},
		{"with extra params", "/upload/home=/file.txt&cloud_domain=2", "/file.txt"},
		{"url encoded", "/upload/home=%2Fpath%2Fto%2Ffile.txt", "/path/to/file.txt"},
		{"no trailing slash", "/upload", ""},
		{"empty home", "/upload/home=", ""},
		{"no home param", "/upload/cloud_domain=2", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUploadHome(tt.urlPath)
			if got != tt.want {
				t.Errorf("parseUploadHome(%q) = %q, want %q", tt.urlPath, got, tt.want)
			}
		})
	}
}
