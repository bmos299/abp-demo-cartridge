// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---

// +kubebuilder:object:generate=true
package commoncrd

type License struct {
	Accept Accept `json:"accept"`
}

// +kubebuilder:validation:Enum=true
type Accept bool
