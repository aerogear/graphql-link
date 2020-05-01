package starwars_characters

import (
	"encoding/base64"
	"fmt"

	"github.com/chirino/graphql/resolvers"
)

type characterMethods interface {
	ToHuman() (human, bool)
	ToDroid() (droid, bool)
}

func (s character) ToHuman() (human, bool) { return s.self.ToHuman() }
func (s human) ToHuman() (human, bool)     { return s, true }
func (s droid) ToHuman() (human, bool)     { return human{}, false }

func (s character) ToDroid() (droid, bool) { return s.self.ToDroid() }
func (s human) ToDroid() (droid, bool)     { return droid{}, false }
func (s droid) ToDroid() (droid, bool)     { return s, true }

type character struct {
	self      characterMethods
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	FriendIds []string `json:"friendIds"`
	AppearsIn []string `json:"appearsIn"`
}

func (s character) Friends(ctx resolvers.ExecutionContext) []character {
	root := ctx.GetRoot().(root)
	return root.characters(s.FriendIds)
}

func (s character) FriendsConnection(ctx resolvers.ExecutionContext, args struct {
	First *int32
	After *string
}) (*friendsConnection, error) {
	root := ctx.GetRoot().(root)
	return root.friendsConnection(s.FriendIds, args.First, args.After)
}

type droid struct {
	character
	PrimaryFunction string `json:"primaryFunction"`
}

type human struct {
	character
	Height    float64  `json:"height"`
	Mass      int      `json:"mass"`
	Starships []string `json:"starships"`
}

type friendsConnection struct {
	ids  []string
	from int
	to   int
}

//		totalCount: Int!
func (s *friendsConnection) TotalCount() int32 {
	return int32(len(s.ids))
}

//		"The edges for each of the character's friends."
//		edges: [FriendsEdge]
func (s *friendsConnection) Edges(ctx resolvers.ExecutionContext) *[]*friendsEdge {
	root := ctx.GetRoot().(root)
	l := make([]*friendsEdge, s.to-s.from)
	for i := range l {
		id := s.ids[s.from+i]
		l[i] = &friendsEdge{
			Cursor: encodeCursor(s.from + i),
			Node:   root.character(id),
		}
	}
	return &l
}

type friendsEdge struct {
	Cursor string     `json:"cursor"`
	Node   *character `json:"node"`
}

func encodeCursor(i int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor%d", i+1)))
}

//		"A list of the friends, as a convenience when edges are not needed."
//		friends: [Character]
func (s *friendsConnection) Friends(ctx resolvers.ExecutionContext) []character {
	root := ctx.GetRoot().(root)
	ids := s.ids[s.from:s.to]
	return root.characters(ids)
}

type pageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

//		"Information for paginating this connection"
//		pageInfo: PageInfo!
func (s *friendsConnection) PageInfo() pageInfo {
	return pageInfo{
		StartCursor: encodeCursor(s.from),
		EndCursor:   encodeCursor(s.to - 1),
		HasNextPage: s.to < len(s.ids),
	}
}
