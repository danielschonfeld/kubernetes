/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package v1

import (
	"fmt"
	"reflect"

	newer "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/conversion"
)

func init() {
	err := newer.Scheme.AddConversionFuncs(
		convert_v1_Container_To_api_Container,
		convert_api_Container_To_v1_Container,
	)
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}

	// Add field conversion funcs.
	err = newer.Scheme.AddFieldLabelConversionFunc("v1", "Pod",
		func(label, value string) (string, string, error) {
			switch label {
			case "metadata.name",
				"metadata.namespace",
				"status.phase",
				"spec.host":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		})
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
	err = newer.Scheme.AddFieldLabelConversionFunc("v1", "Node",
		func(label, value string) (string, string, error) {
			switch label {
			case "metadata.name":
				return label, value, nil
			case "spec.unschedulable":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		})
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
	err = newer.Scheme.AddFieldLabelConversionFunc("v1", "ReplicationController",
		func(label, value string) (string, string, error) {
			switch label {
			case "metadata.name",
				"status.replicas":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		})
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
	err = newer.Scheme.AddFieldLabelConversionFunc("v1", "Event",
		func(label, value string) (string, string, error) {
			switch label {
			case "involvedObject.kind",
				"involvedObject.namespace",
				"involvedObject.name",
				"involvedObject.uid",
				"involvedObject.apiVersion",
				"involvedObject.resourceVersion",
				"involvedObject.fieldPath",
				"reason",
				"source":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		})
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
	err = newer.Scheme.AddFieldLabelConversionFunc("v1", "Namespace",
		func(label, value string) (string, string, error) {
			switch label {
			case "status.phase":
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		})
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}

func convert_v1_Container_To_api_Container(in *Container, out *newer.Container, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*Container))(in)
	}
	out.Name = in.Name
	out.Image = in.Image
	if in.Command != nil {
		out.Command = make([]string, len(in.Command))
		for i := range in.Command {
			out.Command[i] = in.Command[i]
		}
	}
	if in.Args != nil {
		out.Args = make([]string, len(in.Args))
		for i := range in.Args {
			out.Args[i] = in.Args[i]
		}
	}
	out.WorkingDir = in.WorkingDir
	if in.Ports != nil {
		out.Ports = make([]newer.ContainerPort, len(in.Ports))
		for i := range in.Ports {
			if err := convert_v1_ContainerPort_To_api_ContainerPort(&in.Ports[i], &out.Ports[i], s); err != nil {
				return err
			}
		}
	}
	if in.Env != nil {
		out.Env = make([]newer.EnvVar, len(in.Env))
		for i := range in.Env {
			if err := convert_v1_EnvVar_To_api_EnvVar(&in.Env[i], &out.Env[i], s); err != nil {
				return err
			}
		}
	}
	if err := s.Convert(&in.Resources, &out.Resources, 0); err != nil {
		return err
	}
	if in.VolumeMounts != nil {
		out.VolumeMounts = make([]newer.VolumeMount, len(in.VolumeMounts))
		for i := range in.VolumeMounts {
			if err := convert_v1_VolumeMount_To_api_VolumeMount(&in.VolumeMounts[i], &out.VolumeMounts[i], s); err != nil {
				return err
			}
		}
	}
	if in.LivenessProbe != nil {
		out.LivenessProbe = new(newer.Probe)
		if err := convert_v1_Probe_To_api_Probe(in.LivenessProbe, out.LivenessProbe, s); err != nil {
			return err
		}
	} else {
		out.LivenessProbe = nil
	}
	if in.ReadinessProbe != nil {
		out.ReadinessProbe = new(newer.Probe)
		if err := convert_v1_Probe_To_api_Probe(in.ReadinessProbe, out.ReadinessProbe, s); err != nil {
			return err
		}
	} else {
		out.ReadinessProbe = nil
	}
	if in.Lifecycle != nil {
		out.Lifecycle = new(newer.Lifecycle)
		if err := convert_v1_Lifecycle_To_api_Lifecycle(in.Lifecycle, out.Lifecycle, s); err != nil {
			return err
		}
	} else {
		out.Lifecycle = nil
	}
	out.TerminationMessagePath = in.TerminationMessagePath
	out.ImagePullPolicy = newer.PullPolicy(in.ImagePullPolicy)
	if in.SecurityContext != nil {
		if in.SecurityContext.Capabilities != nil {
			if !reflect.DeepEqual(in.SecurityContext.Capabilities.Add, in.Capabilities.Add) ||
				!reflect.DeepEqual(in.SecurityContext.Capabilities.Drop, in.Capabilities.Drop) {
				return fmt.Errorf("container capability settings do not match security context settings, cannot convert")
			}
		}
		if in.SecurityContext.Privileged != nil {
			if in.Privileged != *in.SecurityContext.Privileged {
				return fmt.Errorf("container privileged settings do not match security context settings, cannot convert")
			}
		}
	}
	if in.SecurityContext != nil {
		out.SecurityContext = new(newer.SecurityContext)
		if err := convert_v1_SecurityContext_To_api_SecurityContext(in.SecurityContext, out.SecurityContext, s); err != nil {
			return err
		}
	} else {
		out.SecurityContext = nil
	}
	return nil
}

func convert_api_Container_To_v1_Container(in *newer.Container, out *Container, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*newer.Container))(in)
	}
	out.Name = in.Name
	out.Image = in.Image
	if in.Command != nil {
		out.Command = make([]string, len(in.Command))
		for i := range in.Command {
			out.Command[i] = in.Command[i]
		}
	}
	if in.Args != nil {
		out.Args = make([]string, len(in.Args))
		for i := range in.Args {
			out.Args[i] = in.Args[i]
		}
	}
	out.WorkingDir = in.WorkingDir
	if in.Ports != nil {
		out.Ports = make([]ContainerPort, len(in.Ports))
		for i := range in.Ports {
			if err := convert_api_ContainerPort_To_v1_ContainerPort(&in.Ports[i], &out.Ports[i], s); err != nil {
				return err
			}
		}
	}
	if in.Env != nil {
		out.Env = make([]EnvVar, len(in.Env))
		for i := range in.Env {
			if err := convert_api_EnvVar_To_v1_EnvVar(&in.Env[i], &out.Env[i], s); err != nil {
				return err
			}
		}
	}
	if err := s.Convert(&in.Resources, &out.Resources, 0); err != nil {
		return err
	}
	if in.VolumeMounts != nil {
		out.VolumeMounts = make([]VolumeMount, len(in.VolumeMounts))
		for i := range in.VolumeMounts {
			if err := convert_api_VolumeMount_To_v1_VolumeMount(&in.VolumeMounts[i], &out.VolumeMounts[i], s); err != nil {
				return err
			}
		}
	}
	if in.LivenessProbe != nil {
		out.LivenessProbe = new(Probe)
		if err := convert_api_Probe_To_v1_Probe(in.LivenessProbe, out.LivenessProbe, s); err != nil {
			return err
		}
	} else {
		out.LivenessProbe = nil
	}
	if in.ReadinessProbe != nil {
		out.ReadinessProbe = new(Probe)
		if err := convert_api_Probe_To_v1_Probe(in.ReadinessProbe, out.ReadinessProbe, s); err != nil {
			return err
		}
	} else {
		out.ReadinessProbe = nil
	}
	if in.Lifecycle != nil {
		out.Lifecycle = new(Lifecycle)
		if err := convert_api_Lifecycle_To_v1_Lifecycle(in.Lifecycle, out.Lifecycle, s); err != nil {
			return err
		}
	} else {
		out.Lifecycle = nil
	}
	out.TerminationMessagePath = in.TerminationMessagePath
	out.ImagePullPolicy = PullPolicy(in.ImagePullPolicy)
	if in.SecurityContext != nil {
		out.SecurityContext = new(SecurityContext)
		if err := convert_api_SecurityContext_To_v1_SecurityContext(in.SecurityContext, out.SecurityContext, s); err != nil {
			return err
		}
	} else {
		out.SecurityContext = nil
	}
	// now that we've converted set the container field from security context
	if out.SecurityContext != nil && out.SecurityContext.Privileged != nil {
		out.Privileged = *out.SecurityContext.Privileged
	}
	// now that we've converted set the container field from security context
	if out.SecurityContext != nil && out.SecurityContext.Capabilities != nil {
		out.Capabilities = *out.SecurityContext.Capabilities
	}
	return nil
}
