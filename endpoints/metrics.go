package endpoints

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
)

type EndpointMetrics struct {
	Parent *EndpointCharm
}

type Metrics struct {
	Metric []string
}

func (self *EndpointMetrics) HttpGet() (*Metrics, error) {
	//service := self.Parent.Key

	// TODO: Inject
	// TODO: Use an ES client that isn't a singleton
	api.Domain = "10.0.3.72"
	api.Port = "9200"

	//	var searchQuery = `{
	//		"query": {
	//			"term": {"sex":"female"}
	//		}
	//	}`

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]string{"Type": "logfile"},
		},
	}

	response, err := core.SearchRequest("_all", "message", nil, query)
	if err != nil {
		log.Warn("Error searching elasticsearch", err)
		return nil, fmt.Errorf("Error searching elasticsearch")
	}

	metrics := &Metrics{}
	metrics.Metric = []string{}

	for _, v := range response.Hits.Hits {
		// TODO: Are we serializing and deserializing here??
		json, err := v.Source.MarshalJSON()
		if err != nil {
			log.Warn("Error reading JSON", err)
			return nil, fmt.Errorf("Error searching elasticsearch")
		}

		m := string(json)

		metrics.Metric = append(metrics.Metric, m)

		//		var value map[string]interface{}
		//		err := json.Unmarshal(v.Source, &value)
		//		if err != nil {
		//			log.Warn("Error unmarshalling response", err)
		//			return nil, fmt.Errorf("Error searching elasticsearch")
		//		}
		//		values = append(values, value)
	}
	//	fmt.Println(values)

	return metrics, nil
}
