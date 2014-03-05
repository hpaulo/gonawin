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

package tournaments

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"appengine"

	"github.com/santiaago/purple-wing/helpers"
	"github.com/santiaago/purple-wing/helpers/handlers"
	"github.com/santiaago/purple-wing/helpers/log"
	templateshlp "github.com/santiaago/purple-wing/helpers/templates"

	mdl "github.com/santiaago/purple-wing/models"
)

type MatchJson struct {
	IdNumber   int64
	Date       time.Time
	Team1      string
	Team2      string
	Location   string
	Result1    int64
	Result2    int64
	HasPredict bool
	Predict    string
}

// Json tournament Matches handler
// use this handler to get the matches of a tournament.
// use the filter parameter to specify the matches you want:
// if filter is equal to 'first' you wil get matches of the first phase of the tournament.
// if filter is equal to 'second' you will get the matches of the second  phase of the tournament.
func MatchesJson(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	c := appengine.NewContext(r)
	desc := "Tournament Matches Handler:"

	if r.Method == "GET" {
		tournamentId, err := handlers.PermalinkID(r, c, 3)
		if err != nil {
			log.Errorf(c, "%s error extracting permalink err:%v", desc, err)
			return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}

		var t *mdl.Tournament
		t, err = mdl.TournamentById(c, tournamentId)
		if err != nil {
			log.Errorf(c, "%s tournament with id:%v was not found %v", desc, tournamentId, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}

		filter := r.FormValue("filter")
		// if wrong data we set groupby to "day"
		if filter != "first" && filter != "second" {
			filter = "first"
		}

		log.Infof(c, "%s ready to build days array", desc)
		matchesJson := buildMatchesFromTournament(c, t, u)

		if filter == "first" {
			matchesJson = matchesJson[1:49]
		} else if filter == "second" {
			matchesJson = matchesJson[48:64]
		}
		data := struct {
			Matches []MatchJson
		}{
			matchesJson,
		}

		return templateshlp.RenderJson(w, c, data)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// Update Match handler.
// Update match of tournament with results information.
// from parameter 'result' with format 'result1 result2' the match information is updated accordingly.
func UpdateMatchResultJson(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	c := appengine.NewContext(r)
	desc := "Tournament Update Match Result Handler:"

	if r.Method == "POST" {
		tournamentId, err := handlers.PermalinkID(r, c, 3)

		if err != nil {
			log.Errorf(c, "%s error extracting permalink err:%v", desc, err)
			return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}
		var tournament *mdl.Tournament
		tournament, err = mdl.TournamentById(c, tournamentId)
		if err != nil {
			log.Errorf(c, "%s tournament with id:%v was not found %v", desc, tournamentId, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}

		matchIdNumber, err2 := handlers.PermalinkID(r, c, 5)
		if err2 != nil {
			log.Errorf(c, "%s error extracting permalink err:%v", desc, err2)
			return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeMatchCannotUpdate)}
		}

		match := mdl.GetMatchByIdNumber(c, *tournament, matchIdNumber)
		if match == nil {
			log.Errorf(c, "%s unable to get match with id number :%v", desc, matchIdNumber)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchNotFoundCannotUpdate)}
		}

		result := r.FormValue("result")
		// is result well formated?
		results := strings.Split(result, " ")
		r1 := 0
		r2 := 0
		if len(results) != 2 {
			log.Errorf(c, "%s unable to get results, lenght not right: %v", desc, results)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchCannotUpdate)}
		}
		if r1, err = strconv.Atoi(results[0]); err != nil {
			log.Errorf(c, "%s unable to get results, error: %v not number 1", desc, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchCannotUpdate)}
		}
		if r2, err = strconv.Atoi(results[1]); err != nil {
			log.Errorf(c, "%s unable to get results, error: %v not number 2", desc, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchCannotUpdate)}
		}

		if err = mdl.SetResult(c, match, int64(r1), int64(r2), tournament); err != nil {
			log.Errorf(c, "%s unable to set result for match with id:%v error: %v", desc, match.IdNumber, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchCannotUpdate)}

		}

		// return the updated match
		var mjson MatchJson
		mjson.IdNumber = match.IdNumber
		mjson.Date = match.Date
		rule := strings.Split(match.Rule, " ")

		mapIdTeams := mdl.MapOfIdTeams(c, tournament)

		if len(rule) > 1 {
			mjson.Team1 = rule[0]
			mjson.Team2 = rule[1]
		} else {
			mjson.Team1 = mapIdTeams[match.TeamId1]
			mjson.Team2 = mapIdTeams[match.TeamId2]
		}
		mjson.Location = match.Location

		mjson.Result1 = match.Result1
		mjson.Result2 = match.Result2

		return templateshlp.RenderJson(w, c, mjson)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// From a tournament entity return an array of MatchJson data structure.
// second phase matches will have the specific rules in there team names
func buildMatchesFromTournament(c appengine.Context, t *mdl.Tournament, u *mdl.User) []MatchJson {

	matches := mdl.Matches(c, t.Matches1stStage)
	matches2ndPhase := mdl.Matches(c, t.Matches2ndStage)

	var predicts mdl.Predicts
	predicts = mdl.PredictsByIds(c, u.PredictIds)

	mapIdTeams := mdl.MapOfIdTeams(c, t)

	matchesJson := make([]MatchJson, len(matches))
	for i, m := range matches {
		matchesJson[i].IdNumber = m.IdNumber
		matchesJson[i].Date = m.Date
		matchesJson[i].Team1 = mapIdTeams[m.TeamId1]
		matchesJson[i].Team2 = mapIdTeams[m.TeamId2]
		matchesJson[i].Location = m.Location
		matchesJson[i].Result1 = m.Result1
		matchesJson[i].Result2 = m.Result2
		if hasMatch, j := predicts.ContainsMatchId(m.Id); hasMatch == true {
			matchesJson[i].HasPredict = true
			matchesJson[i].Predict = fmt.Sprintf("%v - %v", predicts[j].Result1, predicts[j].Result2)
		} else {
			matchesJson[i].HasPredict = false
		}
	}

	// append 2nd round to first one
	for _, m := range matches2ndPhase {
		var matchJson2ndPhase MatchJson
		matchJson2ndPhase.IdNumber = m.IdNumber
		matchJson2ndPhase.Date = m.Date
		rule := strings.Split(m.Rule, " ")
		if len(rule) == 2 {
			matchJson2ndPhase.Team1 = rule[0]
			matchJson2ndPhase.Team2 = rule[1]
		} else {
			matchJson2ndPhase.Team1 = mapIdTeams[m.TeamId1]
			matchJson2ndPhase.Team2 = mapIdTeams[m.TeamId2]
		}

		matchJson2ndPhase.Location = m.Location
		matchJson2ndPhase.Result1 = m.Result1
		matchJson2ndPhase.Result2 = m.Result2

		if hasMatch, j := predicts.ContainsMatchId(m.Id); hasMatch == true {
			matchJson2ndPhase.HasPredict = true
			matchJson2ndPhase.Predict = fmt.Sprintf("%v - %v", predicts[j].Result1, predicts[j].Result2)
		} else {
			matchJson2ndPhase.HasPredict = false
		}

		// append second round results
		matchesJson = append(matchesJson, matchJson2ndPhase)
	}

	return matchesJson
}