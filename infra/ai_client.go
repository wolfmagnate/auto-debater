package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"google.golang.org/genai"
)

// ChatCompletionHandler は、指定されたモデルとプロンプトを使用してLLMからテキストを生成し、
// 指定されたスキーマTに結果を非整列化します。Thinking機能もサポートします。
func ChatCompletionHandler[T any](ctx context.Context, prompt string, thinkingBudget *int32) (*T, *genai.GenerateContentResponseUsageMetadata, error) {
	// 1. genai.Clientを初期化します。
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, nil, errors.New("GOOGLE_API_KEY環境変数が設定されていません")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("クライアントの作成に失敗しました: %w", err)
	}

	// 2. レスポンススキーマを生成します。
	var targetSchema T
	schemaType := reflect.TypeOf(targetSchema)
	responseSchema, err := generateSchemaFromType(schemaType)
	if err != nil {
		return nil, nil, fmt.Errorf("型からのスキーマ生成に失敗しました: %w", err)
	}

	// 3. GenerateContentConfig を作成します。
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   responseSchema,
		// Tools: []*genai.Tool{
		//	{
		//		GoogleSearch: &genai.GoogleSearch{},
		//	},
		//},
	}

	if thinkingBudget != nil {
		config.ThinkingConfig = &genai.ThinkingConfig{
			ThinkingBudget: thinkingBudget,
		}
	}

	// 4. LLMモデルを選択し、コンテンツを生成します。
	modelIdentifier := "gemini-2.5-flash-preview-05-20"
	contents := []*genai.Content{genai.NewContentFromText(prompt, genai.RoleUser)}

	resp, err := client.Models.GenerateContent(ctx, modelIdentifier, contents, config)
	if err != nil {
		return nil, nil, fmt.Errorf("コンテンツの生成に失敗しました: %w", err)
	}

	// 5. レスポンスを処理します。
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, resp.UsageMetadata, errors.New("モデルからの応答がありません")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var jsonText string

	if part != nil { // part 自体がnilでないことを確認
		if part.Text != "" {
			jsonText = part.Text
		} else if part.InlineData != nil && part.InlineData.MIMEType == "application/json" {
			jsonText = string(part.InlineData.Data)
		} else {
			return nil, resp.UsageMetadata, fmt.Errorf("応答の最初のパートに予期されるJSONテキストが含まれていません。受信パート: %+v", part)
		}
	} else {
		return nil, resp.UsageMetadata, errors.New("モデル応答の最初のパートがnilです")
	}

	// 6. JSONテキストをターゲットスキーマTに非整列化します。
	var result T
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, resp.UsageMetadata, fmt.Errorf("JSONの非整列化に失敗しました: %w (JSON: %s)", err, jsonText)
	}

	return &result, resp.UsageMetadata, nil
}

// generateSchemaFromType は、指定されたGoの型からgenai.Schemaオブジェクトを生成するヘルパー関数です。
func generateSchemaFromType(t reflect.Type) (*genai.Schema, error) {
	schema := &genai.Schema{}

	switch t.Kind() {
	case reflect.Struct:
		schema.Type = genai.TypeObject
		schema.Properties = make(map[string]*genai.Schema)
		var requiredFields []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			fieldName := field.Name
			omitempty := false
			if jsonTag != "" && jsonTag != "-" {
				parts := strings.Split(jsonTag, ",")
				fieldName = parts[0]
				if len(parts) > 1 {
					for _, opt := range parts[1:] {
						if opt == "omitempty" {
							omitempty = true
						}
					}
				}
			}
			if jsonTag == "-" {
				continue
			}

			if field.Type == t {
				schema.Properties[fieldName] = &genai.Schema{Type: genai.TypeString, Description: fmt.Sprintf("Recursive field %s, using string as placeholder", fieldName)}
				continue
			}

			propSchema, err := generateSchemaFromType(field.Type)
			if err != nil {
				return nil, fmt.Errorf("フィールド '%s' のスキーマ生成に失敗しました: %w", fieldName, err)
			}
			schema.Properties[fieldName] = propSchema
			if !omitempty {
				requiredFields = append(requiredFields, fieldName)
			}
		}
		if len(requiredFields) > 0 {
			schema.Required = requiredFields
		}
	case reflect.String:
		schema.Type = genai.TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = genai.TypeInteger
	case reflect.Float32, reflect.Float64:
		schema.Type = genai.TypeNumber
	case reflect.Bool:
		schema.Type = genai.TypeBoolean
	case reflect.Slice, reflect.Array:
		schema.Type = genai.TypeArray
		if t.Elem() == t {
			schema.Items = &genai.Schema{Type: genai.TypeString, Description: "Recursive array/slice element, using string as placeholder"}
		} else {
			elemSchema, err := generateSchemaFromType(t.Elem())
			if err != nil {
				return nil, fmt.Errorf("配列/スライス要素のスキーマ生成に失敗しました: %w", err)
			}
			schema.Items = elemSchema
		}
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			schema.Type = genai.TypeNumber
			schema.Description = fmt.Sprintf("Map with string keys and %s values. Note: genai.Schema for map values might require specific handling or manual definition for 'additionalProperties'.", t.Elem().Kind().String())
		} else {
			return nil, fmt.Errorf("キーが文字列ではないマップは、OpenAPIスキーマとして直接表現するのが困難です: %s", t.Kind())
		}
	case reflect.Ptr:
		if t.Elem() == t {
			return &genai.Schema{Type: genai.TypeString, Description: fmt.Sprintf("Recursive pointer type %s, using string as placeholder", t.String())}, nil
		}
		return generateSchemaFromType(t.Elem())
	default:
		schema.Type = genai.TypeString
		schema.Description = fmt.Sprintf("Unsupported type %s encountered, treated as string.", t.Kind())
	}
	return schema, nil
}
