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

package teams

import (
	"errors"
	"fmt"
	"net/http"

	"appengine"

	"github.com/santiaago/gonawin/helpers"
	"github.com/santiaago/gonawin/helpers/log"
	templateshlp "github.com/santiaago/gonawin/helpers/templates"

	mdl "github.com/santiaago/gonawin/models"
)

// AddAdmin handler, use it to add an admin to a team.
//
// Use this handler to add a user as admin of current team.
//	GET	/j/teams/:teamId/admin/add/:userId
//
func AddAdmin(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	if r.Method != "POST" {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	c := appengine.NewContext(r)
	desc := "Team add admin Handler:"
	rc := requestContext{c, desc, r}

	var team *mdl.Team
	var err error
	team, err = rc.team()
	if err != nil {
		return err
	}

	var newAdmin *mdl.User
	newAdmin, err = rc.user()
	if err != nil {
		return err
	}

	if err = team.AddAdmin(c, newAdmin.Id); err != nil {
		log.Errorf(c, "%s error on AddAdmin to team: %v", desc, err)
		return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeInternal)}
	}

	var tJson mdl.TeamJson
	fieldsToKeep := []string{"Id", "Name", "AdminIds", "Private"}
	helpers.InitPointerStructure(team, &tJson, fieldsToKeep)

	msg := fmt.Sprintf("You added %s as admin of team %s.", newAdmin.Name, team.Name)
	data := struct {
		MessageInfo string `json:",omitempty"`
		Team        mdl.TeamJson
	}{
		msg,
		tJson,
	}

	return templateshlp.RenderJson(w, c, data)
}

// RemoveAdmin handler, use it to remove an admin from a team.
//
// Use this handler to remove a user as admin of the current team.
//	GET	/j/teams/:teamId/admin/remove/:userId
//
func RemoveAdmin(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	if r.Method != "POST" {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	c := appengine.NewContext(r)
	desc := "Team remove admin Handler:"
	rc := requestContext{c, desc, r}

	var team *mdl.Team
	var err error
	team, err = rc.team()
	if err != nil {
		return err
	}

	var oldAdmin *mdl.User
	oldAdmin, err = rc.user()
	if err != nil {
		return err
	}

	if err = team.RemoveAdmin(c, oldAdmin.Id); err != nil {
		log.Errorf(c, "%s error on RemoveAdmin to team: %v.", desc, err)
		return &helpers.InternalServerError{Err: err}
	}

	var tJson mdl.TeamJson
	fieldsToKeep := []string{"Id", "Name", "AdminIds", "Private"}
	helpers.InitPointerStructure(team, &tJson, fieldsToKeep)

	msg := fmt.Sprintf("You removed %s as admin of team %s.", oldAdmin.Name, team.Name)
	data := struct {
		MessageInfo string `json:",omitempty"`
		Team        mdl.TeamJson
	}{
		msg,
		tJson,
	}
	return templateshlp.RenderJson(w, c, data)
}
