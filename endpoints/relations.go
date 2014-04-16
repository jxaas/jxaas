package endpoints

type EndpointRelations struct {
	Parent *EndpointInstance
}

func (self *EndpointRelations) Item(key string) *EndpointRelation {
	child := &EndpointRelation{}
	child.Parent = self
	child.RelationKey = key
	return child
}
