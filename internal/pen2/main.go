package pen

import (
	"log"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/swag/gen"
)

func pen() {
	doc, err := loads.Spec("../template/swagger.yaml")
	if err != nil {
		log.Fatalf("Failed to load spec: %v", err)
	}

	specDoc := doc.Spec()

	// 生成 model 文件
	modelName := "User" // 指定要生成的 model 类型名称
	modelPath := "model" // 指定 model 文件输出的路径
	models := []*spec.Schema{
		specDoc.Components.Schemas[modelName],
	}

	err = gen.WriteModels(
		swag.ToFileName(modelName, modelPath),
		models,
		gen.NewDefaultGenOpts(),
	)
	if err != nil {
		log.Fatalf("Failed to write models: %v", err)
	}

	// 生成 logic 文件
	logicPath := "bizlogic" // 指定 bizlogic 文件输出的路径

	err = gen.WriteServerOperation(
		specDoc,
		logicPath,
		gen.NewServerOperationOpts(),
	)
	if err != nil {
		log.Fatalf("Failed to write server operation: %v", err)
	}

	log.Println("Code generation completed successfully!")
}
