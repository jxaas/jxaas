package core

type StubPortAssigner struct {
}

func (self *StubPortAssigner) AssignPort() (int, error) {
	return 0, nil
}

func (self *StubPortAssigner) GetAssignedPort() (int, bool) {
	return 0, false
}

type InstancePublicPortAssigner struct {
	Instance *Instance
	Assigned bool
	Port     int
}

func (self *InstancePublicPortAssigner) AssignPort() (int, error) {
	primaryServiceId := self.Instance.primaryServiceId
	port, assigned, err := self.Instance.huddle.assignPublicPort(primaryServiceId)
	if assigned {
		self.Assigned = true
	}
	self.Port = port
	return port, err
}

func (self *InstancePublicPortAssigner) GetAssignedPort() (int, bool) {
	return self.Port, self.Assigned
}
