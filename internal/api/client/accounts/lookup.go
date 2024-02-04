// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountLookupGETHandler swagger:operation GET /api/v1/accounts/lookup accountLookupGet
//
// Quickly lookup a username to see if it is available, skipping WebFinger resolution.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: acct
//		type: string
//		description: The username or Webfinger address to lookup.
//		in: query
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			name: lookup result
//			description: Result of the lookup.
//			schema:
//				"$ref": "#/definitions/account"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AccountLookupGETHandler(c *gin.Context) {
	var authed *oauth.Auth
	var err error
	if config.GetInstanceExposeLookup() {
		// If the lookup endpoint is allowed to be exposed, still check if we
		// can extract various authentication properties, but don't require them.
		authed, err = oauth.Authed(c, false, false, false, false)
	} else {
		authed, err = oauth.Authed(c, true, true, true, true)
	}

	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	query, errWithCode := apiutil.ParseSearchLookup(c.Query(apiutil.SearchLookupKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	account, errWithCode := m.processor.Search().Lookup(c.Request.Context(), authed.Account, query)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, account)
}
