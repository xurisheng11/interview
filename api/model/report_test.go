package model

import "testing"

// TestCalcGrade 分级边界值：<60 待提升 / [60,75) 及格 / [75,90) 良好 / >=90 优秀
func TestCalcGrade(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{0, "待提升"},
		{59, "待提升"},
		{60, "及格"},
		{74, "及格"},
		{75, "良好"},
		{89, "良好"},
		{90, "优秀"},
		{100, "优秀"},
	}
	for _, c := range cases {
		if got := CalcGrade(c.score); got != c.want {
			t.Errorf("CalcGrade(%d) = %q, want %q", c.score, got, c.want)
		}
	}
}

// TestCalcPassStatus 通过判定：分数边界 + 跳过比例规则
func TestCalcPassStatus(t *testing.T) {
	cases := []struct {
		name       string
		score      int
		totalCount int
		skipped    int
		wantStatus string
		wantReason string
	}{
		{"59分未通过", 59, 10, 0, "fail", ""},
		{"60分边界待定", 60, 10, 0, "pending", ""},
		{"74分待定", 74, 10, 0, "pending", ""},
		{"75分边界通过", 75, 10, 0, "pass", ""},
		{"100分通过", 100, 10, 0, "pass", ""},
		{"跳过超过一半直接fail", 90, 10, 6, "fail", "跳过题目过多"},
		{"跳过恰好一半不触发fail", 75, 10, 5, "pass", ""},
		{"totalCount为0不除零_按分数判定", 80, 0, 0, "pass", ""},
		{"totalCount为0低分fail", 30, 0, 0, "fail", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, reason := CalcPassStatus(c.score, c.totalCount, c.skipped)
			if status != c.wantStatus || reason != c.wantReason {
				t.Errorf("CalcPassStatus(%d, %d, %d) = (%q, %q), want (%q, %q)",
					c.score, c.totalCount, c.skipped, status, reason, c.wantStatus, c.wantReason)
			}
		})
	}
}

