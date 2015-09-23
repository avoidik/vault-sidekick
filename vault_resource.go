/*
Copyright 2015 Home Office All rights reserved.

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
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	// OptionFilename ... option to set the filename of the resource
	OptionFilename = "fn"
	// OptionFormat ... option to set the output format (yaml, xml, json)
	OptionFormat = "fmt"
	// OptionCommonName ... use by the PKI resource
	OptionCommonName = "cn"
	// OptionTemplatePath ... the full path to a template
	OptionTemplatePath = "tpl"
	// OptionRenewal ... a duration to renew the resource
	OptionRenewal = "rn"
	// OptionRevoke ... revoke an old lease when retrieving a new one
	OptionRevoke = "rv"
	// OptionUpdate ... override the lease of the resource
	OptionUpdate = "up"
)

var (
	resourceFormatRegex = regexp.MustCompile("^(yaml|json|ini|txt|cert|csv)$")

	// a map of valid resource to retrieve from vault
	validResources = map[string]bool{
		"pki":       true,
		"aws":       true,
		"secret":    true,
		"mysql":     true,
		"tpl":       true,
		"postgres":  true,
		"cassandra": true,
	}
)

func defaultVaultResource() *VaultResource {
	return &VaultResource{
		format:    "yaml",
		renewable: false,
		revoked:   false,
		options:   make(map[string]string, 0),
	}
}

// VaultResource ... the structure which defined a resource set from vault
type VaultResource struct {
	// the namespace of the resource
	resource string
	// the name of the resource
	name string
	// the format of the resource
	format string
	// whether the resource should be renewed?
	renewable bool
	// whether the resource should be revoked?
	revoked bool
	// the lease duration
	update time.Duration
	// additional options to the resource
	options map[string]string
}

// GetFilename ... generates a resource filename by default the resource name and resource type, which
// can override by the OPTION_FILENAME option
func (r VaultResource) GetFilename() string {
	if path, found := r.options[OptionFilename]; found {
		return path
	}

	return fmt.Sprintf("%s.%s", r.name, r.resource)
}

// IsValid ... checks to see if the resource is valid
func (r *VaultResource) IsValid() error {
	// step: check the resource type
	if _, found := validResources[r.resource]; !found {
		return fmt.Errorf("unsupported resource type: %s", r.resource)
	}

	// step: check the options
	if err := r.isValidOptions(); err != nil {
		return fmt.Errorf("invalid resource options, %s", err)
	}

	// step: check is have all the required options to this resource type
	if err := r.isValidResource(); err != nil {
		return fmt.Errorf("invalid resource: %s, %s", r, err)
	}

	return nil
}

// isValidResource ... validate the resource meets the requirements
func (r *VaultResource) isValidResource() error {
	switch r.resource {
	case "pki":
		if _, found := r.options[OptionCommonName]; !found {
			return fmt.Errorf("pki resource requires a common name specified")
		}
	case "tpl":
		if _, found := r.options[OptionTemplatePath]; !found {
			return fmt.Errorf("template resource requires a template path option")
		}
	}

	return nil
}

// isValidOptions ... iterates through the options, converts the options and so forth
func (r *VaultResource) isValidOptions() error {
	// check the filename directive
	for opt, val := range r.options {
		switch opt {
		case OptionFormat:
			if matched := resourceFormatRegex.MatchString(r.options[OptionFormat]); !matched {
				return fmt.Errorf("unsupported output format: %s", r.options[OptionFormat])
			}
			r.format = val
		case OptionUpdate:
			duration, err := time.ParseDuration(val)
			if err != nil {
				return fmt.Errorf("the update option: %s is not value, should be a duration format", val)
			}
			r.update = duration
		case OptionRevoke:
			choice, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("the revoke option: %s is invalid, should be a boolean", val)
			}
			r.revoked = choice
		case OptionRenewal:
			choice, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("the renewal option: %s is invalid, should be a boolean", val)
			}
			r.renewable = choice
		case OptionFilename:
			// @TODO need to check it's valid filename / path
		case OptionCommonName:
			// @TODO need to check it's a valid hostname
		case OptionTemplatePath:
			if exists, _ := fileExists(val); !exists {
				return fmt.Errorf("the template file: %s does not exist", val)
			}
		}
	}

	return nil
}

// String ... a string representation of the struct
func (r VaultResource) String() string {
	return fmt.Sprintf("%s/%s", r.resource, r.name)
}
