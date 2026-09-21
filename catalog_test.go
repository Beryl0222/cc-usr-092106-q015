package catalog

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLoadFixture(t *testing.T) {
	file, err := os.Open("fixtures/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	cat, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Purposes) < 2 {
		t.Fatalf("使用目的条目不足: %d", len(cat.Purposes))
	}
	if len(cat.RiskSignals) < 2 {
		t.Fatalf("风险信号条目不足: %d", len(cat.RiskSignals))
	}
}

// validCatalog 返回一份满足全部不变量的目录，测试用例在其上做单点破坏。
func validCatalog() *Catalog {
	return &Catalog{
		Purposes: []Purpose{
			{ID: "p1", Name: "目的一", Description: "说明"},
		},
		MemoryKinds: []MemoryKind{
			{ID: "m1", Name: "对话上下文", Retention: RetentionConversation, Revocable: true},
			{ID: "m2", Name: "偏好", Retention: RetentionPersistent, Revocable: true},
		},
		Capabilities: []Capability{
			{ID: "c1", Name: "情绪陪伴", Boundary: "不做诊断"},
		},
		AgeTiers: []AgeTier{
			{ID: "a1", Name: "成年"},
		},
		RiskSignals: []RiskSignal{
			{ID: "s1", Name: "普通低落", Severity: SeverityOrdinaryEmotion},
			{ID: "s2", Name: "紧迫危险", Severity: SeverityClearUrgent, RuleID: "r1"},
		},
		Resources: []Resource{
			{ID: "res1", Name: "紧急电话", Channel: "拨打急救电话"},
		},
		Rules: []PublicRule{
			{ID: "r1", Name: "危险规则", Basis: "具体计划与即时时间", ResourceIDs: []string{"res1"}},
		},
		Interventions: []Intervention{
			{
				ID:                "i1",
				Name:              "危机提示",
				VisibleBasis:      "命中规则 r1",
				Action:            "展示求助渠道",
				ExpiresAfter:      "720h",
				ReviewMinFragment: "仅单条触发消息",
			},
		},
		EventTypes: []EventType{
			{ID: "e1", Code: "model-upgrade", Name: "模型升级", Description: "版本变化"},
			{ID: "e2", Code: "message-edit-delete", Name: "消息删改", Description: "删改"},
			{ID: "e3", Code: "cross-device-retry", Name: "跨设备重试", Description: "重试"},
			{ID: "e4", Code: "false-positive-appeal", Name: "误报申诉", Description: "申诉"},
			{ID: "e5", Code: "long-high-frequency-use", Name: "长期高频使用", Description: "高频"},
			{ID: "e6", Code: "deletion-cascade", Name: "删除级联", Description: "级联"},
		},
		Deletion: DeletionCascade{Targets: []string{ReplicaIndex, ReplicaResearch}},
	}
}

func loadCatalog(t *testing.T, c *Catalog) error {
	t.Helper()
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Load(strings.NewReader(string(data)))
	return err
}

func TestValidCatalogAccepted(t *testing.T) {
	if err := loadCatalog(t, validCatalog()); err != nil {
		t.Fatalf("合法目录被拒绝: %v", err)
	}
}

func TestValidationRejectsViolations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Catalog)
		want   string
	}{
		{
			name:   "使用目的为空",
			mutate: func(c *Catalog) { c.Purposes = nil },
			want:   "使用目的",
		},
		{
			name:   "记忆不可撤销",
			mutate: func(c *Catalog) { c.MemoryKinds[0].Revocable = false },
			want:   "撤销",
		},
		{
			name:   "记忆保留范围非法",
			mutate: func(c *Catalog) { c.MemoryKinds[0].Retention = "forever" },
			want:   "保留范围",
		},
		{
			name:   "能力缺少边界说明",
			mutate: func(c *Catalog) { c.Capabilities[0].Boundary = "" },
			want:   "边界",
		},
		{
			name:   "年龄级别为空",
			mutate: func(c *Catalog) { c.AgeTiers = nil },
			want:   "年龄级别",
		},
		{
			name:   "普通情绪被诊断化",
			mutate: func(c *Catalog) { c.RiskSignals[0].RuleID = "r1" },
			want:   "诊断化",
		},
		{
			name:   "紧迫信号缺少公开规则",
			mutate: func(c *Catalog) { c.RiskSignals[1].RuleID = "" },
			want:   "公开规则",
		},
		{
			name:   "紧迫信号引用不存在规则",
			mutate: func(c *Catalog) { c.RiskSignals[1].RuleID = "nope" },
			want:   "不存在的规则",
		},
		{
			name:   "规则缺少可见依据",
			mutate: func(c *Catalog) { c.Rules[0].Basis = "" },
			want:   "依据",
		},
		{
			name:   "规则未引用现实资源",
			mutate: func(c *Catalog) { c.Rules[0].ResourceIDs = nil },
			want:   "现实支持资源",
		},
		{
			name:   "资源缺少渠道",
			mutate: func(c *Catalog) { c.Resources[0].Channel = "" },
			want:   "渠道",
		},
		{
			name:   "没有现实支持资源",
			mutate: func(c *Catalog) { c.Resources = nil },
			want:   "现实支持资源",
		},
		{
			name:   "干预缺少可见依据",
			mutate: func(c *Catalog) { c.Interventions[0].VisibleBasis = "" },
			want:   "可见依据",
		},
		{
			name:   "干预缺少动作",
			mutate: func(c *Catalog) { c.Interventions[0].Action = "" },
			want:   "动作",
		},
		{
			name:   "干预未声明最小复核片段",
			mutate: func(c *Catalog) { c.Interventions[0].ReviewMinFragment = "" },
			want:   "最小片段",
		},
		{
			name:   "干预永久不解除",
			mutate: func(c *Catalog) { c.Interventions[0].ExpiresAfter = "0s" },
			want:   "解除",
		},
		{
			name:   "干预解除时间非法",
			mutate: func(c *Catalog) { c.Interventions[0].ExpiresAfter = "soon" },
			want:   "解除时间",
		},
		{
			name:   "缺少必须可解释事件",
			mutate: func(c *Catalog) { c.EventTypes = c.EventTypes[:5] },
			want:   "可解释的事件",
		},
		{
			name:   "删除级联漏掉研究副本",
			mutate: func(c *Catalog) { c.Deletion.Targets = []string{ReplicaIndex} },
			want:   "研究副本",
		},
		{
			name:   "删除级联漏掉索引",
			mutate: func(c *Catalog) { c.Deletion.Targets = []string{ReplicaResearch} },
			want:   "索引",
		},
		{
			name:   "标识重复",
			mutate: func(c *Catalog) { c.AgeTiers[0].ID = "p1" },
			want:   "重复",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validCatalog()
			tt.mutate(c)
			err := loadCatalog(t, c)
			if err == nil {
				t.Fatalf("期望拒绝（含 %q），实际通过", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("错误 %q 不含 %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRejectUnknownFields(t *testing.T) {
	raw := `{
		"purposes": [{"id":"p1","name":"目的","description":"说明"}],
		"memory_kinds": [{"id":"m1","name":"上下文","retention":"conversation","revocable":true}],
		"capabilities": [{"id":"c1","name":"陪伴","boundary":"边界"}],
		"age_tiers": [{"id":"a1","name":"成年"}],
		"risk_signals": [],
		"resources": [],
		"rules": [],
		"interventions": [],
		"event_types": [],
		"deletion_cascade": {"targets": []},
		"diagnosis_labels": []
	}`
	if _, err := Load(strings.NewReader(raw)); err == nil {
		t.Fatal("含未知字段的目录应被拒绝")
	}
}
