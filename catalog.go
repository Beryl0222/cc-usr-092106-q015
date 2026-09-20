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

// Load 读取目录，并拒绝缺少标识、名称或重复标识的数据。
func Load(reader io.Reader) ([]Entry, error) {
    var entries []Entry
    if err := json.NewDecoder(reader).Decode(&entries); err != nil {
        return nil, fmt.Errorf("读取目录失败: %w", err)
    }
    seen := map[string]bool{}
    for _, entry := range entries {
        if entry.ID == "" || entry.Name == "" {
            return nil, fmt.Errorf("目录条目缺少标识或名称")
        }
        if seen[entry.ID] {
            return nil, fmt.Errorf("目录标识重复: %s", entry.ID)
        }
        seen[entry.ID] = true
    }
    return entries, nil
}
