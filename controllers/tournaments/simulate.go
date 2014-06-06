/*
 * Copyright (c) 2014 Santiago Arias | Remy Jourde
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
	"math/rand"
	"net/http"

	"appengine"

	"github.com/santiaago/gonawin/helpers"
	"github.com/santiaago/gonawin/helpers/handlers"
	"github.com/santiaago/gonawin/helpers/log"
	templateshlp "github.com/santiaago/gonawin/helpers/templates"
	mdl "github.com/santiaago/gonawin/models"
)

// Simulate the scores of a phase in a tournament.
//    POST /j/tournaments/[0-9]+/matches/simulate?phase=:phaseName
func SimulateMatches(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	c := appengine.NewContext(r)

	if r.Method == "POST" {
		tournamentId, err := handlers.PermalinkID(r, c, 3)

		if err != nil {
			log.Errorf(c, "Tournament Simulate Matches Handler: error extracting permalink err:%v", err)
			return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}
		var t *mdl.Tournament
		t, err = mdl.TournamentById(c, tournamentId)
		if err != nil {
			log.Errorf(c, "Tournament Simulate Matches Handler: tournament with id:%v was not found %v", tournamentId, err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeTournamentNotFound)}
		}

		phase := r.FormValue("phase")
		allMatches := mdl.GetAllMatchesFromTournament(c, t)
		phases := mdl.MatchesGroupByPhase(allMatches)

		mapIdTeams := mdl.MapOfIdTeams(c, t)
		phaseId := -1
		var results1 []int64
		var results2 []int64
		var matches []*mdl.Tmatch
		for i, ph := range phases {
			if ph.Name != phase {
				continue
			}
			phaseId = i
			for _, d := range ph.Days {
				for j, m := range d.Matches {
					// simulate match here (call set results)
					r1 := int64(rand.Intn(5))
					r2 := int64(rand.Intn(5))
					results1 = append(results1, r1)
					results2 = append(results2, r2)
					matches = append(matches, &d.Matches[j])
					log.Infof(c, "Tournament Simulate Matches: Match#%v: %v - %v | %v - %v", m.Id, mapIdTeams[m.TeamId1], mapIdTeams[m.TeamId2], r1, r2)
				}
			}
			// phase done we and not break
			break
		}
		if err = mdl.SetResults(c, matches, results1, results2, t); err != nil {
			log.Errorf(c, "Tournament Simulate Matches: unable to set result for matches error: %v", err)
			return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeMatchesCannotUpdate)}
		}

		// publish activities:
		for i, match := range matches {
			object := mdl.ActivityEntity{Id: match.TeamId1, Type: "tteam", DisplayName: mapIdTeams[match.TeamId1]}
			target := mdl.ActivityEntity{Id: match.TeamId2, Type: "tteam", DisplayName: mapIdTeams[match.TeamId2]}
			verb := ""
			if results1[i] > results2[i] {
				verb = fmt.Sprintf("won %d-%d against", results1[i], results2[i])
			} else if results1[i] < results2[i] {
				verb = fmt.Sprintf("lost %d-%d against", results1[i], results2[i])
			} else {
				verb = fmt.Sprintf("tied %d-%d against", results1[i], results2[i])
			}
			t.Publish(c, "match", verb, object, target)
		}

		if phaseId >= 0 {
			// only return update phase
			matchesJson := buildMatchesFromTournament(c, t, u)
			phasesJson := matchesGroupByPhase(matchesJson)

			data := struct {
				Phase PhaseJson
			}{
				phasesJson[phaseId],
			}
			return templateshlp.RenderJson(w, c, data)
		}
		return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeInternal)}
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}
