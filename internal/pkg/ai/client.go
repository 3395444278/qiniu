package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	apiBaseURL string
	model      string
}

func NewClient(apiKey string) *Client {
	log.Printf("Initializing AI client with key: %s...", apiKey[:10])

	return &Client{
		apiKey:     apiKey,
		apiBaseURL: "https://api.deepseek.com/chat/completions",
		model:      "deepseek-chat",
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) EvaluateDeveloper(ctx context.Context, info map[string]interface{}) (*EvaluationResult, error) {
	prompt := buildEvaluationPrompt(info)

	request := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "你是一个专业的技术人才评估专家。请根据提供的信息评估开发者的技术能力，并以 JSON 格式返回评估结果。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	log.Printf("Sending request to DeepSeek API with URL: %s", c.apiBaseURL)
	log.Printf("Request headers: %+v", req.Header)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("DeepSeek API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DeepSeek API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from DeepSeek API")
	}

	log.Printf("DeepSeek API Response: %s", chatResp.Choices[0].Message.Content)

	result, err := parseAIResponse(chatResp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	return result, nil
}

func buildEvaluationPrompt(info map[string]interface{}) string {
	return fmt.Sprintf(`请根据以下信息评估该开发者的技术能力，重点分析其可能所在的国家/地区。请特别关注：
1. 用户名特征
2. 代码注释语言
3. 项目描述语言
4. 技术栈特点
5. 活跃时间规律

开发者信息：
- 用户名：%v
- 名称：%v
- 邮箱：%v
- 位置：%v
- GitHub主页：%v
- 技术栈：%v
- 编程语言：%v
- 主要仓库：%v
- 提交数：%v
- Stars数：%v
- Forks数：%v
- 最近活跃：%v

请以JSON格式返回以下信息：
{
    "nation": "CN",  // 两位国家代码：CN中国,US美国,JP日本,KR韩国,SG新加坡等
    "confidence": 85, // 置信度0-100
    "reasons": ["原因1", "原因2"], // 判断依据
    "specialties": ["专长1", "专长2"],
    "evaluation": "整体评价"
}`,
		info["username"], info["name"], info["email"], info["location"],
		info["profile_url"], info["skills"], info["languages"],
		info["repos"], info["commits"], info["stars"],
		info["forks"], info["last_active"])
}

type EvaluationResult struct {
	Specialties   []string          `json:"specialties"`
	Experience    map[string]string `json:"experience"`
	AIEvaluation  string            `json:"evaluation"`
	Nation        string            `json:"nation"`
	Confidence    float64           `json:"confidence"`
	LastEvaluated time.Time         `json:"last_evaluated"`
}

func parseAIResponse(response string) (*EvaluationResult, error) {
	log.Printf("Raw AI response: %s", response)

	cleanedResponse := response
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json\n") + 8
		end := strings.Index(response[start:], "```")
		if end != -1 {
			cleanedResponse = response[start : start+end]
		}
	}

	// 定义一个临时结构体来解析完整的响应
	var temp struct {
		Nation      string   `json:"nation"`
		Confidence  float64  `json:"confidence"`
		Reasons     []string `json:"reasons"`
		Specialties []string `json:"specialties"`
		Evaluation  string   `json:"evaluation"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &temp); err != nil {
		log.Printf("Error parsing AI response: %v", err)
		return nil, fmt.Errorf("解析 AI 响应失败: %v", err)
	}

	// 转换为 EvaluationResult
	result := &EvaluationResult{
		Specialties:   temp.Specialties,
		Experience:    make(map[string]string),
		AIEvaluation:  temp.Evaluation,
		Nation:        temp.Nation,
		Confidence:    temp.Confidence,
		LastEvaluated: time.Now(),
	}

	return result, nil
}
