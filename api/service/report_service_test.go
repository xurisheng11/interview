package service

import (
	"strings"
	"testing"

	"interview-sim/model"
)

// 说明：GetSharedReport 的脱敏逻辑与 repository.Get（Redis）强耦合（包级函数、
// 无接口注入点），无法在不修改生产代码结构的前提下 mock，故本批次跳过，
// 详见测试计划说明。本文件覆盖同文件内不依赖外部的纯函数。

// TestCalcModuleScores_LevelBoundaries 模块分级边界：<60 待提升 / [60,75) 及格 / [75,90) 良好 / >=90 优秀
func TestCalcModuleScores_LevelBoundaries(t *testing.T) {
	cases := []struct {
		score     int
		wantLevel string
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
		got := calcModuleScores(map[string][]int{"模块": {c.score}})
		if len(got) != 1 {
			t.Fatalf("calcModuleScores 单模块应返回1项, got %d", len(got))
		}
		if got[0].Level != c.wantLevel {
			t.Errorf("分数 %d 的 level = %q, want %q", c.score, got[0].Level, c.wantLevel)
		}
	}
}

// TestCalcModuleScores_Average 平均分与题数统计
func TestCalcModuleScores_Average(t *testing.T) {
	got := calcModuleScores(map[string][]int{"算法": {80, 90, 70}})
	if len(got) != 1 {
		t.Fatalf("want 1 module, got %d", len(got))
	}
	m := got[0]
	if m.Module != "算法" {
		t.Errorf("Module = %q, want 算法", m.Module)
	}
	if m.AvgScore != 80.0 {
		t.Errorf("AvgScore = %v, want 80", m.AvgScore)
	}
	if m.Count != 3 {
		t.Errorf("Count = %d, want 3", m.Count)
	}
}

// TestCalcModuleScores_FloatAverage 非整除平均值为浮点数
func TestCalcModuleScores_FloatAverage(t *testing.T) {
	got := calcModuleScores(map[string][]int{"网络": {70, 75}})
	if got[0].AvgScore != 72.5 {
		t.Errorf("AvgScore = %v, want 72.5", got[0].AvgScore)
	}
}

// TestCalcModuleScores_Empty 空 map / nil 防御
func TestCalcModuleScores_Empty(t *testing.T) {
	if got := calcModuleScores(map[string][]int{}); len(got) != 0 {
		t.Errorf("空map应返回空结果, got %d 项", len(got))
	}
	if got := calcModuleScores(nil); len(got) != 0 {
		t.Errorf("nil map应返回空结果, got %d 项", len(got))
	}
}

// TestCalcModuleScores_MultiModule 多模块各自独立统计
func TestCalcModuleScores_MultiModule(t *testing.T) {
	got := calcModuleScores(map[string][]int{
		"算法":   {90, 92},
		"数据库": {50},
	})
	if len(got) != 2 {
		t.Fatalf("want 2 modules, got %d", len(got))
	}
	byName := map[string]model.ModuleScore{}
	for _, m := range got {
		byName[m.Module] = m
	}
	if m := byName["算法"]; m.AvgScore != 91 || m.Level != "优秀" {
		t.Errorf("算法 = %+v, want avg 91 优秀", m)
	}
	if m := byName["数据库"]; m.AvgScore != 50 || m.Level != "待提升" {
		t.Errorf("数据库 = %+v, want avg 50 待提升", m)
	}
}

// TestBuildScoresSummary 得分摘要拼装：跳过题标注为0分、正常题带分数
func TestBuildScoresSummary(t *testing.T) {
	t.Run("空切片返回空串", func(t *testing.T) {
		if got := buildScoresSummary(nil); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("跳过与作答混合", func(t *testing.T) {
		qs := []model.ReportQuestion{
			{Index: 0, Tags: []string{"Go"}, Score: 85},
			{Index: 1, Tags: []string{"Redis"}, Skipped: true},
		}
		got := buildScoresSummary(qs)
		if !strings.Contains(got, "题1[Go]: 85分") {
			t.Errorf("缺少作答题摘要: %q", got)
		}
		if !strings.Contains(got, "题2[Redis]: 跳过(0分)") {
			t.Errorf("缺少跳过题摘要: %q", got)
		}
	})
}
