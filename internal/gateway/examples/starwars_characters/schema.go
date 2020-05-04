package starwars_characters

var Schema = `
	schema {
		query: Query
	}
	"The query type, represents the entry points related to characters in the Starwars universe."
	type Query {
		hero(episode: Episode = NEWHOPE): Character
		search(text: String!): [SearchResult]!
		character(id: ID!): Character
		droid(id: ID!): Droid
		human(id: ID!): Human
	}
	"The id of a Starship"
	scalar Starship
	"The episodes in the Star Wars trilogy"
	enum Episode {
		"Star Wars Episode IV: A New Hope, released in 1977."
		NEWHOPE
		"Star Wars Episode V: The Empire Strikes Back, released in 1980."
		EMPIRE
		"Star Wars Episode VI: Return of the Jedi, released in 1983."
		JEDI
	}
	"A character from the Star Wars universe"
	interface Character {
		"The ID of the character"
		id: ID!
		"The name of the character"
		name: String!
		"The friends of the character, or an empty list if they have none"
		friends: [Character]
		"The friends of the character exposed as a connection with edges"
		friendsConnection(first: Int, after: ID): FriendsConnection!
		"The movies this character appears in"
		appearsIn: [Episode!]!
	}
	"Units of height"
	enum LengthUnit {
		"The standard unit around the world"
		METER
		"Primarily used in the United States"
		FOOT
	}
	"A humanoid creature from the Star Wars universe"
	type Human implements Character {
		"The ID of the human"
		id: ID!
		"What this human calls themselves"
		name: String!
		"Height in the preferred unit, default is meters"
		height(unit: LengthUnit = METER): Float!
		"mass in kilograms, or null if unknown"
		mass: Float
		"This human's friends, or an empty list if they have none"
		friends: [Character]
		"The friends of the human exposed as a connection with edges"
		friendsConnection(first: Int, after: ID): FriendsConnection!
		"The movies this human appears in"
		appearsIn: [Episode!]!
		"A list of starships this person has piloted, or an empty list if none"
		starships: [Starship]
	}
	"An autonomous mechanical character in the Star Wars universe"
	type Droid implements Character {
		"The ID of the droid"
		id: ID!
		"What others call this droid"
		name: String!
		"This droid's friends, or an empty list if they have none"
		friends: [Character]
		"The friends of the droid exposed as a connection with edges"
		friendsConnection(first: Int, after: ID): FriendsConnection!
		"The movies this droid appears in"
		appearsIn: [Episode!]!
		"This droid's primary function"
		primaryFunction: String
	}
	"A connection object for a character's friends"
	type FriendsConnection {
		"The total number of friends"
		totalCount: Int!
		"The edges for each of the character's friends."
		edges: [FriendsEdge]
		"A list of the friends, as a convenience when edges are not needed."
		friends: [Character]
		"Information for paginating this connection"
		pageInfo: PageInfo!
	}
	"An edge object for a character's friends"
	type FriendsEdge {
		"A cursor used for pagination"
		cursor: ID!
		"The character represented by this friendship edge"
		node: Character
	}
	"Information for paginating this connection"
	type PageInfo {
		startCursor: ID
		endCursor: ID
		hasNextPage: Boolean!
	}
	union SearchResult = Human | Droid
`
