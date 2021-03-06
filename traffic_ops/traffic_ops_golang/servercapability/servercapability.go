package servercapability

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/apache/trafficcontrol/lib/go-log"
	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/lib/go-tc/tovalidate"
	"github.com/apache/trafficcontrol/lib/go-util"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/dbhelpers"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

type TOServerCapability struct {
	api.APIInfoImpl `json:"-"`
	RequestedName   string
	tc.ServerCapability
}

func (v *TOServerCapability) SetLastUpdated(t tc.TimeNoMod) { v.LastUpdated = &t }
func (v *TOServerCapability) NewReadObj() interface{}       { return &tc.ServerCapability{} }
func (v *TOServerCapability) InsertQuery() string {
	return `
INSERT INTO server_capability (
  name
)
VALUES (
  :name
)
RETURNING last_updated
`
}

func (v *TOServerCapability) SelectQuery() string {
	return `
SELECT
  name,
  last_updated
FROM
  server_capability sc
`
}

func (v *TOServerCapability) updateQuery() string {
	return `
UPDATE server_capability sc SET
	name = $1
WHERE sc.name = $2
RETURNING sc.name, sc.last_updated
`
}

func (v *TOServerCapability) DeleteQuery() string {
	return `
DELETE FROM server_capability WHERE name=:name
`
}

func (v *TOServerCapability) ParamColumns() map[string]dbhelpers.WhereColumnInfo {
	return map[string]dbhelpers.WhereColumnInfo{
		"name": {"sc.name", nil},
	}
}

func (v TOServerCapability) GetKeyFieldsInfo() []api.KeyFieldInfo {
	return []api.KeyFieldInfo{{"name", api.GetStringKey}}
}

// Implementation of the Identifier, Validator interface functions
func (v TOServerCapability) GetKeys() (map[string]interface{}, bool) {
	return map[string]interface{}{"name": v.Name}, true
}

func (v *TOServerCapability) SetKeys(keys map[string]interface{}) {
	v.RequestedName = v.Name
	v.Name, _ = keys["name"].(string)
}

func (v *TOServerCapability) GetAuditName() string {
	return v.Name
}

func (v *TOServerCapability) GetType() string {
	return "server capability"
}

func (v *TOServerCapability) Validate() error {
	rule := validation.NewStringRule(tovalidate.IsAlphanumericUnderscoreDash, "must consist of only alphanumeric, dash, or underscore characters")
	errs := validation.Errors{
		"name": validation.Validate(v.Name, validation.Required, rule),
	}
	return util.JoinErrs(tovalidate.ToErrors(errs))
}

func (v *TOServerCapability) Read(h http.Header, useIMS bool) ([]interface{}, error, error, int, *time.Time) {
	api.DefaultSort(v.APIInfo(), "name")
	return api.GenericRead(h, v, useIMS)
}

func (v *TOServerCapability) Update(h http.Header) (error, error, int) {
	sc, userErr, sysErr, errCode, _ := v.Read(h, false)
	if userErr != nil || sysErr != nil {
		return userErr, sysErr, errCode
	}
	if len(sc) != 1 {
		return fmt.Errorf("cannot find exactly one server capability with the query string provided"), nil, http.StatusBadRequest
	}

	// check if the name was being updated by someone else
	var existingLastUpdated *tc.TimeNoMod
	q := `SELECT last_updated FROM server_capability WHERE name = $1`
	if err := v.ReqInfo.Tx.QueryRow(q, v.Name).Scan(&existingLastUpdated); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("server capability was not found"), nil, http.StatusNotFound
		}
		return nil, errors.New("server capability update: querying: " + err.Error()), http.StatusInternalServerError
	}
	if !api.IsUnmodified(h, existingLastUpdated.Time) {
		return errors.New("the resource has been modified since the time specified by the request headers"), nil, http.StatusPreconditionFailed
	}

	// udpate server capability name
	rows, err := v.ReqInfo.Tx.Query(v.updateQuery(), v.RequestedName, v.Name)
	if err != nil {
		return nil, fmt.Errorf("server capability update: error setting the name for server capability %v: %v", v.Name, err.Error()), http.StatusInternalServerError
	}
	defer log.Close(rows, "unable to close DB connection")

	for rows.Next() {
		err = rows.Scan(&v.Name, &v.LastUpdated)
		if err != nil {
			return api.ParseDBError(err)
		}
	}
	return nil, nil, http.StatusOK
}

func (v *TOServerCapability) SelectMaxLastUpdatedQuery(where, orderBy, pagination, tableName string) string {
	return `SELECT max(t) from (
		SELECT max(sc.last_updated) as t from server_capability sc ` + where + orderBy + pagination +
		` UNION ALL
	select max(l.last_updated) as t from last_deleted l where l.table_name='server_capability') as res`
}

func (v *TOServerCapability) Create() (error, error, int) { return api.GenericCreateNameBasedID(v) }
func (v *TOServerCapability) Delete() (error, error, int) { return api.GenericDelete(v) }
