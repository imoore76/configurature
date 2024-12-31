// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
This file contains the Value interface implementation for the ConfigFile type
which is used to specify a configuration file field on a configurature struct
*/
package configurature

// Type representing a config file setting
type ConfigFile string

func (f *ConfigFile) String() string {
	return (string)(*f)
}

func (f *ConfigFile) Set(v string) error {
	*f = (ConfigFile)(v)
	return nil
}

func (f *ConfigFile) Type() string {
	return "configFile"
}
