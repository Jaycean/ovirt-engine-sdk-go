//
// Copyright (c) 2020 huihui <huihui.fu@cs2c.com.cn>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package examples

import (
	"fmt"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func changeHostCluster() {
	inputRawURL := "https://10.1.111.229/ovirt-engine/api"

	conn, err := ovirtsdk4.NewConnectionBuilder().
		URL(inputRawURL).
		Username("admin@internal").
		Password("qwer1234").
		Insecure(true).
		Compress(true).
		Timeout(time.Second * 10).
		Build()
	if err != nil {
		fmt.Printf("Make connection failed, reason: %v\n", err)
		return
	}
	defer conn.Close()

	// To use `Must` methods, you should recover it if panics
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panics occurs, try the non-Must methods to find the reason")
		}
	}()

	// Get the reference to the service that manages the hosts:
	hostsService := conn.SystemService().HostsService()
	listResp, err := hostsService.List().Search("name=myhost").Send()
	if err != nil {
		fmt.Printf("Failed to search host list, reason: %v\n", err)
		return
	}

	// Find the host:
	hostSlice, _ := listResp.Hosts()
	host := hostSlice.Slice()[0]

	// Get the reference to the service that manages the host
	hostService := hostsService.HostService(host.MustId())

	// Put host into maintenance:
	if host.MustStatus() == ovirtsdk4.HOSTSTATUS_MAINTENANCE {
		hostService.Deactivate().MustSend()

		// Wait till the host is in maintenance:
		for {
			time.Sleep(5 * time.Second)
			getHostResp, err := hostService.Get().Send()
			if err != nil {
				continue
			}
			if getHost, ok := getHostResp.Host(); ok {
				if getHost.MustStatus() == ovirtsdk4.HOSTSTATUS_MAINTENANCE {
					break
				}
			}
		}
	}

	// Change the host cluster:
	newCluster := &ovirtsdk4.Cluster{}
	newCluster.SetName("mycluster")
	host.SetCluster(newCluster)
	hostService.Update().Host(host).Send()

	//# Activate the host again:
	hostService.Activate().MustSend()

	// Wait till the host is in maintenance:
	for {
		time.Sleep(5 * time.Second)
		getHostResp, err := hostService.Get().Send()
		if err != nil {
			continue
		}
		if getHost, ok := getHostResp.Host(); ok {
			if getHost.MustStatus() == ovirtsdk4.HOSTSTATUS_UP {
				break
			}
		}
	}

}
