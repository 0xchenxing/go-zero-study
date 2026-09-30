package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=devuser password=devpass dbname=appdb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:           "./rpc/internal/dal/query",
		ModelPkgPath:      "model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldWithIndexTag: true,
	})

	g.UseDB(db)

	// 精确指定表，避免 GenerateAllTable 引入元数据表导致失败。
	// password 的 json tag 置为 "-"，避免生成的模型被直接序列化时泄露密码哈希。
	g.ApplyBasic(
		g.GenerateModel("users",
			gen.FieldJSONTag("password", "-"),
		),
		g.GenerateModel("todos"),
	)

	g.Execute()
	fmt.Println("code generation done")
}
