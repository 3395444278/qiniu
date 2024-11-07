package models

import (
	"strings"
)

type PredictionResult struct {
	Nation     string   `json:"nation"`
	Confidence float64  `json:"confidence"`
	Factors    []string `json:"factors"`
}

// ExtractNation 从位置字符串中提取国家信息
func ExtractNation(location string) string {
	if location == "" {
		return ""
	}

	// 简单的国家关键词匹配
	countryKeywords := map[string]string{
		// 亚洲
		"中国":        "CN",
		"china":     "CN",
		"beijing":   "CN",
		"shanghai":  "CN",
		"shenzhen":  "CN",
		"japan":     "JP",
		"日本":        "JP",
		"tokyo":     "JP",
		"osaka":     "JP",
		"korea":     "KR",
		"韩国":        "KR",
		"seoul":     "KR",
		"singapore": "SG",
		"新加坡":       "SG",
		"india":     "IN",
		"印度":        "IN",
		"bangalore": "IN",
		"vietnam":   "VN",
		"越南":        "VN",

		// 北美
		"usa":           "US",
		"united states": "US",
		"美国":            "US",
		"california":    "US",
		"san francisco": "US",
		"new york":      "US",
		"seattle":       "US",
		"canada":        "CA",
		"加拿大":           "CA",
		"toronto":       "CA",

		// 欧洲
		"uk":             "GB",
		"united kingdom": "GB",
		"england":        "GB",
		"london":         "GB",
		"germany":        "DE",
		"德国":             "DE",
		"berlin":         "DE",
		"france":         "FR",
		"法国":             "FR",
		"paris":          "FR",
		"netherlands":    "NL",
		"荷兰":             "NL",
		"amsterdam":      "NL",
		"sweden":         "SE",
		"瑞典":             "SE",
		"stockholm":      "SE",
	}

	locationLower := strings.ToLower(location)
	for keyword, code := range countryKeywords {
		if strings.Contains(locationLower, strings.ToLower(keyword)) {
			return code
		}
	}

	return ""
}
