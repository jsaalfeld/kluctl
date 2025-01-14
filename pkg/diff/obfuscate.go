package diff

import (
	"fmt"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
	"github.com/ohler55/ojg/jp"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

var secretGvk = schema.GroupKind{Group: "", Kind: "Secret"}

type Obfuscator struct {
}

func (o *Obfuscator) Obfuscate(ref k8s.ObjectRef, changes []types.Change) error {
	if ref.GVK.GroupKind() == secretGvk {
		err := o.obfuscateSecret(ref, changes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Obfuscator) obfuscateSecret(ref k8s.ObjectRef, changes []types.Change) error {
	replaceValues := func(x any, v string) any {
		if x == nil {
			return nil
		}
		if m, ok := x.(map[string]any); ok {
			for k, _ := range m {
				m[k] = v
			}
			return m
		} else if a, ok := x.([]any); ok {
			for i, _ := range a {
				a[i] = v
			}
			return a
		}
		return v
	}

	for i, _ := range changes {
		c := &changes[i]
		j, err := jp.ParseString(c.JsonPath)
		if err != nil {
			return err
		}
		if len(j) == 0 {
			return fmt.Errorf("unexpected empty jsonPath")
		}
		child, ok := j[0].(jp.Child)
		if !ok {
			return fmt.Errorf("unexpected jsonPath fragment: %s", c.JsonPath)
		}

		if child == "data" || child == "stringData" {
			c.NewValue = replaceValues(c.NewValue, "*****a")
			c.OldValue = replaceValues(c.OldValue, "*****b")
			_ = updateUnifiedDiff(c)
			c.NewValue = replaceValues(c.NewValue, "*****")
			c.OldValue = replaceValues(c.OldValue, "*****")
			c.UnifiedDiff = strings.ReplaceAll(c.UnifiedDiff, "*****a", "***** (obfuscated)")
			c.UnifiedDiff = strings.ReplaceAll(c.UnifiedDiff, "*****b", "***** (obfuscated)")
		}
	}
	return nil
}
