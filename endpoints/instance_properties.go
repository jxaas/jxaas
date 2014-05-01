package endpoints

type EndpointInstanceProperties struct {
	Parent *EndpointInstance
}

//type Properties struct {
//	Properties map[string]string
//}

//func (self *EndpointInstanceProperties) HttpGet(apiclient *juju.Client) (*Properties, error) {
//	service := self.Parent.ServiceName()
//
//	apiclient.GetProperties(serviceId)
//
//	lines := &Lines{}
//	lines.Line = make([]string, 0)
//
//	logfile.ReadLines(func(line string) (bool, error) {
//		lines.Line = append(lines.Line, line)
//		return true, nil
//	})
//
//	return lines, nil
//}
