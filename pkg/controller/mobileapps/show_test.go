// Copyright 2020 the Exposure Notifications Verification Server authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mobileapps_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/exposure-notifications-verification-server/internal/envstest"
	"github.com/google/exposure-notifications-verification-server/internal/project"
	"github.com/google/exposure-notifications-verification-server/pkg/controller"
	"github.com/google/exposure-notifications-verification-server/pkg/controller/mobileapps"
	"github.com/google/exposure-notifications-verification-server/pkg/database"
	"github.com/google/exposure-notifications-verification-server/pkg/rbac"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestHandleShow(t *testing.T) {
	t.Parallel()

	ctx := project.TestContext(t)
	harness := envstest.NewServerConfig(t, testDatabaseInstance)

	c := mobileapps.New(harness.Database, harness.Renderer)
	handler := harness.WithCommonMiddlewares(c.HandleShow())

	t.Run("middleware", func(t *testing.T) {
		t.Parallel()

		envstest.ExerciseSessionMissing(t, handler)
		envstest.ExerciseMembershipMissing(t, handler)
		envstest.ExercisePermissionMissing(t, handler)
		envstest.ExerciseIDNotFound(t, &database.Membership{
			Realm:       &database.Realm{},
			User:        &database.User{},
			Permissions: rbac.MobileAppRead,
		}, handler)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()

		c := mobileapps.New(harness.BadDatabase, harness.Renderer)
		handler := c.HandleShow()

		ctx := ctx
		ctx = controller.WithSession(ctx, &sessions.Session{})
		ctx = controller.WithMembership(ctx, &database.Membership{
			Realm:       &database.Realm{},
			User:        &database.User{},
			Permissions: rbac.MobileAppRead,
		})

		w, r := envstest.BuildFormRequest(ctx, t, http.MethodGet, "/", nil)
		r = mux.SetURLVars(r, map[string]string{"id": "1"})
		handler.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("expected %d to be %d: %s", got, want, w.Body.String())
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		realm, err := harness.Database.FindRealm(1)
		if err != nil {
			t.Fatal(err)
		}

		app := &database.MobileApp{
			RealmID: realm.ID,
			Name:    "Appy",
			AppID:   "com.example.app",
			URL:     "https://app.example.com",
			OS:      database.OSTypeIOS,
		}
		if err := harness.Database.SaveMobileApp(app, database.SystemTest); err != nil {
			t.Fatal(err)
		}

		ctx := ctx
		ctx = controller.WithSession(ctx, &sessions.Session{})
		ctx = controller.WithMembership(ctx, &database.Membership{
			Realm:       realm,
			User:        &database.User{},
			Permissions: rbac.MobileAppRead,
		})

		w, r := envstest.BuildFormRequest(ctx, t, http.MethodGet, "/", nil)
		r = mux.SetURLVars(r, map[string]string{"id": fmt.Sprintf("%d", app.ID)})
		handler.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Errorf("expected %d to be %d: %s", got, want, w.Body.String())
		}
	})
}