// TestCalcAvgExpressionScore 平均表达得分：空/nil 防御 + 过滤跳过题与零分
func TestCalcAvgExpressionScore(t *testing.T) {
	t.Run("nil切片返回0", func(t *testing.T) {
		if got := CalcAvgExpressionScore(nil); got != 0 {
			t.Errorf("CalcAvgExpressionScore(nil) = %d, want 0", got)
		}
	})

	t.Run("空切片返回0", func(t *testing.T) {
		if got := CalcAvgExpressionScore([]ReportQuestion{}); got != 0 {
			t.Errorf("CalcAvgExpressionScore(empty) = %d, want 0", got)
		}
	})

	t.Run("全部跳过返回0", func(t *testing.T) {
		qs := []ReportQuestion{
			{Skipped: true, ExpressionScore: 80},
			{Skipped: true, ExpressionScore: 90},
		}
		if got := CalcAvgExpressionScore(qs); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("零分不参与平均", func(t *testing.T) {
		qs := []ReportQuestion{
			{Skipped: false, ExpressionScore: 0},  // 零值过滤
			{Skipped: false, ExpressionScore: 80},
			{Skipped: false, ExpressionScore: 90},
			{Skipped: true, ExpressionScore: 10},  // 跳过过滤
		}
		if got := CalcAvgExpressionScore(qs); got != 85 {
			t.Errorf("got %d, want 85", got)
		}
	})

	t.Run("整数除法截断", func(t *testing.T) {
		qs := []ReportQuestion{
			{ExpressionScore: 80},
			{ExpressionScore: 85},
			{ExpressionScore: 84},
		}
		// (80+85+84)/3 = 83 (249/3)
		if got := CalcAvgExpressionScore(qs); got != 83 {
			t.Errorf("got %d, want 83", got)
		}
	})
}

// TestCalcAvgSpeechRate 平均语速：nil metrics 防御 + 过滤零值
func TestCalcAvgSpeechRate(t *testing.T) {
	t.Run("nil切片返回0", func(t *testing.T) {
		if got := CalcAvgSpeechRate(nil); got != 0 {
			t.Errorf("got %v, want 0", got)
		}
	})

	t.Run("metrics为nil不panic且返回0", func(t *testing.T) {
		qs := []ReportQuestion{
			{Skipped: false, NonVerbalMetrics: nil},
		}
		if got := CalcAvgSpeechRate(qs); got != 0 {
			t.Errorf("got %v, want 0", got)
		}
	})

	t.Run("过滤零语速和跳过题后取平均", func(t *testing.T) {
		qs := []ReportQuestion{
			{NonVerbalMetrics: &NonVerbalMetrics{SpeechRate: 120}},
			{NonVerbalMetrics: &NonVerbalMetrics{SpeechRate: 140}},
			{NonVerbalMetrics: &NonVerbalMetrics{SpeechRate: 0}},                 // 零值过滤
			{Skipped: true, NonVerbalMetrics: &NonVerbalMetrics{SpeechRate: 999}}, // 跳过过滤
		}
		if got := CalcAvgSpeechRate(qs); got != 130 {
			t.Errorf("got %v, want 130", got)
		}
	})
}

// TestCalcAvgThinkDuration 平均思考时长：空/nil 防御 + 过滤零值 + 整数截断
func TestCalcAvgThinkDuration(t *testing.T) {
	t.Run("nil切片返回0", func(t *testing.T) {
		if got := CalcAvgThinkDuration(nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("metrics为nil不panic且返回0", func(t *testing.T) {
		qs := []ReportQuestion{{NonVerbalMetrics: nil}}
		if got := CalcAvgThinkDuration(qs); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("过滤零值后整数平均", func(t *testing.T) {
		qs := []ReportQuestion{
			{NonVerbalMetrics: &NonVerbalMetrics{ThinkDuration: 10}},
			{NonVerbalMetrics: &NonVerbalMetrics{ThinkDuration: 15}},
			{NonVerbalMetrics: &NonVerbalMetrics{ThinkDuration: 0}},              // 零值过滤
			{Skipped: true, NonVerbalMetrics: &NonVerbalMetrics{ThinkDuration: 100}}, // 跳过过滤
		}
		// (10+15)/2 = 12（整数截断）
		if got := CalcAvgThinkDuration(qs); got != 12 {
			t.Errorf("got %d, want 12", got)
		}
	})
}

// TestCalcVerbalTicReport 口头禅汇总：无口头禅返回 nil，有则统计频次
func TestCalcVerbalTicReport(t *testing.T) {
	t.Run("无口头禅返回nil", func(t *testing.T) {
		qs := []ReportQuestion{
			{NonVerbalMetrics: &NonVerbalMetrics{}},
			{NonVerbalMetrics: nil},
		}
		if got := CalcVerbalTicReport(qs); got != nil {
			t.Errorf("got %+v, want nil", got)
		}
	})

	t.Run("汇总频次_跳过题不计入", func(t *testing.T) {
		qs := []ReportQuestion{
			{NonVerbalMetrics: &NonVerbalMetrics{VerbalTics: []string{"嗯", "然后"}}},
			{NonVerbalMetrics: &NonVerbalMetrics{VerbalTics: []string{"嗯"}}},
			{Skipped: true, NonVerbalMetrics: &NonVerbalMetrics{VerbalTics: []string{"那个"}}},
		}
		got := CalcVerbalTicReport(qs)
		if got == nil {
			t.Fatal("got nil, want report")
		}
		if got.TicFrequency != 3 {
			t.Errorf("TicFrequency = %d, want 3", got.TicFrequency)
		}
		if len(got.DetectedTics) != 2 {
			t.Errorf("DetectedTics len = %d, want 2（跳过题的口头禅不应计入）", len(got.DetectedTics))
		}
	})
}
