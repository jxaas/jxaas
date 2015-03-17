package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"

	elasticgo_api "github.com/mattbaird/elastigo/api"
	elastigo_core "github.com/mattbaird/elastigo/core"
)

func (self *Instance) readMetrics(jujuUnitNames []string, metricId string) (*model.MetricDataset, error) {
	instance := self

	keyValue := metricId
	keyTimestamp := "Timestamp"
	keyHostname := "Hostname"
	keyType := "Type"

	huddle := instance.huddle

	es := huddle.SharedServices["jx-elasticsearch"]
	if es == nil {
		return nil, rs.ErrNotFound()
	}

	if len(jujuUnitNames) == 0 {
		return nil, nil
	}

	// TODO: Inject
	// TODO: Use an ES client that isn't a singleton
	elasticgo_api.Domain = es.PublicAddress
	elasticgo_api.Port = "9200"

	// TODO: We need to make sure that most fields are _not_ analyzed
	// That is why we have match below, not term

	filters := []interface{}{}

	{
		match := map[string]string{}
		match[keyHostname] = jujuUnitNames[0]
		filter := map[string]interface{}{"query": map[string]interface{}{"match": match}}
		filters = append(filters, filter)
	}

	{
		// TODO: Hard-coded
		match := map[string]string{}
		match[keyType] = "LoadAverage"
		filter := map[string]interface{}{"query": map[string]interface{}{"match": match}}
		filters = append(filters, filter)
	}

	if len(filters) > 1 {
		and := map[string]interface{}{"and": filters}
		filters = []interface{}{and}
	}

	match_all := map[string]interface{}{"match_all": map[string]string{}}
	filtered := map[string]interface{}{"filter": filters[0], "query": match_all}

	query := map[string]interface{}{"filtered": filtered}

	body := map[string]interface{}{"query": query}

	args := map[string]interface{}{}
	args["size"] = 1000

	response, err := elastigo_core.SearchRequest("_all", "message", args, body)
	if err != nil {
		log.Warn("Error searching elasticsearch", err)
		return nil, fmt.Errorf("Error searching elasticsearch")
	}

	metrics := &model.MetricDataset{}
	metrics.Points = []model.MetricDatapoint{}

	for _, hit := range response.Hits.Hits {
		// TODO: Are we serializing and deserializing here??
		jsonBytes, err := hit.Source.MarshalJSON()
		if err != nil {
			log.Warn("Error reading JSON", err)
			return nil, fmt.Errorf("Error searching elasticsearch")
		}

		//log.Info("Found metric: %v", string(jsonBytes))

		var value map[string]interface{}
		err = json.Unmarshal(jsonBytes, &value)
		if err != nil {
			log.Warn("Error unmarshalling response", err)
			return nil, fmt.Errorf("Error searching elasticsearch")
		}

		// Post-filter the ES results...
		// TODO: See if we can persuade ES to filter it correctly
		hostname, found := value[keyHostname]
		if !found {
			log.Debug("No hostname in %v", string(jsonBytes))
			continue
		}

		hostnameStr, ok := hostname.(string)
		if !ok {
			log.Debug("Cannot cast hostname to string: %v", hostname)
			continue
		}

		if hostnameStr != jujuUnitNames[0] {
			log.Debug("Post-filtering query results from ES")
			continue
		}

		// Grab the timestamp & value
		t, found := value[keyTimestamp]
		if !found {
			log.Debug("No timestamp in %v", string(jsonBytes))
			continue
		}

		tStr, ok := t.(string)
		if !ok {
			log.Debug("Cannot cast timestamp to string: %v", t)
			continue
		}

		timeFormat := time.RFC3339
		tVal, err := time.Parse(timeFormat, tStr)
		if err != nil {
			log.Debug("Cannot parse timestamp: %v", tStr, err)
			continue
		}

		y, found := value[keyValue]
		if !found {
			log.Debug("No value (%v) in %v", keyValue, string(jsonBytes))
			continue
		}

		yStr, ok := y.(string)
		if !ok {
			log.Debug("Cannot cast value to string: %v", y)
			continue
		}

		yVal, err := strconv.ParseFloat(yStr, 32)
		if err != nil {
			log.Debug("Error parsing value as float: %v", yStr, err)
			continue
		}

		p := model.MetricDatapoint{}
		p.T = tVal.Unix()
		p.V = float32(yVal)
		metrics.Points = append(metrics.Points, p)
	}
	//	fmt.Println(values)

	sort.Sort(metrics.Points)

	return metrics, nil
}

// Retrieves metrics that apply to the instance
func (self *Instance) GetMetricInfo() (*model.Metrics, error) {
	metrics := &model.Metrics{}

	// TODO: Hard-coded
	// TODO: Store in metadata file?
	metrics.Metric = append(metrics.Metric, "Load1Min")
	metrics.Metric = append(metrics.Metric, "Load5Min")
	metrics.Metric = append(metrics.Metric, "Load15Min")

	return metrics, nil
}

// Retrieves metrics that apply to the instance
func (self *Instance) GetAllMetrics() (*model.Metrics, error) {
	huddle := self.huddle

	jujuUnitName := self.jujuPrefix + "metrics"

	es := huddle.SharedServices["jx-elasticsearch"]
	if es == nil {
		return nil, rs.ErrNotFound()
	}

	// TODO: Inject
	// TODO: Use an ES client that isn't a singleton
	elasticgo_api.Domain = es.PublicAddress
	elasticgo_api.Port = "9200"

	// TODO: We need to make sure that most fields are _not_ analyzed
	// That is why we have match below, not term
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]string{"Hostname": jujuUnitName},
		},
	}

	//	query := map[string]interface{}{
	//		"query": map[string]interface{}{
	//			"match_all": map[string]string {},
	//		},
	//	}

	args := map[string]interface{}{}
	args["size"] = 1000

	response, err := elastigo_core.SearchRequest("_all", "message", args, query)
	if err != nil {
		log.Warn("Error searching elasticsearch", err)
		return nil, fmt.Errorf("Error searching elasticsearch")
	}

	metrics := &model.Metrics{}
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

// Retrieves a specific metric-dataset for the instance
func (self *Instance) GetMetricValues(key string) (*model.MetricDataset, error) {
	state, err := self.getState0()
	if err != nil {
		return nil, err
	}

	primaryServiceId := self.primaryServiceId

	jujuUnitNames := []string{}

	units := state.Units[primaryServiceId]
	for jujuUnitName, _ := range units {
		unitId := juju.ParseUnit(jujuUnitName)
		metricUnit := self.jujuPrefix + "metrics" + "/" + unitId

		jujuUnitNames = append(jujuUnitNames, metricUnit)
	}

	log.Debug("Searching with names: %v", jujuUnitNames)

	return self.readMetrics(jujuUnitNames, key)
}
