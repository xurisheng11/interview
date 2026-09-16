package service

import (
	"reflect"
	"testing"
)

// 验证 DeepSeek 点评 JSON 的各类漂移形态都能被宽松解析
func TestParseReviewJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want ReviewResult
	}{
		{
			name: `标准对象`,
			raw:  `{"score":85,"pros":["优点1","优点2"],"cons":["不足1"],"referenceAnswer":"参考答案"}`,
			want: ReviewResult{Score: 85, Pros: []string{"优点1", "优点2"}, Cons: []string{"不足1"}, ReferenceAnswer: "参考答案"},
		},
		{
			name: "前后带说明文字",
			raw:  "好的，以下是评分结果：\n{\"score\":70,\"pros\":[\"a\"],\"cons\":[],\"referenceAnswer\":\"答案\"}\n希望有帮助！",
			want: ReviewResult{Score: 70, Pros: []string{"a"}, ReferenceAnswer: "答案"},
		},
		{
			name: "代码块包裹",
			raw:  "```json\n{\"score\":60,\"pros\":[\"x\"],\"cons\":[\"y\"],\"referenceAnswer\":\"ref\"}\n```",
			want: ReviewResult{Score: 60, Pros: []string{"x"}, Cons: []string{"y"}, ReferenceAnswer: "ref"},
		},
		{
			name: `pros漂移为字符串、score漂移为字符串`,
			raw:  `{"score":"80","pros":"表达清晰","cons":[],"referenceAnswer":"r"}`,
			want: ReviewResult{Score: 80, Pros: []string{"表达清晰"}, ReferenceAnswer: "r"},
		},
		{
			name: `数组内混数字与嵌套`,
			raw:  `{"score":88,"pros":["要点1",2,{"k":"v"}],"cons":[],"referenceAnswer":"r","expressionScore":"90","detectedVerbalTics":["然后"]}`,
			want: ReviewResult{Score: 88, Pros: []string{"要点1", "2", `{"k":"v"}`}, ReferenceAnswer: "r", ExpressionScore: 90, DetectedVerbalTics: []string{"然后"}},
		},
		{
			name: `JSON字符串内含花括号`,
			raw:  `{"score":75,"pros":[],"cons":[],"referenceAnswer":"HashMap 结构是 {key:value} 形式"}`,
			want: ReviewResult{Score: 75, ReferenceAnswer: "HashMap 结构是 {key:value} 形式"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got ReviewResult
			if err := parseReviewJSON(c.raw, &got); err != nil {
				t.Fatalf("解析失败: %v", err)
			}
			if got.Score != c.want.Score || got.ExpressionScore != c.want.ExpressionScore || got.ReferenceAnswer != c.want.ReferenceAnswer {
				t.Fatalf("标量字段不符: %+v want %+v", got, c.want)
			}
			if len(c.want.Pros) > 0 && !reflect.DeepEqual(got.Pros, c.want.Pros) {
				t.Fatalf("pros 不符: got %q want %q", got.Pros, c.want.Pros)
			}
			if len(c.want.DetectedVerbalTics) > 0 && !reflect.DeepEqual(got.DetectedVerbalTics, c.want.DetectedVerbalTics) {
				t.Fatalf("口头禅不符: got %q want %q", got.DetectedVerbalTics, c.want.DetectedVerbalTics)
			}
		})
	}
}

// 缺 score / 完全非 JSON 应明确报错
func TestParseReviewJSONError(t *testing.T) {
	var got ReviewResult
	if err := parseReviewJSON(`{"pros":["a"]}`, &got); err == nil {
		t.Fatal("缺 score 应报错")
	}
	if err := parseReviewJSON("模型服务异常，稍后再试", &got); err == nil {
		t.Fatal("非 JSON 应报错")
	}
}
