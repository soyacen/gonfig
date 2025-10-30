package config

import (
	"context"
	"errors"
	"testing"

	"github.com/go-leo/gonfig/test"
	"google.golang.org/protobuf/types/known/structpb"
)

// 模拟resource.Resource接口
type mockLoadResource struct {
	value *structpb.Struct
	err   error
}

func (m *mockLoadResource) Load(ctx context.Context) (*structpb.Struct, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return m.value, m.err
	}
}

func (m *mockLoadResource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
	return nil, nil
}

func TestLoad(t *testing.T) {
	// 测试用例1: 单个资源加载成功
	t.Run("SingleResourceSuccess", func(t *testing.T) {
		testStruct, _ := structpb.NewStruct(map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		})
		res := &mockLoadResource{
			value: testStruct,
		}
		ctx := context.Background()
		result, err := Load[*test.Config](ctx, res)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Field1 != "value1" && result.Field2 != "value2" {
			t.Errorf("Expected to be 'value1' and 'value2', got '%s' and '%s'", result.Field1, result.Field2)
		}
	})

	// 测试用例2: 多个资源加载成功
	t.Run("MultipleResourcesSuccess", func(t *testing.T) {
		testStruct1, _ := structpb.NewStruct(map[string]interface{}{
			"field1": "value1",
		})
		testStruct2, _ := structpb.NewStruct(map[string]interface{}{
			"field2": "value2",
		})
		res1 := &mockLoadResource{value: testStruct1}
		res2 := &mockLoadResource{value: testStruct2}

		ctx := context.Background()
		result, err := Load[*test.Config](ctx, res1, res2)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Field1 != "value1" && result.Field2 != "value2" {
			t.Errorf("Expected to be 'value1' and 'value2', got '%s' and '%s'", result.Field1, result.Field2)
		}
	})

	// 测试用例3: 资源加载失败
	t.Run("ResourceLoadFailure", func(t *testing.T) {
		expectedErr := errors.New("load error")
		res := &mockLoadResource{
			err: expectedErr,
		}

		ctx := context.Background()
		_, err := Load[*test.Config](ctx, res)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	// 测试用例4: 上下文取消
	t.Run("ContextCancellation", func(t *testing.T) {
		res := &mockLoadResource{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := Load[*test.Config](ctx, res)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got '%v'", err)
		}
	})

	// 测试用例5: JSON转换失败
	t.Run("JSONMarshalFailure", func(t *testing.T) {
		// 创建一个无法被JSON序列化的无效结构体
		invalidStruct := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"invalid": {Kind: &structpb.Value_NumberValue{NumberValue: 0}},
			},
		}

		res := &mockLoadResource{value: invalidStruct}

		ctx := context.Background()
		_, err := Load[*test.Config](ctx, res)
		if err == nil {
			t.Error("Expected JSON marshal error, got nil")
		}
	})

	// 测试用例6: 空资源列表
	t.Run("EmptyResources", func(t *testing.T) {
		ctx := context.Background()
		_, err := Load[*test.Config](ctx)
		if err != nil {
			t.Errorf("Expected no error with empty resources, got %v", err)
		}
	})
}
