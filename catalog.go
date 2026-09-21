// Package catalog 定义陪伴式对话产品的安全边界目录。
//
// 目录把彼此独立的六类资料放在一起维护：使用目的、记忆类别、角色能力、
// 年龄级别、风险信号、现实支持资源，并附公开干预规则、有限干预模板与
// 可解释事件类型。Load 负责强制本包的安全不变量；本包不分析对话内容，
// 也不直接实施干预。
package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// 风险信号级别。普通情绪表达不得被诊断化；只有明确而紧迫的危险信号
// 才能挂接公开规则，提示现实求助。
const (
	SeverityOrdinaryEmotion = "ordinary-emotion"
	SeverityClearUrgent     = "clear-urgent-danger"
)

// 记忆保留范围。默认只保留当前对话所需信息。
const (
	RetentionConversation = "conversation"
	RetentionPersistent   = "persistent"
)

// 删除级联必须同步失效的副本类别。
const (
	ReplicaIndex    = "index"
	ReplicaResearch = "research"
)

// RequiredEventCodes 是规格要求可解释、不得静默发生的六类事件：
// 模型升级、消息删改、跨设备重试、误报申诉、长期高频使用、删除级联。
var RequiredEventCodes = []string{
	"model-upgrade",
	"message-edit-delete",
	"cross-device-retry",
	"false-positive-appeal",
	"long-high-frequency-use",
	"deletion-cascade",
}

// Purpose 说明一项数据或功能被收集/启用的使用目的。目的轴线与其他
// 轴线独立登记，便于单独向用户展示与收缩。
type Purpose struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MemoryKind 是一类可被记住的内容。所有记忆类别必须可由用户查看、
// 收缩或删除（Revocable），默认只在当前对话保留。
type MemoryKind struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Retention string `json:"retention"` // conversation 或 persistent
	Revocable bool   `json:"revocable"`
}

// Capability 描述陪伴角色能做什么、以及明确做不到什么（Boundary），
// 避免把短暂安慰表述成长期福祉或临床能力。
type Capability struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Boundary string `json:"boundary"`
}

// AgeTier 是独立的年龄级别轴线，不与记忆或风险判定耦合。
type AgeTier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RiskSignal 是一类风险信号。ordinary-emotion 不得挂接规则（不得被
// 诊断化）；clear-urgent-danger 必须挂接一条存在的公开规则。
type RiskSignal struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Severity string `json:"severity"`
	RuleID   string `json:"rule_id,omitempty"`
}

// Resource 是一条现实支持资源，供公开规则引用。
type Resource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Channel string `json:"channel"`
}

// PublicRule 是触发现实求助提示所依据的公开规则：可见依据与可提供的
// 现实支持资源都必须写明。
type PublicRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Basis       string   `json:"basis"`
	ResourceIDs []string `json:"resource_ids"`
}

// Intervention 是一次有限干预的模板。每次干预都必须能向用户说明
// 可见依据、采取的动作与解除时间；人工复核只接触最小片段。
type Intervention struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	VisibleBasis      string `json:"visible_basis"`
	Action            string `json:"action"`
	ExpiresAfter      string `json:"expires_after"` // 有限时长，如 "720h"，不得为永久
	ReviewMinFragment string `json:"review_min_fragment"`
}

// EventType 是一类必须留痕、可向用户解释的事件。旧状态既不会悄然
// 消失，也不会因一次事件被永久贴标签（干预均有解除时间）。
type EventType struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeletionCascade 声明数据删除后必须同步失效的副本。相关索引与研究
// 副本无一例外都要在列。
type DeletionCascade struct {
	Targets []string `json:"targets"`
}

// Catalog 是完整的安全边界目录。
type Catalog struct {
	Purposes      []Purpose       `json:"purposes"`
	MemoryKinds   []MemoryKind    `json:"memory_kinds"`
	Capabilities  []Capability    `json:"capabilities"`
	AgeTiers      []AgeTier       `json:"age_tiers"`
	RiskSignals   []RiskSignal    `json:"risk_signals"`
	Resources     []Resource      `json:"resources"`
	Rules         []PublicRule    `json:"rules"`
	Interventions []Intervention  `json:"interventions"`
	EventTypes    []EventType     `json:"event_types"`
	Deletion      DeletionCascade `json:"deletion_cascade"`
}

