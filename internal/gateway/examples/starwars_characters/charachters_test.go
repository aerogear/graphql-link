package starwars_characters_test

import (
	"encoding/json"
	"testing"

	"github.com/chirino/graphql-gw/internal/gateway/examples/starwars_characters"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {

	engine := starwars_characters.New()
	result := json.RawMessage{}
	engine.Exec(nil, &result, starwars_characters.ComplexStarWarsCharacterQuery, "episode", "JEDI",
		"withoutFriends", true,
		"withFriends", false)
	assert.JSONEq(t, starwars_characters.ComplexStarWarsCharacterQueryResult, string(result))
}
