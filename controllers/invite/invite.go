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

// Package invite provides the JSON handlers to send invitations to gonawin app.
package invite

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"appengine"
	"appengine/taskqueue"

	"github.com/santiaago/gonawin/helpers"
	"github.com/santiaago/gonawin/helpers/log"
	templateshlp "github.com/santiaago/gonawin/helpers/templates"

	mdl "github.com/santiaago/gonawin/models"
)

const inviteMessage = `
Hi there,
Join us at gonawin.

You will be able to join your friends and compete with them by predicting the results of your favorite sports events!

Sign in here: %s

Have fun,
Your friends @ Gonawin


`

// Invite handler, use it to invite users to use gonawin.
func Invite(w http.ResponseWriter, r *http.Request, u *mdl.User) error {
	if r.Method != "POST" {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	desc := "invite handler:"
	c := appengine.NewContext(r)

	emailsList := r.FormValue("emails")

	if len(emailsList) <= 0 {
		return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeInviteNoEmailAddr)}
	}
	splitemails := strings.Split(emailsList, ",")

	// remove leading and trailing spaces from each email.
	emails := make([]string, 0)
	for _, e := range splitemails {
		emails = append(emails, strings.Trim(e, " "))
	}

	// validate emails
	if !helpers.AreEmailsValid(emails) {
		log.Errorf(c, "%s emails not valid dude!", desc, emails)
		return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeInviteEmailsInvalid)}
	}

	currenturl := fmt.Sprintf("http://%s/#", r.Host)
	body := fmt.Sprintf(inviteMessage, currenturl)

	bname, errname := json.Marshal(u.Name)
	if errname != nil {
		log.Errorf(c, "%s Error marshaling", desc, errname)
	}

	bbody, errbody := json.Marshal(body)
	if errbody != nil {
		log.Errorf(c, "%s Error marshaling", desc, errbody)
	}

	for _, email := range emails {

		bemail, errm := json.Marshal(email)
		if errm != nil {
			log.Errorf(c, "%s Error marshaling", desc, errm)
		}

		task := taskqueue.NewPOSTTask("/a/invite/", url.Values{
			"email": []string{string(bemail)},
			"name":  []string{string(bname)},
			"body":  []string{string(bbody)},
		})
		if _, err := taskqueue.Add(c, task, ""); err != nil {
			log.Errorf(c, "%s unable to add task to taskqueue.", desc)
			return err
		} else {
			log.Infof(c, "%s add task to taskqueue successfully", desc)
		}
	}
	msg := fmt.Sprintf("Email invitations have been successfully sent.")
	data := struct {
		MessageInfo string `json:",omitempty"`
	}{
		msg,
	}

	return templateshlp.RenderJson(w, c, data)
}
