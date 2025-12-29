package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	m "mc.data/models"
)

func (pg *Postgres) GetScenarios(ctx context.Context) ([]*m.Scenerio, error) {
	scenarioQuery := `
		SELECT
			id,
			name,
			floated_weight,
			created_at,
			updated_at
		FROM scenario_configuration
		WHERE deleted_at IS NULL
	`

	scenarios, err := Query[m.ScenarioConfiguration](ctx, pg, scenarioQuery, pgx.NamedArgs{})
	if err != nil {
		return nil, fmt.Errorf("unable to get scenarios: %w", err)
	}

	componentQuery := `
		SELECT
			configuration_id
			asset_id,
			weight
		FROM scenario_configuration_component scc
		JOIN scenario_configuration sc ON scc.configuration_id = sc.id
		WHERE sc.deleted_at IS NULL
	`

	components, err := Query[m.ScenarioConfigurationComponent](ctx, pg, componentQuery, pgx.NamedArgs{})
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario components: %w", err)
	}

	scenarioComponentsLookup := make(map[int32][]m.ScenarioConfigurationComponent)
	for _, v := range components {
		if scenarioComponentsLookup[v.ConfigurationId] == nil {
			scenarioComponentsLookup[v.ConfigurationId] = []m.ScenarioConfigurationComponent{}
		}
		scenarioComponentsLookup[v.ConfigurationId] = append(scenarioComponentsLookup[v.ConfigurationId], *v)
	}

	res := make([]*m.Scenerio, len(scenarios))
	for _, v := range scenarios {
		res = append(res, &m.Scenerio{
			ScenarioConfiguration: *v,
			Components:            scenarioComponentsLookup[v.Id],
		})
	}

	return res, nil
}

func (pg *Postgres) InsertNewScenario(ns m.NewScenario) error {
	// do this in a transaction

	// insert the one line item into the scenario page, return the scenario id
	// bulk insert the configuration lines with the returned scenario id
	return nil
}

func (pg *Postgres) UpdateExistingScenerio() error {
	return nil
}
