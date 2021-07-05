package resource

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
	This helper is used to replace developing provider which contains a sdk upgrade and released provider.
    By deploying resources with released provider and importing resources with developing provider, to verify whether there's backend breaking change introduced.
*/

func isCrossVersionImportEnabled() bool {
	if enabled, err := strconv.ParseBool(os.Getenv("TF_ACC_CROSS_VERSION_IMPORT")); err == nil {
		return enabled
	}
	return true
}

func getStandardProviderName() string {
	if provider := os.Getenv("TF_ACC_PROVIDER"); provider != "" {
		return provider
	}
	return "azurerm"
}

func getStandardProviderNamespace() string {
	if namespace := os.Getenv("TF_ACC_PROVIDER_NAMESPACE"); namespace != "" {
		return namespace
	}
	return "hashicorp"
}

func getStandardProviderVersion() string {
	return os.Getenv("TF_ACC_PROVIDER_VERSION")
}

func isFixImportStepEnabled() bool {
	if enabled, err := strconv.ParseBool(os.Getenv("TF_ACC_FIX_IMPORT")); err == nil {
		return enabled
	}
	return true
}

/*
replace develop provider with released provider
*/
func useExternalProvider(c TestCase) func() (*schema.Provider, error) {
	if isCrossVersionImportEnabled() {
		provider := getStandardProviderName()
		externalProvider := ExternalProvider{
			Source: fmt.Sprintf("registry.terraform.io/%s/%s", getStandardProviderNamespace(), provider),
		}
		if version := getStandardProviderVersion(); version != "" {
			externalProvider.VersionConstraint = "=" + version
		}
		c.ExternalProviders[provider] = externalProvider
		backupProviderFactory := c.ProviderFactories[provider]
		delete(c.ProviderFactories, provider)
		return backupProviderFactory
	}
	return nil
}

/*
replace released provider with develop provider
*/
func useDevelopProvider(c TestCase, backupProviderFactory func() (*schema.Provider, error)) {
	if isCrossVersionImportEnabled() {
		provider := getStandardProviderName()
		c.ProviderFactories[provider] = backupProviderFactory
		delete(c.ExternalProviders, provider)
	}
}

func addImportSteps(steps []TestStep) []TestStep {
	if !isFixImportStepEnabled() || len(steps) == 0 || len(steps[0].ResourceName) == 0 {
		return []TestStep{}
	}
	resourceName := steps[0].ResourceName
	results := make([]TestStep, 0)
	for i, step := range steps {
		results = append(results, step)
		if !step.ImportState && step.ExpectError == nil && (i == len(steps)-1 || !steps[i+1].ImportState) {
			results = append(results, TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			})
		}
	}
	return results
}
