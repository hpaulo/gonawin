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

package users

import (
	"errors"
	"net/http"
	"strconv"

	"appengine"

	"github.com/taironas/route"

	"github.com/santiaago/gonawin/helpers"
	"github.com/santiaago/gonawin/helpers/log"
	templateshlp "github.com/santiaago/gonawin/helpers/templates"
	mdl "github.com/santiaago/gonawin/models"
)

// User score user handler.
func Score(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	if r.Method != "GET" {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	desc := "User Score Handler:"
	c := appengine.NewContext(r)

	// get user id
	strUserId, err := route.Context.Get(r, "userId")
	if err != nil {
		log.Errorf(c, "%s error getting user id, err:%v", desc, err)
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeUserNotFound)}
	}

	var userId int64
	userId, err = strconv.ParseInt(strUserId, 0, 64)
	if err != nil {
		log.Errorf(c, "%s error converting user id from string to int64, err:%v", desc, err)
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeUserNotFound)}
	}

	var user *mdl.User
	user, err = mdl.UserById(c, userId)
	if err != nil {
		log.Errorf(c, "%s user not found", desc)
		return &helpers.NotFound{Err: errors.New(helpers.ErrorCodeUserNotFound)}
	}

	//scores := user.Scores(c)
	scores := user.TournamentsScores(c)
	// data
	data := struct {
		Scores []*mdl.ScoreOverall
	}{
		scores,
	}

	return templateshlp.RenderJson(w, c, data)
}
