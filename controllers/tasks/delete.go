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

package tasks

import (
	"encoding/json"
	"errors"
	"net/http"

	"appengine"

	"github.com/taironas/gonawin/helpers"
	"github.com/taironas/gonawin/helpers/log"
	mdl "github.com/taironas/gonawin/models"
)

// DeleteUserPredicts task handler, use it to delete the predictions of a given user.
//
func DeleteUserPredicts(w http.ResponseWriter, r *http.Request) error {

	c := appengine.NewContext(r)
	desc := "Task queue - DeleteUserPredicts Handler:"

	if r.Method != "POST" {
		return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
	}

	predictIdsBlob := []byte(r.FormValue("predict_ids"))

	var predictIds []int64
	if err := json.Unmarshal(predictIdsBlob, &predictIds); err != nil {
		log.Errorf(c, "%s unable to extract predictIds from data, %v", desc, err)
	}

	if err := mdl.DestroyPredicts(c, predictIds); err != nil {
		log.Errorf(c, "%s predicts have not been deleted. %v", desc, err)
	}

	return nil
}
