//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

// Code generated by deepcopy-gen. DO NOT EDIT.

package api

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthConfig) DeepCopyInto(out *AuthConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthConfig.
func (in *AuthConfig) DeepCopy() *AuthConfig {
	if in == nil {
		return nil
	}
	out := new(AuthConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	if in.Contexts != nil {
		in, out := &in.Contexts, &out.Contexts
		*out = make(map[string]*ContextConfig, len(*in))
		for key, val := range *in {
			var outVal *ContextConfig
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(ContextConfig)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = make(map[string]*AuthConfig, len(*in))
		for key, val := range *in {
			var outVal *AuthConfig
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(AuthConfig)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContextConfig) DeepCopyInto(out *ContextConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContextConfig.
func (in *ContextConfig) DeepCopy() *ContextConfig {
	if in == nil {
		return nil
	}
	out := new(ContextConfig)
	in.DeepCopyInto(out)
	return out
}
