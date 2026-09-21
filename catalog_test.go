package catalog

import (
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
    entries, err := Load(file)
    if err != nil {
        t.Fatal(err)
    }
    if len(entries) < 2 {
        t.Fatalf("目录条目不足: %d", len(entries))
    }
    // 六个核心维度彼此独立，且每个维度至少有一条术语。
    required := map[string]bool{
        "使用目的": false, "记忆类别": false, "角色能力": false,
        "年龄级别": false, "风险信号": false, "支持资源": false,
    }
    for _, entry := range entries {
        if covered, ok := required[entry.Kind]; ok && !covered {
            required[entry.Kind] = true
        }
    }
    for kind, covered := range required {
        if !covered {
            t.Errorf("维度缺少条目: %s", kind)
        }
    }
}

func TestLoadRejectsUnknownKind(t *testing.T) {
    _, err := Load(strings.NewReader(`[{"id":"x","kind":"诊断标签","name":"抑郁"}]`))
    if err == nil || !strings.Contains(err.Error(), "类别不受控") {
        t.Fatalf("应拒绝不受控类别，得到: %v", err)
    }
}

func TestLoadRejectsDuplicateID(t *testing.T) {
    input := `[
        {"id":"dup","kind":"使用目的","name":"甲"},
        {"id":"dup","kind":"使用目的","name":"乙"}
    ]`
    _, err := Load(strings.NewReader(input))
    if err == nil || !strings.Contains(err.Error(), "标识重复") {
        t.Fatalf("应拒绝重复标识，得到: %v", err)
    }
}

func TestLoadRejectsMissingFields(t *testing.T) {
    if _, err := Load(strings.NewReader(`[{"id":"","kind":"使用目的","name":"x"}]`)); err == nil {
        t.Fatal("应拒绝缺少标识的条目")
    }
    if _, err := Load(strings.NewReader(`[{"id":"x","kind":"使用目的","name":""}]`)); err == nil {
        t.Fatal("应拒绝缺少名称的条目")
    }
}
