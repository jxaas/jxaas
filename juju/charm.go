package juju

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/charm"
	"launchpad.net/juju-core/state/api"
)

// addCharmViaAPI calls the appropriate client API calls to add the
// given charm URL to state. Also displays the charm URL of the added
// charm on stdout.
func addCharmViaAPI(client *api.Client, curl *charm.URL, repo charm.Repository) (*charm.URL, error) {
	if curl.Revision < 0 {
		latest, err := charm.Latest(repo, curl)
		if err != nil {
			log.Info("Error find latest version for: %v", curl.String(), err)
			return nil, err
		}
		curl = curl.WithRevision(latest)
	}
	switch curl.Schema {
	case "local":
		ch, err := repo.Get(curl)
		if err != nil {
			return nil, err
		}
		stateCurl, err := client.AddLocalCharm(curl, ch)
		if err != nil {
			return nil, err
		}
		curl = stateCurl
	case "cs":
		err := client.AddCharm(curl)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported charm URL schema: %q", curl.Schema)
	}
	log.Info("Added charm %q to the environment.", curl)
	return curl, nil
}

// resolveCharmURL returns a resolved charm URL, given a charm location string.
// If the series is not resolved, the environment default-series is used, or if
// not set, the series is resolved with the state server.
func resolveCharmURL(client *api.Client, url string, defaultSeries string) (*charm.URL, error) {
	ref, series, err := charm.ParseReference(url)
	if err != nil {
		return nil, err
	}
	// If series is not set, use configured default series
	if series == "" {
		series = defaultSeries
	}
	// Otherwise, look up the best supported series for this charm
	if series == "" {
		return client.ResolveCharm(ref)
	}
	return &charm.URL{Reference: ref, Series: series}, nil
}

func getCharmInfo(client *api.Client, charmName string, localRepoPath string, defaultSeries string) (*api.CharmInfo, error) {
	curl, err := resolveCharmURL(client, charmName, defaultSeries)
	if err != nil {
		return nil, err
	}

	repo, err := charm.InferRepository(curl.Reference, localRepoPath)
	if err != nil {
		return nil, err
	}

	//	repo = config.SpecializeCharmRepo(repo, defaultSeries)

	curl, err = addCharmViaAPI(client, curl, repo)
	if err != nil {
		return nil, err
	}

	charmInfo, err := client.CharmInfo(curl.String())
	if err != nil {
		log.Info("Error getting charm info for: %v", curl.String(), err)
		return nil, err
	}
	return charmInfo, nil
}
