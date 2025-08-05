package tui

import (
	"testing"
)

func TestNewModel(t *testing.T) {
	model := NewModel()
	
	// Test initial state
	if model.currentStep != StepRegionSelection {
		t.Errorf("Expected initial step to be StepRegionSelection, got %v", model.currentStep)
	}
	
	if len(model.availableResources) == 0 {
		t.Error("Expected available resources to be populated")
	}
	
	if len(model.availableRegions) == 0 {
		t.Error("Expected available regions to be populated")
	}
	
	// Test that we have a comprehensive list of regions (should be more than 30)
	if len(model.availableRegions) < 30 {
		t.Errorf("Expected at least 30 regions, got %d", len(model.availableRegions))
	}
	
	if model.selectedRegion != "us-east-1" {
		t.Errorf("Expected default region to be us-east-1, got %s", model.selectedRegion)
	}
	
	if model.outputFolder != "./infrastructure" {
		t.Errorf("Expected default output folder to be ./infrastructure, got %s", model.outputFolder)
	}
	
	// Test new state file field
	if model.stateFile {
		t.Error("Expected state file generation to be false by default")
	}
	
	if model.verbose {
		t.Error("Expected verbose mode to be false by default")
	}
	
	// Test region scroll offset
	if model.regionScrollOffset != 0 {
		t.Error("Expected region scroll offset to be 0 initially")
	}
}

func TestStepTypes(t *testing.T) {
	// Test that all step types are defined
	steps := []StepType{
		StepRegionSelection,
		StepServiceSelection,
		StepOutputFolder,
		StepConfirmation,
		StepDiscovery,
		StepResults,
		StepGeneration,
	}
	
	if len(steps) != 7 {
		t.Errorf("Expected 7 step types, got %d", len(steps))
	}
}

func TestStateFileConfiguration(t *testing.T) {
	model := NewModel()
	
	// Test initial state
	if model.stateFile {
		t.Error("Expected state file to be false initially")
	}
	
	// Test toggling state file
	model.stateFile = true
	if !model.stateFile {
		t.Error("Expected state file to be true after setting")
	}
	
	// Test toggling back
	model.stateFile = false
	if model.stateFile {
		t.Error("Expected state file to be false after toggling back")
	}
}

func TestRegionListCompleteness(t *testing.T) {
	model := NewModel()
	
	// Test that important regions are included
	importantRegions := []string{
		"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1",
		"eu-central-1", "ap-northeast-1", "sa-east-1",
	}
	
	for _, region := range importantRegions {
		found := false
		for _, availableRegion := range model.availableRegions {
			if availableRegion == region {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected region %s to be in available regions list", region)
		}
	}
	
	// Test that we have regions from different continents
	continentRegions := map[string][]string{
		"US": {"us-east-1", "us-west-2"},
		"Europe": {"eu-west-1", "eu-central-1"},
		"Asia": {"ap-southeast-1", "ap-northeast-1"},
		"South America": {"sa-east-1"},
	}
	
	for continent, regions := range continentRegions {
		found := false
		for _, region := range regions {
			for _, availableRegion := range model.availableRegions {
				if availableRegion == region {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("Expected at least one region from %s to be in available regions list", continent)
		}
	}
} 

func TestRegionCount(t *testing.T) {
	model := NewModel()
	totalRegions := len(model.availableRegions)
	t.Logf("Total AWS regions available: %d", totalRegions)
	
	// Verify we have a comprehensive list
	if totalRegions < 30 {
		t.Errorf("Expected at least 30 regions, got %d", totalRegions)
	}
} 