// Load 读取并校验目录。任何违反安全不变量的目录都会被拒绝。
func Load(reader io.Reader) (*Catalog, error) {
	var catalog Catalog
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}
	if err := catalog.validate(); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func (c *Catalog) validate() error {
	ids := map[string]bool{}
	check := func(id, name string) error {
		if id == "" || name == "" {
			return fmt.Errorf("目录条目缺少标识或名称")
		}
		if ids[id] {
			return fmt.Errorf("目录标识重复: %s", id)
		}
		ids[id] = true
		return nil
	}

	if len(c.Purposes) == 0 {
		return fmt.Errorf("使用目的不得为空：必须说明数据为何而用")
	}
	for _, p := range c.Purposes {
		if err := check(p.ID, p.Name); err != nil {
			return err
		}
		if p.Description == "" {
			return fmt.Errorf("使用目的缺少说明: %s", p.ID)
		}
	}

	if len(c.MemoryKinds) == 0 {
		return fmt.Errorf("记忆类别不得为空：至少要登记当前对话所需信息")
	}
	for _, m := range c.MemoryKinds {
		if err := check(m.ID, m.Name); err != nil {
			return err
		}
		if !m.Revocable {
			return fmt.Errorf("记忆类别必须可由用户撤销: %s", m.ID)
		}
		switch m.Retention {
		case RetentionConversation, RetentionPersistent:
		default:
			return fmt.Errorf("记忆类别保留范围非法: %s", m.ID)
		}
	}

	if len(c.Capabilities) == 0 {
		return fmt.Errorf("角色能力不得为空：必须说清产品做不到什么")
	}
	for _, capItem := range c.Capabilities {
		if err := check(capItem.ID, capItem.Name); err != nil {
			return err
		}
		if capItem.Boundary == "" {
			return fmt.Errorf("角色能力缺少边界说明: %s", capItem.ID)
		}
	}

	if len(c.AgeTiers) == 0 {
		return fmt.Errorf("年龄级别不得为空")
	}
	for _, a := range c.AgeTiers {
		if err := check(a.ID, a.Name); err != nil {
			return err
		}
	}

	resources := map[string]bool{}
	for _, r := range c.Resources {
		if err := check(r.ID, r.Name); err != nil {
			return err
		}
		if r.Channel == "" {
			return fmt.Errorf("现实支持资源缺少渠道: %s", r.ID)
		}
		resources[r.ID] = true
	}
	if len(c.Resources) == 0 {
		return fmt.Errorf("现实支持资源不得为空：紧迫信号必须有现实出口")
	}

	rules := map[string]bool{}
	for _, r := range c.Rules {
		if err := check(r.ID, r.Name); err != nil {
			return err
		}
		if r.Basis == "" {
			return fmt.Errorf("公开规则缺少可见依据: %s", r.ID)
		}
		if len(r.ResourceIDs) == 0 {
			return fmt.Errorf("公开规则必须引用现实支持资源: %s", r.ID)
		}
		for _, rid := range r.ResourceIDs {
			if !resources[rid] {
				return fmt.Errorf("公开规则 %s 引用了不存在的资源: %s", r.ID, rid)
			}
		}
		rules[r.ID] = true
	}

	if len(c.RiskSignals) == 0 {
		return fmt.Errorf("风险信号不得为空：普通情绪与紧迫信号需分别登记")
	}
	for _, s := range c.RiskSignals {
		if err := check(s.ID, s.Name); err != nil {
			return err
		}
		switch s.Severity {
		case SeverityOrdinaryEmotion:
			if s.RuleID != "" {
				return fmt.Errorf("普通情绪表达不得被诊断化或挂接规则: %s", s.ID)
			}
		case SeverityClearUrgent:
			if s.RuleID == "" {
				return fmt.Errorf("明确紧迫信号必须挂接公开规则: %s", s.ID)
			}
			if !rules[s.RuleID] {
				return fmt.Errorf("风险信号 %s 引用了不存在的规则: %s", s.ID, s.RuleID)
			}
		default:
			return fmt.Errorf("风险信号级别非法: %s", s.ID)
		}
	}

	if len(c.Interventions) == 0 {
		return fmt.Errorf("有限干预模板不得为空")
	}
	for _, iv := range c.Interventions {
		if err := check(iv.ID, iv.Name); err != nil {
			return err
		}
		if iv.VisibleBasis == "" {
			return fmt.Errorf("干预必须说明可见依据: %s", iv.ID)
		}
		if iv.Action == "" {
			return fmt.Errorf("干预必须说明采取的动作: %s", iv.ID)
		}
		if iv.ReviewMinFragment == "" {
			return fmt.Errorf("干预必须声明人工复核的最小片段: %s", iv.ID)
		}
		d, err := time.ParseDuration(iv.ExpiresAfter)
		if err != nil {
			return fmt.Errorf("干预解除时间非法: %s: %w", iv.ID, err)
		}
		if d <= 0 {
			return fmt.Errorf("干预必须是有限且可解除的: %s", iv.ID)
		}
	}

	if len(c.EventTypes) == 0 {
		return fmt.Errorf("可解释事件类型不得为空")
	}
	events := map[string]bool{}
	for _, e := range c.EventTypes {
		if err := check(e.ID, e.Name); err != nil {
			return err
		}
		if e.Code == "" || e.Description == "" {
			return fmt.Errorf("事件类型缺少代码或说明: %s", e.ID)
		}
		if events[e.Code] {
			return fmt.Errorf("事件代码重复: %s", e.Code)
		}
		events[e.Code] = true
	}
	for _, code := range RequiredEventCodes {
		if !events[code] {
			return fmt.Errorf("缺少必须可解释的事件类型: %s", code)
		}
	}

	hasIndex, hasResearch := false, false
	for _, t := range c.Deletion.Targets {
		switch t {
		case ReplicaIndex:
			hasIndex = true
		case ReplicaResearch:
			hasResearch = true
		}
	}
	if !hasIndex || !hasResearch {
		return fmt.Errorf("删除级联必须覆盖相关索引与研究副本")
	}

	return nil
}
