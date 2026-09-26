package testx

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestGenerate(t *testing.T) {
	value, err := Generate("user-{uuid}")
	if err != nil {
		t.Fatalf("生成字符串失败：%v", err)
	}
	if !strings.HasPrefix(value, "user-") || len(value) == len("user-") {
		t.Fatalf("生成结果不符合预期：%q", value)
	}
}

func TestStruct(t *testing.T) {
	type user struct {
		Name  string `fake:"{firstname}"`
		Email string `fake:"{email}"`
	}

	var value user
	if err := Struct(&value); err != nil {
		t.Fatalf("填充结构体失败：%v", err)
	}
	if value.Name == "" || value.Email == "" {
		t.Fatalf("结构体字段未正确填充：%+v", value)
	}
}

func TestAddFuncLookup(t *testing.T) {
	const name = "testxcustomvalue"
	gofakeit.RemoveFuncLookup(name)
	t.Cleanup(func() { gofakeit.RemoveFuncLookup(name) })

	AddFuncLookup(name, Info{
		Generate: func(_ *Faker, _ *MapParams, _ *Info) (any, error) {
			return "custom", nil
		},
	})

	value, err := Generate("{" + name + "}")
	if err != nil {
		t.Fatalf("调用自定义生成函数失败：%v", err)
	}
	if value != "custom" {
		t.Fatalf("自定义生成结果不符合预期：%q", value)
	}
}
