package catalog

import (
    "encoding/json"
    "fmt"
    "io"
)

// Entry 表示业务目录中的一条基础资料。
type Entry struct {
    ID   string `json:"id"`
    Kind string `json:"kind"`
    Name string `json:"name"`
}

// Kinds 列出彼此独立的受控维度。一条目录条目只能归属一个维度，
// 这样使用目的、记忆、能力、年龄级别、风险信号与支持资源不会相互混用。
var Kinds = []string{
    "使用目的",
    "记忆类别",
    "角色能力",
    "年龄级别",
    "风险信号",
    "支持资源",
    "干预动作",
    "人工复核",
    "可解释事件",
    "数据副本",
}

// Load 读取目录，并拒绝缺少标识、名称、类别非法或标识重复的数据。
func Load(reader io.Reader) ([]Entry, error) {
    var entries []Entry
    if err := json.NewDecoder(reader).Decode(&entries); err != nil {
        return nil, fmt.Errorf("读取目录失败: %w", err)
    }
    allowedKinds := make(map[string]bool, len(Kinds))
    for _, kind := range Kinds {
        allowedKinds[kind] = true
    }
    seen := map[string]bool{}
    for _, entry := range entries {
        if entry.ID == "" || entry.Name == "" {
            return nil, fmt.Errorf("目录条目缺少标识或名称")
        }
        if !allowedKinds[entry.Kind] {
            return nil, fmt.Errorf("目录条目类别不受控: %s", entry.Kind)
        }
        if seen[entry.ID] {
            return nil, fmt.Errorf("目录标识重复: %s", entry.ID)
        }
        seen[entry.ID] = true
    }
    return entries, nil
}
