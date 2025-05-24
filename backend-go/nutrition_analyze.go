package main

import (
	"context"
	"fmt"

	dashscopego "github.com/eswulei/dashscope-go"
	"github.com/eswulei/dashscope-go/qwen"
)

func AnalyzeFoodImage2(ImageURL string) *FoodAnalysisResult {
	model := qwen.QwenVLPlus
	token := "sk-6326ac6e33024d8fa3abeeedef503abd"

	if token == "" {
		panic("token is empty")
	}

	cli := dashscopego.NewTongyiClient(model, token)

	sysContent := qwen.VLContentList{
		{
			Text: "You are a helpful assistant.",
		},
	}
	userContent := qwen.VLContentList{
		{
			Text: "What is the nutritional content with accurate numbers in this picture and what foods are in it and output ingredients and nutrition only, no explanations",
		},
		{
			Image: ImageURL,
		},
	}

	input := dashscopego.VLInput{
		Messages: []dashscopego.VLMessage{
			{Role: "system", Content: &sysContent},
			{Role: "user", Content: &userContent},
		},
	}

	// (可选 SSE开启)需要流式输出时 通过该 Callback Function 获取结果
	streamCallbackFn := func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}
	req := &dashscopego.VLRequest{
		Input:       input,
		StreamingFn: streamCallbackFn,
	}

	ctx := context.TODO()
	resp, err := cli.CreateVLCompletion(ctx, req)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nnon-stream result: ")
	fmt.Println(resp.Output.Choices[0].Message.Content.ToString())

	return processFoodAnalysis(resp.Output.Choices[0].Message.Content.ToString())
}
