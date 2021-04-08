// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/config"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/operator"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/producer"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/server"
	// +kubebuilder:scaffold:imports
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal("failed to parse environment configuration")
	}
	switch cfg.Function {
	case "operator":
		operator.Start(cfg)
	case "server":
		server.Start()
	case "producer":
		producer.Start(cfg)
	default:
		log.Fatalf("FUNCTION \"%s\" not recognised", cfg.Function)
	}
}
