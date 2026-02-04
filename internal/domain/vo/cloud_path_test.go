package vo

import "testing"

func TestNewCloudPath_normalize(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty string", "", "/"},
		{"whitespace only", "   ", "/   "},
		{"root slash", "/", "/"},
		{"no leading slash", "foo/bar", "/foo/bar"},
		{"trailing slashes", "/foo/bar///", "/foo/bar"},
		{"double slashes", "/foo//bar", "/foo/bar"},
		{"dot segment", "/foo/./bar", "/foo/bar"},
		{"dotdot segment", "/foo/bar/../baz", "/foo/baz"},
		{"simple path", "/documents/photos", "/documents/photos"},
		{"single segment", "file.txt", "/file.txt"},
		{"spaces in path are preserved", "/foo bar/baz", "/foo bar/baz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCloudPath(tt.raw).String()
			if got != tt.want {
				t.Errorf("NewCloudPath(%q).String() = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCloudPath_Name(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"/foo", "foo"},
		{"/foo/bar.txt", "bar.txt"},
		{"/a/b/c", "c"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := NewCloudPath(tt.path).Name()
			if got != tt.want {
				t.Errorf("CloudPath(%q).Name() = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestCloudPath_Parent(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"/foo", "/"},
		{"/foo/bar", "/foo"},
		{"/a/b/c", "/a/b"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := NewCloudPath(tt.path).Parent().String()
			if got != tt.want {
				t.Errorf("CloudPath(%q).Parent() = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestCloudPath_IsRoot(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"", true},
		{"/foo", false},
		{"/foo/bar", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := NewCloudPath(tt.path).IsRoot()
			if got != tt.want {
				t.Errorf("CloudPath(%q).IsRoot() = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCloudPath_Join(t *testing.T) {
	tests := []struct {
		base string
		name string
		want string
	}{
		{"/", "foo", "/foo"},
		{"/foo", "bar", "/foo/bar"},
		{"/foo/bar", "baz.txt", "/foo/bar/baz.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.base+"+"+tt.name, func(t *testing.T) {
			got := NewCloudPath(tt.base).Join(tt.name).String()
			if got != tt.want {
				t.Errorf("CloudPath(%q).Join(%q) = %q, want %q", tt.base, tt.name, got, tt.want)
			}
		})
	}
}

func TestCloudPath_HasPrefix(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{"child of root", "/foo", "/", false},
		{"direct child", "/foo/bar", "/foo", true},
		{"deeper descendant", "/foo/bar/baz", "/foo", true},
		{"same path", "/foo", "/foo", false},
		{"no relation", "/bar", "/foo", false},
		{"similar prefix not boundary", "/foobar", "/foo", false},
		{"root of root", "/", "/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCloudPath(tt.path).HasPrefix(NewCloudPath(tt.prefix))
			if got != tt.want {
				t.Errorf("CloudPath(%q).HasPrefix(%q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}
