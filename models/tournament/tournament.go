/*
 * Copyright (c) 2013 Santiago Arias | Remy Jourde | Carlos Bernal
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package tournament

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/santiaago/purple-wing/helpers"
	"github.com/santiaago/purple-wing/helpers/log"
	tournamentinvidmdl "github.com/santiaago/purple-wing/models/tournamentInvertedIndex"
	tournamentrelmdl "github.com/santiaago/purple-wing/models/tournamentrel"
	tournamentteamrelmdl "github.com/santiaago/purple-wing/models/tournamentteamrel"
)

type Tournament struct {
	Id              int64
	KeyName         string
	Name            string
	Description     string
	Start           time.Time
	End             time.Time
	AdminId         int64
	Created         time.Time
	GroupIds        []int64
	Matches1stStage []int64
	Matches2ndStage []int64
}

type TournamentJson struct {
	Id              *int64     `json:",omitempty"`
	KeyName         *string    `json:",omitempty"`
	Name            *string    `json:",omitempty"`
	Description     *string    `json:",omitempty"`
	Start           *time.Time `json:",omitempty"`
	End             *time.Time `json:",omitempty"`
	AdminId         *int64     `json:",omitempty"`
	Created         *time.Time `json:",omitempty"`
	GroupIds        *[]int64   `json:",omitempty"`
	Matches1stStage *[]int64   `json:",omitempty"`
	Matches2ndStage *[]int64   `json:",omitempty"`
}

// Create tournament entity given a name and description.
func Create(c appengine.Context, name string, description string, start time.Time, end time.Time, adminId int64) (*Tournament, error) {

	tournamentID, _, err := datastore.AllocateIDs(c, "Tournament", nil, 1)
	if err != nil {
		return nil, err
	}

	key := datastore.NewKey(c, "Tournament", "", tournamentID, nil)

	// empty groups and tournaments for now
	groupIds := make([]int64, 0)
	matches1stStageIds := make([]int64, 0)
	matches2ndStageIds := make([]int64, 0)

	tournament := &Tournament{tournamentID, helpers.TrimLower(name), name, description, start, end, adminId, time.Now(), groupIds, matches1stStageIds, matches2ndStageIds}

	_, err = datastore.Put(c, key, tournament)
	if err != nil {
		return nil, err
	}

	tournamentinvidmdl.Add(c, helpers.TrimLower(name), tournamentID)
	return tournament, nil
}

// Destroy a tournament entity given a tournament id.
func Destroy(c appengine.Context, tournamentId int64) error {

	if tournament, err := ById(c, tournamentId); err != nil {
		return errors.New(fmt.Sprintf("Cannot find tournament with tournamentId=%d", tournamentId))
	} else {
		key := datastore.NewKey(c, "Tournament", "", tournament.Id, nil)
		return datastore.Delete(c, key)
	}
}

// Find all tournaments in datastore given a filter and value.
func Find(c appengine.Context, filter string, value interface{}) []*Tournament {

	q := datastore.NewQuery("Tournament").Filter(filter+" =", value)
	var tournaments []*Tournament
	if _, err := q.GetAll(c, &tournaments); err == nil {
		return tournaments
	} else {
		log.Errorf(c, " Tournament.Find, error occurred during GetAll: %v", err)
		return nil
	}
}

// Get a pointer to a tournament given a tournament id.
func ById(c appengine.Context, id int64) (*Tournament, error) {

	var t Tournament
	key := datastore.NewKey(c, "Tournament", "", id, nil)
	if err := datastore.Get(c, key, &t); err != nil {
		log.Errorf(c, " tournament not found : %v", err)
		return &t, err
	}
	return &t, nil
}

// Get a pointer to a tournament key given a tournament id.
func KeyById(c appengine.Context, id int64) *datastore.Key {

	key := datastore.NewKey(c, "Tournament", "", id, nil)
	return key
}

// Update a tournament given a tournament id and a tournament pointer.
func Update(c appengine.Context, id int64, t *Tournament) error {

	// update key name
	t.KeyName = helpers.TrimLower(t.Name)
	k := KeyById(c, id)
	oldTournament := new(Tournament)
	if err := datastore.Get(c, k, oldTournament); err == nil {
		if _, err = datastore.Put(c, k, t); err != nil {
			return err
		}
		// use name with trim lower as tournament inverted index stores lower key names.
		tournamentinvidmdl.Update(c, oldTournament.KeyName, t.KeyName, id)
	}
	return nil
}

// Find all tournaments in the datastore.
func FindAll(c appengine.Context) []*Tournament {

	q := datastore.NewQuery("Tournament")
	var tournaments []*Tournament
	if _, err := q.GetAll(c, &tournaments); err != nil {
		log.Errorf(c, " Tournament.FindAll, error occurred during GetAll call: %v", err)
	}
	return tournaments
}

// Find all tournaments with respect to array of ids.
func ByIds(c appengine.Context, ids []int64) []*Tournament {

	var tournaments []*Tournament
	for _, id := range ids {
		if tournament, err := ById(c, id); err == nil {
			tournaments = append(tournaments, tournament)
		} else {
			log.Errorf(c, " Tournament.ByIds, error occurred during ByIds call: %v", err)
		}
	}
	return tournaments
}

// Checks if a user has joined a tournament.
func Joined(c appengine.Context, tournamentId int64, userId int64) bool {
	tournamentRel := tournamentrelmdl.FindByTournamentIdAndUserId(c, tournamentId, userId)
	return tournamentRel != nil
}

// Makes a user join a tournament.
func Join(c appengine.Context, tournamentId int64, userId int64) error {
	if tournamentRel, err := tournamentrelmdl.Create(c, tournamentId, userId); tournamentRel == nil {
		return errors.New(fmt.Sprintf(" Tournament.Join, error during tournament relationship creation: %v", err))
	}

	return nil
}

// Makes a user leave a tournament.
// Todo: should we check that user is indeed a member of the tournament?
func Leave(c appengine.Context, tournamentId int64, userId int64) error {
	return tournamentrelmdl.Destroy(c, tournamentId, userId)
}

// Checks if user is admin of given tournament.
func IsTournamentAdmin(c appengine.Context, tournamentId int64, userId int64) bool {
	if tournament, err := ById(c, tournamentId); err == nil {
		return tournament.AdminId == userId
	}

	return false
}

// Check if a Team has joined the tournament.
func TeamJoined(c appengine.Context, tournamentId int64, teamId int64) bool {
	tournamentteamRel := tournamentteamrelmdl.FindByTournamentIdAndTeamId(c, tournamentId, teamId)
	return tournamentteamRel != nil
}

// Team joins the Tournament.
func TeamJoin(c appengine.Context, tournamentId int64, teamId int64) error {
	if tournamentteamRel, err := tournamentteamrelmdl.Create(c, tournamentId, teamId); tournamentteamRel == nil {
		return errors.New(fmt.Sprintf(" Tournament.TeamJoin, error during tournament team relationship creation: %v", err))
	}

	return nil
}

// Team leaves the Tournament.
func TeamLeave(c appengine.Context, tournamentId int64, teamId int64) error {
	return tournamentteamrelmdl.Destroy(c, tournamentId, teamId)
}

// Get the frequency of given word with respect to tournament id.
func GetWordFrequencyForTournament(c appengine.Context, id int64, word string) int64 {

	if tournaments := Find(c, "Id", id); tournaments != nil {
		return helpers.CountTerm(strings.Split(tournaments[0].KeyName, " "), word)
	}
	return 0
}

// Reset tournament values: Points, GoalsF, GoalsA to zero.
func Reset(c appengine.Context, t *Tournament) error {
	groups := Groups(c, t.GroupIds)
	for _, g := range groups {
		g.Points = make([]int64, len(g.Teams))
		g.GoalsF = make([]int64, len(g.Teams))
		g.GoalsA = make([]int64, len(g.Teams))
		for _, m := range g.Matches {
			m.Result1 = 0
			m.Result2 = 0
			if err := UpdateMatch(c, &m); err != nil {
				return err
			}
		}
		if err := UpdateGroup(c, g); err != nil {
			return err
		}
	}
	// reset all matches rules
	return nil
}
