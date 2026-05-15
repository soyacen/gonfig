package file

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/soyacen/gonfig/format/json"
	_ "github.com/soyacen/gonfig/format/yaml"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		expectErr string
	}{
		{"Valid YAML File", "test.yaml", ""},
		{"Valid JSON File", "test.json", ""},
		{"Empty Extension", "test", "config: file extension is empty"},
		{"Unsupported Extension", "test.txt", "config: not found formatter for txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(tt.filename)
			if tt.expectErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectErr) {
					t.Errorf("expected error %q; got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if resource == nil {
				t.Errorf("expected non-nil Resource")
			} else {
				if resource.filename != tt.filename {
					t.Errorf("expected filename %q; got %q", tt.filename, resource.filename)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	content := `
key:
  nested_key: value
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	structData, err := resource.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}

	expectedStruct, _ := structpb.NewStruct(map[string]any{
		"key": map[string]any{
			"nested_key": "value",
		},
	})

	if !reflect.DeepEqual(structData, expectedStruct) {
		t.Errorf("expected %v; got %v", expectedStruct, structData)
	}
}

func TestWatch(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, time.Now().Format(time.DateTime)+"_test.yaml")
	content := `
key:
  nested_key: value
`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = resource.Load(ctx)

	c := make(chan *structpb.Struct)
	notifyFunc := func(newValue *structpb.Struct) {
		if newValue == nil {
			t.Error("expected non-nil struct from watch")
			return
		}
		c <- newValue
	}
	errFunc := func(err error) {
		t.Errorf("Error: %v", err)
	}

	stop, err := resource.Watch(ctx, notifyFunc, errFunc)
	if err != nil {
		t.Fatal(err)
	}
	defer stop(ctx)

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	newValue := <-c
	value := newValue.GetFields()["key"].GetStructValue().GetFields()["nested_key"].GetStringValue()
	if value != "value" {
		t.Errorf("expected value 'value'; got %q", value)
	}

	// 修改文件以触发 Watcher
	newContent := `
key:
  nested_key: updated_value
`
	err = os.WriteFile(testFile, []byte(newContent), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	newValue = <-c
	value = newValue.GetFields()["key"].GetStructValue().GetFields()["nested_key"].GetStringValue()
	if value != "updated_value" {
		t.Errorf("expected value 'updated_value'; got %q", value)
	}
}

func TestFactory_New(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "simple path",
			dsn:      "/path/to/config.yaml",
			wantFile: "/path/to/config.yaml",
			wantErr:  false,
		},
		{
			name:     "with scheme",
			dsn:      "file:///path/to/config.yaml",
			wantFile: "/path/to/config.yaml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Factory{}
			r, err := f.New(context.Background(), tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Factory.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if r.(*Resource).filename != tt.wantFile {
				t.Errorf("filename = %v, want %v", r.(*Resource).filename, tt.wantFile)
			}
		})
	}
}

func TestResource_New(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "default file",
			dsn:      "/path/to/config.yaml",
			wantFile: "/path/to/config.yaml",
			wantErr:  false,
		},
		{
			name:     "file scheme",
			dsn:      "file:///path/to/config.yaml",
			wantFile: "/path/to/config.yaml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Factory{}
			r, err := f.New(context.Background(), tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Factory.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if r.(*Resource).filename != tt.wantFile {
				t.Errorf("filename = %v, want %v", r.(*Resource).filename, tt.wantFile)
			}
		})
	}
}

func TestWatch_NilNotifyFunc(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	_ = os.WriteFile(testFile, []byte("key: value"), 0o644)

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = resource.Watch(ctx, nil, func(error) {})
	if err == nil {
		t.Error("expected error for nil notifyFunc")
	}
}

func TestWatch_NilErrFunc(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	_ = os.WriteFile(testFile, []byte("key: value"), 0o644)

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	stop, err := resource.Watch(ctx, func(*structpb.Struct) {}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stop == nil {
		t.Fatal("expected non-nil stop func")
	}
	_ = stop(ctx)
}

func TestWatch_CancelledContext(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	_ = os.WriteFile(testFile, []byte("key: value"), 0o644)

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = resource.Watch(ctx, func(*structpb.Struct) {}, func(error) {})
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	resource, err := New("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = resource.Load(ctx)
	if err == nil {
		t.Error("expected error when file does not exist")
	}
}

func TestLoad_ParseError(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	_ = os.WriteFile(testFile, []byte("invalid: yaml: ["), 0o644)

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = resource.Load(ctx)
	if err == nil {
		t.Error("expected error for invalid file content")
	}
}

func TestWatch_FileContentUnchanged(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	content := "key: value\n"
	_ = os.WriteFile(testFile, []byte(content), 0o644)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = resource.Load(ctx)

	notifyCount := 0
	notifyFunc := func(newValue *structpb.Struct) {
		notifyCount++
	}
	errFunc := func(err error) {
		t.Errorf("Error: %v", err)
	}

	stop, err := resource.Watch(ctx, notifyFunc, errFunc)
	if err != nil {
		t.Fatal(err)
	}
	defer stop(ctx)

	// 写入相同内容，不应触发通知
	_ = os.WriteFile(testFile, []byte(content), 0o644)

	// 等待一小段时间确保事件已处理
	time.Sleep(200 * time.Millisecond)

	if notifyCount != 0 {
		t.Errorf("expected no notifications for unchanged content, got %d", notifyCount)
	}
}

func TestWatch_IrrelevantEvent(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.yaml")
	otherFile := filepath.Join(tempDir, "other.yaml")
	_ = os.WriteFile(testFile, []byte("key: value\n"), 0o644)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := New(testFile)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = resource.Load(ctx)

	notifyCount := 0
	notifyFunc := func(newValue *structpb.Struct) {
		notifyCount++
	}
	errFunc := func(err error) {
		t.Errorf("Error: %v", err)
	}

	stop, err := resource.Watch(ctx, notifyFunc, errFunc)
	if err != nil {
		t.Fatal(err)
	}
	defer stop(ctx)

	// 创建不相关文件，不应触发通知
	_ = os.WriteFile(otherFile, []byte("other: data\n"), 0o644)

	// 等待一小段时间确保事件已处理
	time.Sleep(200 * time.Millisecond)

	if notifyCount != 0 {
		t.Errorf("expected no notifications for irrelevant file, got %d", notifyCount)
	}
}
