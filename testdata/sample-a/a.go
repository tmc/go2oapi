package main

// SampleFunction is a function that exists to serve as an example.
func SampleFunction(a string, b int) (string, error) {
	return a, nil
}

// SampleFunctionB is a sample function.
func SampleFunctionB(paramToEcho struct {
	// Field A is great!
	FieldA string // This is after the fact.

	// Field B rocks.
	FieldB string `enum:"foo,bar" json:"fieldb"`

	// Field C broh
	FieldC []int
}) (string, error) {
	return paramToEcho.FieldA, nil
}

type NewWidgetFactoryOptions struct {
	// The name of the factory
	FactoryName string
	// Category
	Category string `enum:"foo,bar" json:"fieldb"`
	// InventoryLevels
	InventoryLevels []int
	Operational     bool
}

// NewWidgetFactory creates a new widget factory.
func NewWidgetFactory(factoryInfo NewWidgetFactoryOptions) (string, error) {
	return "", nil
}
