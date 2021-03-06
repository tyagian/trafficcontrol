/*
Name
	to_requester - ORT Traffic Ops Requestor

Synopsis
	to_requester [-hI] [-D value] [-d value] [-e value] [-H value] [-i value] \
		[-l value] [-P value] [-t value] [-u value] [-U value]

Description
  The to_requester is used get update status, package information, linux
  chkconfig status, system info and status from Traffic Ops, see the
  --get-data option.  If no --get-data option is specified, the servers
  system-info is fetched and returned.

Options
	-D, --get-data=value
      	non-config-file Traffic Ops Data to get. Valid values are
        all, update-status, packages, chkconfig, system-info, and
        statuses [all]
	-d, --log-location-debug=value
        Where to log debugs. May be a file path, stdout or stderr.
        Default is no debug logging.
	-e, --log-location-error=value
        Where to log errors. May be a file path, stdout, or stderr.
        Default is stderr.
	-i, --log-location-info=value
        Where to log infos. May be a file path, stdout or stderr.
        Default is stderr.
	-H, --cache-host-name=value
     		Host name of the cache to generate config for. Must be the
        server host name in Traffic Ops, not a URL, and not the FQDN.
        Defaults to the OS configured hostname.
	-h, --help  Print usage information and exit
 	-I, --traffic-ops-insecure
				[true | false] ignore certificate errors from Traffic Ops
	-l, --login-dispersion=value
        [seconds] wait a random number of seconds between 0 and
        [seconds] before login to traffic ops, default 0
	-P, --traffic-ops-password=value
        Traffic Ops password. Required. May also be set with the
        environment variable TO_PASS
	-t, --traffic-ops-timeout-milliseconds=value
        Timeout in milli-seconds for Traffic Ops requests, default
        is 30000 [30000]
	-u, --traffic-ops-url=value
        Traffic Ops URL. Must be the full URL, including the scheme.
        Required. May also be set with     the environment variable
        TO_URL
	-U, --traffic-ops-user=value
        Traffic Ops username. Required. May also be set with the
        environment variable TO_USER
*/

package main

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
	"errors"
	"fmt"
	"os"

	"github.com/apache/trafficcontrol/lib/go-log"
	"github.com/apache/trafficcontrol/traffic_ops_ort/atstccfg/toreq"
	"github.com/apache/trafficcontrol/traffic_ops_ort/atstccfg/toreqnew"
	"github.com/apache/trafficcontrol/traffic_ops_ort/to_requester/config"
	"github.com/apache/trafficcontrol/traffic_ops_ort/to_requester/getdata"
)

var (
	cfg config.Cfg
)

func main() {
	var err error

	cfg, err = config.InitConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	} else {
		log.Infoln("configuration initialized")
		cfg.PrintConfig()
	}

	// login to traffic ops.
	tccfg, err := toConnect()
	if err != nil {
		log.Errorf("%s\n", err)
		os.Exit(2)
	}
	if tccfg.Cfg.GetData != "" {
		if err := getdata.WriteData(*tccfg); err != nil {
			log.Errorf("writing data: %s\n", err.Error())
			os.Exit(3)
		}
	}
}

/*
 * connect and login to traffic ops
 */
func toConnect() (*config.TCCfg, error) {
	_toClient, err := toreq.New(cfg.TOURL, cfg.TOUser, cfg.TOPass, cfg.TOInsecure, cfg.TOTimeoutMS, config.UserAgent)
	if err != nil {
		return nil, errors.New("failed to connect to traffic ops: " + err.Error())
	}
	_toClientNew, err := toreqnew.New(_toClient.Cookies(cfg.TOURL), cfg.TOURL, cfg.TOUser, cfg.TOPass, cfg.TOInsecure, cfg.TOTimeoutMS, config.UserAgent)

	tccfg := config.TCCfg{
		Cfg:         cfg,
		TOClient:    _toClient,
		TOClientNew: _toClientNew,
	}

	return &tccfg, nil
}
