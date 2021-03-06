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
	"fmt"
	"net/http"

	"appengine"

	"github.com/taironas/gonawin/helpers"
	"github.com/taironas/gonawin/helpers/log"
	templateshlp "github.com/taironas/gonawin/helpers/templates"
	mdl "github.com/taironas/gonawin/models"
)

type searchUserViewModel struct {
	Id       int64 `json:"Id"`
	Username string
	Alias    string
	Score    int64
	ImageURL string
}

// Search handler returns the result of a user search in a JSON format.
// It uses parameter 'q' to make the query.
//
//	GET	/j/user/search/			Search for all users respecting the query "q"
//
func Search(w http.ResponseWriter, r *http.Request, u *mdl.User) error {

	keywords := r.FormValue("q")
	if r.Method != "GET" || len(keywords) == 0 {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	c := appengine.NewContext(r)
	desc := "User Search Handler:"

	words := helpers.SetOfStrings(keywords)

	var ids []int64
	var err error
	if ids, err = mdl.GetUserInvertedIndexes(c, words); err != nil {
		return unableToPerformSearch(c, w, desc, err)
	}

	result := mdl.UserScore(c, keywords, ids)

	var users []*mdl.User
	if users, err = mdl.UsersByIds(c, result); len(users) == 0 || err != nil {
		return notFound(c, w, keywords)
	}

	uvm := buildSearchUserViewModel(users)

	data := struct {
		Users []searchUserViewModel `json:",omitempty"`
	}{
		uvm,
	}
	return templateshlp.RenderJSON(w, c, data)
}

func buildSearchUserViewModel(users []*mdl.User) []searchUserViewModel {
	uvm := make([]searchUserViewModel, len(users))
	for i, u := range users {
		uvm[i].Id = u.Id
		uvm[i].Username = u.Username
		uvm[i].Alias = u.Alias
		uvm[i].Score = u.Score
		uvm[i].ImageURL = helpers.UserImageURL(u.Name, u.Id)
	}
	return uvm
}

func notFound(c appengine.Context, w http.ResponseWriter, keywords string) error {
	msg := fmt.Sprintf("Oops! Your search - %s - did not match any %s.", keywords, "user")
	data := struct {
		MessageInfo string `json:",omitempty"`
	}{
		msg,
	}
	return templateshlp.RenderJSON(w, c, data)
}

func unableToPerformSearch(c appengine.Context, w http.ResponseWriter, desc string, err error) error {
	log.Errorf(c, "%s users.Index, error occurred when getting indexes of words: %v", desc, err)
	data := struct {
		MessageDanger string `json:",omitempty"`
	}{
		"Oops! something went wrong, we are unable to perform search query.",
	}
	return templateshlp.RenderJSON(w, c, data)
}
