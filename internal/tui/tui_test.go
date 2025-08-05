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
	
	// Test new state file field
	if model.stateFile {
		t.Error("Expected state file generation to be false by default")
	}
	
	if model.verbose {
		t.Error("Expected verbose mode to be false by default")
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