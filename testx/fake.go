package testx

import "github.com/brianvoe/gofakeit/v7"

// Faker、Info、MapParams 和 Param 是 gofakeit 对应类型的别名。
type (
	Faker     = gofakeit.Faker
	Info      = gofakeit.Info
	MapParams = gofakeit.MapParams
	Param     = gofakeit.Param
)

// AddFuncLookup、Generate 和 Struct 是 gofakeit 对应函数的别名。
var (
	AddFuncLookup = gofakeit.AddFuncLookup
	Generate      = gofakeit.Generate
	Struct        = gofakeit.Struct
)
