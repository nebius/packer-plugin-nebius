// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package image

// packersdk.Artifact implementation
type Artifact struct {
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
	imageID   string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return a.imageID
}

func (a *Artifact) String() string {
	return a.Id()
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return nil
}
