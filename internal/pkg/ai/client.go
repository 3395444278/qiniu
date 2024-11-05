package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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
	return fmt.Sprintf(`请根据以下信息全面评估该开发者的技术能力，并以 JSON 格式返回评估结果：

开发者基本信息：
- 用户名：%v
- 姓名：%v
- 简介：%v
- 位置：%v
- 邮箱：%v

在线资料：
- GitHub 主页：%v
- 博客：%v
- 个人网站：%v

技术信息：
- 技术栈：%v
- 编程语言：%v
- 主要仓库：%v
- 仓库链接：%v

贡献统计：
- Stars 总数：%v
- Commits 总数：%v
- Forks 总数：%v
- 各仓库 Stars：%v

活跃度：
- 最近活跃：%v
- 账号创建：%v
- 最后更新：%v

请返回以下格式的 JSON：
{
    "specialties": ["主要专长领域1", "主要专长领域2", ...],
    "experience": {
        "技术1": "详细的经验评估",
        "技术2": "详细的经验评估",
        ...
    },
    "evaluation": "全面的技术能力评价，包括技术深度、广度、项目质量等方面"
}
`,
		info["username"], info["name"], info["bio"], info["location"], info["email"],
		info["profile_url"], info["blog"], info["personal_site"],
		info["skills"], info["languages"], info["repos"], info["repo_urls"],
		info["stars"], info["commits"], info["forks"], info["repo_stars"],
		info["last_active"], info["created_at"], info["updated_at"])
}

type EvaluationResult struct {
	Specialties   []string          `json:"specialties"`
	Experience    map[string]string `json:"experience"`
	AIEvaluation  string            `json:"evaluation"`
	LastEvaluated time.Time         `json:"last_evaluated"`
}

func parseAIResponse(response string) (*EvaluationResult, error) {
	log.Printf("Raw AI response: %s", response)

	// 首先尝试清理响应中的 JSON 字符串
	cleanedResponse := response
	if strings.Contains(response, "```json") {
		// 提取 JSON 部分
		start := strings.Index(response, "```json\n") + 8
		end := strings.Index(response[start:], "```")
		if end != -1 {
			cleanedResponse = response[start : start+end]
		}
	}

	// 尝试解析 JSON
	var result struct {
		Specialties []string          `json:"specialties"`
		Experience  map[string]string `json:"experience"`
		Evaluation  string            `json:"evaluation"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		log.Printf("Error parsing AI response: %v", err)

		// 如果解析失败，尝试提取有用信息
		specialties := extractSpecialties(response)
		experience := extractExperience(response)
		evaluation := extractEvaluation(response)

		// 确保返回的 map 不为 nil
		if experience == nil {
			experience = make(map[string]string)
		}

		// 返回提取的信息
		return &EvaluationResult{
			Specialties:   specialties,
			Experience:    experience,
			AIEvaluation:  evaluation,
			LastEvaluated: time.Now(),
		}, nil
	}

	// 确保返回的 map 不为 nil
	if result.Experience == nil {
		result.Experience = make(map[string]string)
	}
	if result.Specialties == nil {
		result.Specialties = make([]string, 0)
	}

	// JSON 解析成功
	return &EvaluationResult{
		Specialties:   result.Specialties,
		Experience:    result.Experience,
		AIEvaluation:  result.Evaluation,
		LastEvaluated: time.Now(),
	}, nil
}

// 辅助函数：提取专长信息
func extractSpecialties(text string) []string {
	specialties := make([]string, 0)
	re := regexp.MustCompile(`专长[：:]\s*(.*?)(?:\n|$)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		// 分割并清理结果
		for _, s := range strings.Split(matches[1], ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				specialties = append(specialties, s)
			}
		}
	}
	return specialties
}

// 辅助函数：提取经验信息
func extractExperience(text string) map[string]string {
	experience := make(map[string]string)
	re := regexp.MustCompile(`(\w+)[：:]\s*(.*?)(?:\n|$)`)
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 2 {
			tech := strings.TrimSpace(match[1])
			level := strings.TrimSpace(match[2])
			if tech != "" && level != "" {
				experience[tech] = level
			}
		}
	}
	return experience
}

// 辅助函数：提取评估信息
func extractEvaluation(text string) string {
	// 首先尝试提取 evaluation 字段
	var result struct {
		Evaluation string `json:"evaluation"`
	}
	if err := json.Unmarshal([]byte(text), &result); err == nil && result.Evaluation != "" {
		return result.Evaluation
	}

	// 如果失败，尝试使用正则表达式
	re := regexp.MustCompile(`评[价估][：:]\s*(.*?)(?:\n|$)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 如果都失败了，返回整个文本
	return strings.TrimSpace(text)
}
