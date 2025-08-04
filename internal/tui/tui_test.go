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
	
	if model.selectedRegion != "us-east-1" {
		t.Errorf("Expected default region to be us-east-1, got %s", model.selectedRegion)
	}
	
	if model.outputFolder != "./infrastructure" {
		t.Errorf("Expected default output folder to be ./infrastructure, got %s", model.outputFolder)
	}
}

func TestNewDiscoveryManager(t *testing.T) {
	dm := NewDiscoveryManager("us-west-2", "./test-output", true)
	
	if dm.region != "us-west-2" {
		t.Errorf("Expected region to be us-west-2, got %s", dm.region)
	}
	
	if dm.output != "./test-output" {
		t.Errorf("Expected output to be ./test-output, got %s", dm.output)
	}
	
	if !dm.verbose {
		t.Error("Expected verbose to be true")
	}
	
	if dm.progressCh == nil {
		t.Error("Expected progress channel to be initialized")
	}
	
	if dm.completeCh == nil {
		t.Error("Expected complete channel to be initialized")
	}
	
	if dm.errorCh == nil {
		t.Error("Expected error channel to be initialized")
	}
}

func TestEnhancedModel(t *testing.T) {
	model := NewEnhancedModel()
	
	// Test that enhanced model has base model functionality
	if model.currentStep != StepRegionSelection {
		t.Errorf("Expected initial step to be StepRegionSelection, got %v", model.currentStep)
	}
	
	if len(model.availableResources) == 0 {
		t.Error("Expected available resources to be populated")
	}
	
	if len(model.availableRegions) == 0 {
		t.Error("Expected available regions to be populated")
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
	}
	
	if len(steps) != 6 {
		t.Errorf("Expected 6 step types, got %d", len(steps))
	}
} 