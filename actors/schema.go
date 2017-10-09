package actors

// SchemaDefinition defines the interface to the definition of schema
type SchemaDefinition interface {
	Name() string
	Get() map[string]interface{}
}

type DefaultSchemaDefinition struct{}

func (self *DefaultSchemaDefinition) Name() string {
	return "DefaultSchemaDefinition"
}

func (self *DefaultSchemaDefinition) Get() map[string]interface{} {
	return map[string]interface{}{}
}
