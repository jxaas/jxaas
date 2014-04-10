package bundle

import (
	"testing"

	"github.com/justinsb/gova/log"
)

func TestBundleStore_Get(t *testing.T) {
	store := NewBundleStore("../templates")
	templateContext := &TemplateContext{}
	tenant := "123"
	serviceType := "mysql"
	name := "test"

	templateContext.NumberUnits = 3
	templateContext.Options = map[string]string{}
	templateContext.Options["performance"] = "high"

	bundle, err := store.GetBundle(templateContext, tenant, serviceType, name)

	if err != nil {
		t.Fatal("Unable to load bundle", err)
	}

	if bundle == nil {
		t.Fatal("Bundle was nil")
	}

	prefix := buildPrefix(tenant, serviceType, name)

	service, found := bundle.Services[prefix+"mysql"]
	if !found {
		log.Info("Services: %v", bundle.Services)

		t.Fatal("mysql service not found")
	}

	if service.NumberUnits != 3 {
		t.Fatal("NumberUnits was not copied")
	}

	if service.Options["performance"] != "high" {
		t.Fatal("Performance option was not copied")
	}
}
